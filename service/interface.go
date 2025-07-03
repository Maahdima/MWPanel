package service

import (
	"context"
	"go.uber.org/zap"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/api/schema"
)

type WgInterface struct {
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
}

func NewWgInterface(mikrotikAdaptor *mikrotik.Adaptor) *WgInterface {
	return &WgInterface{
		mikrotikAdaptor: mikrotikAdaptor,
		logger:          zap.L().Named("WgInterfaceService"),
	}
}

func (w *WgInterface) GetInterfaces() (*[]mikrotik.WireGuardInterface, error) {
	return nil, nil
}

func (w *WgInterface) CreateInterface(name string, listenPort int) (*mikrotik.WireGuardInterface, error) {
	return nil, nil
}

func (w *WgInterface) DeleteInterface(name string) error {
	return nil
}

func (w *WgInterface) GetInterfacesData() (*schema.InterfacesDataResponse, error) {
	var totalServers int
	var activeServers int

	interfaces, err := w.mikrotikAdaptor.FetchWgInterfaces(context.Background())
	if err != nil {
		w.logger.Error("failed to get wireguard interfaces", zap.Error(err))
		return nil, err
	}

	totalServers = len(*interfaces)

	for _, iface := range *interfaces {
		if iface.Disabled == "false" {
			activeServers++
			continue
		}
	}

	return &schema.InterfacesDataResponse{
		TotalServers:  totalServers,
		ActiveServers: activeServers,
	}, nil
}
