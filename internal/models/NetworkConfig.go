package models

type NetworkInterface struct {
	Name        string   `json:"name" yaml:"name"`
	IPAddresses []string `json:"ip_addresses" yaml:"ip_addresses"`
	MACAddress  string   `json:"mac_address" yaml:"mac_address"`
	MTU         int      `json:"mtu" yaml:"mtu"`
	IsUp        bool     `json:"is_up" yaml:"is_up"`
}

type Route struct {
	Destination string `json:"destination" yaml:"destination"`
	Gateway     string `json:"gateway" yaml:"gateway"`
	Interface   string `json:"interface" yaml:"interface"`
	Metric      int    `json:"metric" yaml:"metric"`
}

type DNSConfig struct {
	Nameservers   []string `json:"nameservers" yaml:"nameservers"`
	SearchDomains []string `json:"search_domains" yaml:"search_domains"`
}

type FirewallRule struct {
	Chain       string `json:"chain" yaml:"chain"`
	Rule        string `json:"rule" yaml:"rule"`
	Action      string `json:"action" yaml:"action"`
	Protocol    string `json:"protocol" yaml:"protocol"`
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
}

type NetworkConfig struct {
	Interfaces    map[string]NetworkInterface `json:"interfaces" yaml:"interfaces"`
	Routes        []Route                     `json:"routes" yaml:"routes"`
	DNS           DNSConfig                   `json:"dns" yaml:"dns"`
	FirewallRules []FirewallRule              `json:"firewall_rules,omitempty" yaml:"firewall_rules,omitempty"`
}
