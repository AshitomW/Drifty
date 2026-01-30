package collector

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/AshitomW/Drifty/internal/models"
)

type packageCollector func(ctx context.Context) (map[string]models.PackageInfo, error)

func (c *Collector) collectPackages(ctx context.Context) (map[string]models.PackageInfo, error) {

	packages := make(map[string]models.PackageInfo)

	var mu sync.Mutex
	var wg sync.WaitGroup

	collectors := make(map[string]packageCollector)

	// Registering collectors based on the configuration

	for _, manager := range c.config.Packages.Managers {
		switch manager {
		case "apt", "dpkg":
			collectors["dpkg"] = c.collectDpkgPackages
		case "yum", "rpm":
			collectors["rpm"] = c.collectRpmPackages
		case "apk":
			collectors["apk"] = c.collectApkPackages
		case "pip", "pip3":
			collectors["pip"] = c.collectPipPackages
		case "npm":
			collectors["npm"] = c.collectNpmPackages
		case "go":
			collectors["go"] = c.collectGoModules
		case "brew":
			collectors["brew"] = c.collectBrewPackages
		}

	}

	for name, collector := range collectors {
		wg.Add(1)
		go func(name string, collect packageCollector) {
			defer wg.Done()
			pkgs, err := collect(ctx)
			if err != nil {
				return // skipping if error
			}
			mu.Lock()
			for k, v := range pkgs {
				packages[name+":"+k] = v
			}
			mu.Unlock()
		}(name, collector)
	}

	wg.Wait()
	return packages, nil

}

func (c *Collector) collectDpkgPackages(ctx context.Context) (map[string]models.PackageInfo, error) {
	packages := make(map[string]models.PackageInfo)

	if runtime.GOOS != "linux" {
		return packages, nil
	}

	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Package}\t${Version}\t${Architecture}\n")

	output, err := cmd.Output()

	if err != nil {
		return packages, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) >= 2 {
			packages[parts[0]] = models.PackageInfo{
				Name:         parts[0],
				Version:      parts[1],
				Architecture: safeIndex(parts, 2),
				Manager:      "dpkg",
				Exists:       true,
			}
		}
	}

	return packages, nil
}

func (c *Collector) collectRpmPackages(ctx context.Context) (map[string]models.PackageInfo, error) {

	packages := make(map[string]models.PackageInfo)

	if runtime.GOOS != "linux" {
		return packages, nil
	}

	cmd := exec.CommandContext(ctx, "rpm", "-qa", "--queryformat", "%{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\n")

	output, err := cmd.Output()

	if err != nil {
		return packages, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) >= 2 {
			packages[parts[0]] = models.PackageInfo{
				Name:         parts[0],
				Version:      parts[1],
				Architecture: safeIndex(parts, 2),
				Manager:      "rpm",
				Exists:       true,
			}
		}
	}
	return packages, nil
}

func (c *Collector) collectApkPackages(ctx context.Context) (map[string]models.PackageInfo, error) {

	packages := make(map[string]models.PackageInfo)

	cmd := exec.CommandContext(ctx, "apk", "list", "--installed")

	output, err := cmd.Output()

	if err != nil {
		return packages, err
	}

	// Format: package-name-version arch {origin} (license) [status]
	re := regexp.MustCompile(`^(.+)-(\d+[^\s]*)\s+(\w+)`)

	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) >= 4 {
			packages[matches[1]] = models.PackageInfo{
				Name:         matches[1],
				Version:      matches[2],
				Architecture: matches[3],
				Manager:      "apk",
				Exists:       true,
			}
		}
	}

	return packages, nil
}

func (c *Collector) collectPipPackages(ctx context.Context) (map[string]models.PackageInfo, error) {

	packages := make(map[string]models.PackageInfo)

	// Will try pip3 first, then pip

	pipCmd := "pip3"
	if _, err := exec.LookPath("pip3"); err != nil {
		pipCmd = "pip"
	}

	cmd := exec.CommandContext(ctx, pipCmd, "list", "--format=freeze")

	output, err := cmd.Output()
	if err != nil {
		return packages, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "=="); idx > 0 {
			name := line[:idx]
			version := line[idx+2:]
			packages[name] = models.PackageInfo{
				Name:    name,
				Version: version,
				Manager: "pip",
				Exists:  true,
			}
		}
	}

	return packages, nil
}

func (c *Collector) collectNpmPackages(ctx context.Context) (map[string]models.PackageInfo, error) {

	packages := make(map[string]models.PackageInfo)

	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0", "--json")

	output, err := cmd.Output()

	if err != nil {
		// npm will return error if some pacakges have issue
		if len(output) == 0 {
			return packages, nil
		}
	}

	// Lightweight parsing using regex, assuming a predictable JSON shape.
	// For full JSON support, a proper decoder would be required.
	re := regexp.MustCompile(`"([^"]+)":\s*{\s*"version":\s*"([^"]+)"`)

	matches := re.FindAllStringSubmatch(string(output), -1)

	for _, match := range matches {
		if len(match) >= 3 {
			packages[match[1]] = models.PackageInfo{
				Name:    match[1],
				Version: match[2],
				Manager: "npm",
				Exists:  true,
			}
		}
	}

	return packages, nil
}

func (c *Collector) collectGoModules(ctx context.Context) (map[string]models.PackageInfo, error) {

	packages := make(map[string]models.PackageInfo)

	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-json", "all")

	output, err := cmd.Output()

	if err != nil {
		return packages, nil
	}

	re := regexp.MustCompile(`"Path":\s*"([^"]+)"[^}]*"Version":\s*"([^"]+)"`)

	matches := re.FindAllStringSubmatch(string(output), -1)

	for _, match := range matches {
		if len(match) >= 3 {
			packages[match[1]] = models.PackageInfo{
				Name:    match[1],
				Version: match[2],
				Manager: "go",
				Exists:  true,
			}
		}
	}

	return packages, nil
}

func (c *Collector) collectBrewPackages(ctx context.Context) (map[string]models.PackageInfo, error) {

	packages := make(map[string]models.PackageInfo)

	if runtime.GOOS != "darwin" {
		return packages, nil
	}

	cmd := exec.CommandContext(ctx, "brew", "list", "--versions")
	output, err := cmd.Output()

	if err != nil {
		return packages, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			packages[parts[0]] = models.PackageInfo{
				Name:    parts[0],
				Version: parts[1],
				Manager: "brew",
				Exists:  true,
			}
		}
	}
	return packages, nil
}

func safeIndex(slice []string, index int) string {
	if index < len(slice) {
		return slice[index]
	}
	return ""
}
