package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/http/schema"
)

var ipv4DefaultInterface = "ether1"

type DeviceData struct {
	mikrotikAdaptor  *mikrotik.Adaptor
	serverService    *Server
	interfaceService *WgInterface
	peerService      *WgPeer
	logger           *zap.Logger
}

func NewDeviceData(mikrotikAdaptor *mikrotik.Adaptor, serverService *Server, interfaceService *WgInterface, peerService *WgPeer) *DeviceData {
	return &DeviceData{
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
	ipv4Addresses, err := d.mikrotikAdaptor.FetchIPv4Addresses(context.Background())
	if err != nil {
		d.logger.Error("failed to fetch IPv4 address", zap.Error(err))
		return nil, err
	}

	for _, ipv4 := range *ipv4Addresses {
		if ipv4.Interface == ipv4DefaultInterface {
			return &schema.DeviceIPv4Address{
				IPv4: ipv4.Address,
				// TODO : implement ISP fetching
				ISP: "Iran telecommunication company",
			}, nil
		}
	}

	return nil, nil
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
