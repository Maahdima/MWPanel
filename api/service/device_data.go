package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
)

type IPApiResponse struct {
	Status  string `json:"status"`
	Country string `json:"country"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

type DeviceData struct {
	db               *gorm.DB
	mikrotikAdaptor  *mikrotik.Adaptor
	serverService    *Server
	interfaceService *WgInterface
	peerService      *WgPeer
	logger           *zap.Logger
}

func NewDeviceData(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, serverService *Server, interfaceService *WgInterface, peerService *WgPeer) *DeviceData {
	return &DeviceData{
		db:               db,
		mikrotikAdaptor:  mikrotikAdaptor,
		serverService:    serverService,
		interfaceService: interfaceService,
		peerService:      peerService,
		logger:           zap.L().Named("DeviceDataService"),
	}
}

func (d *DeviceData) GetDeviceData() (*schema.DeviceStatsResponse, error) {
	serverStats, err := d.serverService.GetServersData()
	if err != nil {
		d.logger.Error("failed to fetch server stats", zap.Error(err))
		return nil, err
	}

	interfaceStats, err := d.interfaceService.GetInterfacesData()
	if err != nil {
		return nil, err
	}

	peerStats, err := d.peerService.GetPeersData()
	if err != nil {
		return nil, err
	}

	info, err := d.getDeviceInfo()
	if err != nil {
		return nil, err
	}

	identity, err := d.getDeviceIdentity()
	if err != nil {
		return nil, err
	}

	ipv4, err := d.getDeviceIpAddress()
	if err != nil {
		return nil, err
	}

	dns, err := d.getDNSConfig()
	if err != nil {
		return nil, err
	}

	return &schema.DeviceStatsResponse{
		ServerInfo:        serverStats,
		InterfaceInfo:     interfaceStats,
		PeerInfo:          peerStats,
		DeviceInfo:        info,
		DeviceIdentity:    identity,
		DeviceIPv4Address: ipv4,
		DNSConfig:         dns,
	}, nil
}

func (d *DeviceData) getDeviceInfo() (*schema.DeviceInfo, error) {
	info, err := d.mikrotikAdaptor.FetchDeviceInfo(context.Background())
	if err != nil {
		d.logger.Error("failed to fetch device resource", zap.Error(err))
		return nil, err
	}

	return &schema.DeviceInfo{
		BoardName:   info.BoardName,
		OSVersion:   info.Version,
		CpuArch:     info.ArchitectureName,
		Uptime:      info.Uptime,
		CpuLoad:     info.CPULoad,
		TotalMemory: info.TotalMemory,
		FreeMemory:  info.FreeMemory,
		TotalDisk:   info.TotalHDDSpace,
		FreeDisk:    info.FreeHDDSpace,
	}, nil
}

func (d *DeviceData) getDeviceIdentity() (*schema.DeviceIdentity, error) {
	identity, err := d.mikrotikAdaptor.FetchDeviceIdentity(context.Background())
	if err != nil {
		d.logger.Error("failed to fetch device identity", zap.Error(err))
		return nil, err
	}

	return &schema.DeviceIdentity{
		Identity: identity.Name,
	}, nil
}

func (d *DeviceData) getDeviceIpAddress() (*schema.DeviceIPv4Address, error) {
	deviceIpData := &schema.DeviceIPv4Address{}

	ipv4Addresses, err := d.mikrotikAdaptor.FetchIPv4Addresses(context.Background())
	if err != nil {
		d.logger.Error("failed to fetch IPv4 address", zap.Error(err))
		return deviceIpData, err
	}

	for _, ipv4 := range ipv4Addresses {
		if ipv4.Interface == common.IPv4DefaultInterface {
			deviceIpData.IPv4 = ipv4.Address
		}
	}

	// TODO : multi-server support
	var server model.Server
	if err := d.db.First(&server).Error; err != nil {
		d.logger.Error("failed to fetch server record from database", zap.Error(err))
		return deviceIpData, err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	ipApiURL := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,city,isp", "37.255.200.79")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ipApiURL, nil)
	if err != nil {
		d.logger.Error("failed to create IP API request", zap.Error(err))
		return deviceIpData, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		d.logger.Error("HTTP request to IP API failed", zap.Error(err))
		return deviceIpData, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d.logger.Error("unexpected response status from IP API", zap.Int("status", resp.StatusCode))
		return deviceIpData, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.logger.Error("failed to read response body from IP API", zap.Error(err))
		return deviceIpData, nil
	}

	var respBody IPApiResponse
	if err := json.Unmarshal(body, &respBody); err != nil {
		d.logger.Error("failed to parse IP API response", zap.Error(err))
		return deviceIpData, nil
	}

	deviceIpData.ISP = respBody.ISP

	return deviceIpData, err

}

func (d *DeviceData) getDNSConfig() (*schema.DNSConfig, error) {
	dns, err := d.mikrotikAdaptor.FetchDNSConfig(context.Background())
	if err != nil {
		d.logger.Error("failed to fetch dns config", zap.Error(err))
		return nil, err
	}

	return &schema.DNSConfig{
		// TODO : check this
		DnsServer: dns.Servers + dns.DynamicServers,
	}, nil
}
