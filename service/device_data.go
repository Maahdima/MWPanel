package service

import (
	"context"
	"go.uber.org/zap"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/api/schema"
)

var ipv4DefaultInterface = "ether1"

type DeviceData struct {
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
}

func NewDeviceData(mikrotikAdaptor *mikrotik.Adaptor) *DeviceData {
	return &DeviceData{
		mikrotikAdaptor: mikrotikAdaptor,
		logger:          zap.L().Named("DeviceDataService"),
	}
}

func (d *DeviceData) GetDeviceData() (*schema.DeviceDataResponse, error) {
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

	return &schema.DeviceDataResponse{
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
