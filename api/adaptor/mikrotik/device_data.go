package mikrotik

import (
	"context"

	"github.com/maahdima/mwp/api/common"
)

type SystemIdentity struct {
	Name string `json:"name"`
}

type SystemInfo struct {
	ArchitectureName     string `json:"architecture-name"`
	BadBlocks            string `json:"bad-blocks"`
	BoardName            string `json:"board-name"`
	BuildTime            string `json:"build-time"`
	CPU                  string `json:"cpu"`
	CPUCount             string `json:"cpu-count"`
	CPUFrequency         string `json:"cpu-frequency"`
	CPULoad              string `json:"cpu-load"`
	FactorySoftware      string `json:"factory-software"`
	FreeHDDSpace         string `json:"free-hdd-space"`
	FreeMemory           string `json:"free-memory"`
	Platform             string `json:"platform"`
	TotalHDDSpace        string `json:"total-hdd-space"`
	TotalMemory          string `json:"total-memory"`
	Uptime               string `json:"uptime"`
	Version              string `json:"version"`
	WriteSectSinceReboot string `json:"write-sect-since-reboot"`
	WriteSectTotal       string `json:"write-sect-total"`
}

type DNSConfig struct {
	CacheMaxTTL              string `json:"cache-max-ttl"`
	DohTimeout               string `json:"doh-timeout"`
	QueryServerTimeout       string `json:"query-server-timeout"`
	QueryTotalTimeout        string `json:"query-total-timeout"`
	AllowRemoteRequests      string `json:"allow-remote-requests"`
	CacheSize                string `json:"cache-size"`
	CacheUsed                string `json:"cache-used"`
	MaxConcurrentQueries     string `json:"max-concurrent-queries"`
	MaxConcurrentTCPSessions string `json:"max-concurrent-tcp-sessions"`
	MaxUDPPacketSize         string `json:"max-udp-packet-size"`
	DynamicServers           string `json:"dynamic-servers"`
	Servers                  string `json:"servers"`
	UseDohServer             string `json:"use-doh-server"`
}

type IPAddress struct {
	ID              string `json:".id"`
	ActualInterface string `json:"actual-interface"`
	Address         string `json:"address"`
	Interface       string `json:"interface"`
	Network         string `json:"network"`
}

func (a *Adaptor) FetchDeviceInfo(c context.Context) (*SystemInfo, error) {
	var systemInfo SystemInfo

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Get(
		c,
		common.DeviceInfoPath,
		&systemInfo,
	)
	if err != nil {
		return nil, err
	}

	return &systemInfo, nil
}

func (a *Adaptor) FetchDeviceIdentity(c context.Context) (*SystemIdentity, error) {
	var systemIdentity SystemIdentity

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Get(
		c,
		common.DeviceIdentityPath,
		&systemIdentity,
	)
	if err != nil {
		return nil, err
	}

	return &systemIdentity, nil
}

func (a *Adaptor) FetchDNSConfig(c context.Context) (*DNSConfig, error) {
	var dnsConfig DNSConfig

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Get(
		c,
		common.DeviceDnsPath,
		&dnsConfig,
	)
	if err != nil {
		return nil, err
	}

	return &dnsConfig, nil
}

func (a *Adaptor) FetchIPv4Addresses(c context.Context) (*[]IPAddress, error) {
	var ipv4Address []IPAddress

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Get(
		c,
		common.DeviceIPv4Path,
		&ipv4Address,
	)
	if err != nil {
		return nil, err
	}

	return &ipv4Address, nil
}
