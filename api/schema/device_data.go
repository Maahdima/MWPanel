package schema

type DeviceDataResponse struct {
	*DeviceIdentity
	*DeviceInfo
	*DeviceIPv4Address
	*DNSConfig
}

type DeviceInfo struct {
	BoardName   string `json:"board_name"`
	OSVersion   string `json:"os_version"`
	CpuArch     string `json:"cpu_arch"`
	Uptime      string `json:"uptime"`
	CpuLoad     string `json:"cpu_load"`
	TotalMemory string `json:"total_memory"`
	FreeMemory  string `json:"free_memory"`
	TotalDisk   string `json:"total_disk"`
	FreeDisk    string `json:"free_disk"`
}

type DeviceResource struct {
	Uptime      string `json:"uptime"`
	CpuUsage    string `json:"cpu_usage"`
	MemoryUsage string `json:"memory_usage"`
	DiskUsage   string `json:"disk_usage"`
}

type DeviceIdentity struct {
	Identity string `json:"identity"`
}

type DeviceIPv4Address struct {
	IPv4 string `json:"ipv4,omitempty"`
	ISP  string `json:"isp,omitempty"`
}

type DNSConfig struct {
	DnsServer string `json:"dns_servers,omitempty"`
}
