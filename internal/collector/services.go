package collector

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/AshitomW/Drifty/internal/models"
)






func (c *Collector) collectServices(ctx context.Context) (map[string]models.ServiceInfo,error){

	switch c.config.Services.InitType{
	case "sysemd":
		return c.collectSystemdServices(ctx)
	case "sysvinit":
		return c.collectSysvinitServices(ctx)
	case "launchd":
		return c.collectLaunchdServices(ctx)
	default:
		if runtime.GOOS == "darwin"{
			return c.collectLaunchdServices(ctx)
		}

		return c.collectSystemdServices(ctx)
	
	}

}


func (c *Collector) collectSystemdServices(ctx context.Context)(map[string]models.ServiceInfo,error){
	services := make(map[string]models.ServiceInfo)


	// retireve list of all services
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--all", "--no-legend", "--no-pager")


	output, err := cmd.Output()
	if err != nil {
		return services, err 
	}


	// Compiling the include / excludwe patterns


	var includePatterns, excludePatterns []*regexp.Regexp


	for _, pattern := range c.config.Services.Include {
		if re, err := regexp.Compile(pattern); err == nil{
			includePatterns = append(includePatterns, re)
		}
	}


	for _ , pattern := range c.config.Services.Exclude{
		if re, err := regexp.Compile(pattern); err == nil{
			excludePatterns = append(excludePatterns, re)
		}
	}



	scanner := bufio.NewScanner(bytes.NewReader(output))


	for scanner.Scan(){
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 4{
			continue 
		}


		name := strings.TrimSuffix(fields[0],".service")

		// Checking for filters


		if !matchesPatterns(name,includePatterns,excludePatterns){
			continue 
		}

		activeState := fields[2]
		subState := fields[3]

		// Determining the status

		status := "unknown"
		switch activeState{
		case "active":
			status = "running"
		case "inactive":
			status ="stopped"
		case "failed":
			status = "failed"
		}


		enabled := c.isServiceEnabled(ctx,name)


		services[name] = models.ServiceInfo{
			Name: name,
			Status: status,
			Enabled: enabled,
			ActiveState: activeState,
			SubState: subState,
			Exists: true,
		}


	}
		return services, nil
}


func matchesPatterns(name string, include, exclude[] *regexp.Regexp) bool{
	if len(include) > 0 {
		matched := false 
		for _ , re := range include{
			if re.MatchString(name){
				matched = true
				break
			}
		}

		if !matched{
			return false;
		}
	}


	for _, re := range exclude{
		if re.MatchString(name){
			return false 
		}
	}
	return true
}




func (c *Collector) isServiceEnabled(ctx context.Context, name string) bool{
	cmd := exec.CommandContext(ctx,"systemctl","is-enabled",name+".service")
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output)) == "enabled"
}



func (c *Collector) collectSysvinitServices(ctx context.Context) (map[string]models.ServiceInfo,error){


	services := make(map[string]models.ServiceInfo)


	cmd := exec.CommandContext(ctx,"service","--status-all")
	output, _ := cmd.CombinedOutput()


	re := regexp.MustCompile(`\[\s*([+-?])\s*\]\s+(\S+)`)
	matches := re.FindAllStringSubmatch(string(output),-1)


	for _, match := range matches{
		if len(match) >= 3{
			name := match[2]
			status:= "unknown"
			switch match[1]{
			case "+": 
				status = "running"
			case "-":
				status = "stopped"
			}

			services[name] = models.ServiceInfo{
				Name: name,
				Status: status,
				Exists: true,
			}

		}
	}

	return services,nil
}




func (c *Collector) collectLaunchdServices(ctx context.Context)(map[string]models.ServiceInfo,error){
	services := make(map[string]models.ServiceInfo)

	cmd := exec.CommandContext(ctx,"launctl","list")

	output, err := cmd.Output()

	if err != nil{
		return services, err 
	}


	scanner := bufio.NewScanner(bytes.NewReader(output))

	scanner.Scan() // skipping the header


	for scanner.Scan(){
		fields := strings.Fields(scanner.Text())


		if len(fields) >= 3{
			name := fields[2]
			pid := fields[0]

			status := "stopped"

			if pid != "-"{
				status = "running"
			}

			services[name] = models.ServiceInfo{
				Name:name,
				Status: status,
				Exists: true,
			}
		}
	}
	return services,nil
}
