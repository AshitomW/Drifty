package collector

import (
	"context"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AshitomW/Drifty/internal/models"
)

func (c *Collector) collectNetworkConfig(ctx context.Context) (models.NetworkConfig, error) {
	config := models.NetworkConfig{
		Interfaces: make(map[string]models.NetworkInterface),
		Routes:     []models.Route{},
		DNS:        models.DNSConfig{},
	}

	if c.config.Network.Interfaces {
		interfaces, err := c.collectNetworkInterfaces(ctx)
		if err == nil {
			config.Interfaces = interfaces
		}
	}

	if c.config.Network.Routes {
		routes, err := c.collectRoutes(ctx)
		if err == nil {
			config.Routes = routes
		}
	}

	if c.config.Network.DNS {
		dns, err := c.collectDNS(ctx)
		if err == nil {
			config.DNS = dns
		}
	}

	if c.config.Network.FirewallRules {
		rules, err := c.collectFirewallRules(ctx)
		if err == nil {
			config.FirewallRules = rules
		}
	}

	return config, nil
}

func (c *Collector) collectNetworkInterfaces(ctx context.Context) (map[string]models.NetworkInterface, error) {
	interfaces := make(map[string]models.NetworkInterface)

	ifaces, err := net.Interfaces()
	if err != nil {
		return interfaces, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ipAddresses []string
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() {
				ipAddresses = append(ipAddresses, ip.String())
			}
		}

		interfaces[iface.Name] = models.NetworkInterface{
			Name:        iface.Name,
			IPAddresses: ipAddresses,
			MACAddress:  iface.HardwareAddr.String(),
			MTU:         iface.MTU,
			IsUp:        iface.Flags&net.FlagUp != 0,
		}
	}

	return interfaces, nil
}

func (c *Collector) collectRoutes(ctx context.Context) ([]models.Route, error) {
	var routes []models.Route

	if runtime.GOOS == "darwin" {
		return c.collectRoutesDarwin(ctx)
	} else if runtime.GOOS == "linux" {
		return c.collectRoutesLinux(ctx)
	}

	return routes, nil
}

func (c *Collector) collectRoutesDarwin(ctx context.Context) ([]models.Route, error) {
	var routes []models.Route

	cmd := exec.CommandContext(ctx, "netstat", "-nr")
	output, err := cmd.Output()
	if err != nil {
		return routes, err
	}

	lines := strings.Split(string(output), "\n")
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		routes = append(routes, models.Route{
			Destination: fields[0],
			Gateway:     fields[1],
			Interface:   fields[len(fields)-1],
		})
	}

	return routes, nil
}

func (c *Collector) collectRoutesLinux(ctx context.Context) ([]models.Route, error) {
	var routes []models.Route

	cmd := exec.CommandContext(ctx, "ip", "route")
	output, err := cmd.Output()
	if err != nil {
		return routes, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "default") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		dest := ""
		gateway := ""
		iface := ""

		for i := 0; i < len(fields); i++ {
			switch fields[i] {
			case "via":
				if i+1 < len(fields) {
					gateway = fields[i+1]
				}
			case "dev":
				if i+1 < len(fields) {
					iface = fields[i+1]
				}
			}
		}

		if dest != "" || gateway != "" {
			routes = append(routes, models.Route{
				Destination: dest,
				Gateway:     gateway,
				Interface:   iface,
			})
		}
	}

	return routes, nil
}

func (c *Collector) collectDNS(ctx context.Context) (models.DNSConfig, error) {
	dns := models.DNSConfig{
		Nameservers:   []string{},
		SearchDomains: []string{},
	}

	if runtime.GOOS == "darwin" {
		data, err := os.ReadFile("/etc/resolv.conf")
		if err != nil {
			return dns, err
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "nameserver") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					dns.Nameservers = append(dns.Nameservers, parts[1])
				}
			} else if strings.HasPrefix(line, "search") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					dns.SearchDomains = append(dns.SearchDomains, parts[1:]...)
				}
			}
		}
	} else if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/etc/resolv.conf")
		if err != nil {
			return dns, err
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "nameserver") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					dns.Nameservers = append(dns.Nameservers, parts[1])
				}
			} else if strings.HasPrefix(line, "search") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					dns.SearchDomains = append(dns.SearchDomains, parts[1:]...)
				}
			}
		}
	}

	return dns, nil
}

func (c *Collector) collectFirewallRules(ctx context.Context) ([]models.FirewallRule, error) {
	var rules []models.FirewallRule

	if runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(ctx, "pfctl", "-s", "rules")
		output, err := cmd.Output()
		if err != nil {
			return rules, nil
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			rules = append(rules, models.FirewallRule{
				Rule: line,
			})
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.CommandContext(ctx, "iptables", "-L", "-n")
		output, err := cmd.Output()
		if err != nil {
			return rules, nil
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "Chain") || strings.HasPrefix(line, "target") {
				continue
			}

			fields := strings.Fields(line)
			chain := ""
			proto := ""
			source := ""
			dest := ""
			action := ""

			if len(fields) > 0 {
				chain = fields[0]
			}
			if len(fields) > 1 {
				proto = fields[1]
			}
			if len(fields) > 2 {
				source = fields[2]
			}
			if len(fields) > 3 {
				dest = fields[3]
			}
			if len(fields) > 4 {
				action = fields[4]
			}

			rules = append(rules, models.FirewallRule{
				Chain:       chain,
				Protocol:    proto,
				Source:      source,
				Destination: dest,
				Action:      action,
				Rule:        line,
			})
		}
	}

	return rules, nil
}
