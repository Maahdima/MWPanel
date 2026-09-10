package service

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
)

type WgInterface struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	peerService     *WgPeer
	logger          *zap.Logger
}

func NewWgInterface(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, peerService *WgPeer) *WgInterface {
	return &WgInterface{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		peerService:     peerService,
		logger:          zap.L().Named("WgInterfaceService"),
	}
}

func (i *WgInterface) GetInterfaces() (*[]schema.InterfaceResponse, error) {
	var interfaces []model.Interface
	if err := i.db.Order("created_at desc").Find(&interfaces).Error; err != nil {
		i.logger.Error("failed to get wireguard interfaces from database", zap.Error(err))
		return nil, err
	}

	var wgInterfaces []schema.InterfaceResponse
	for _, iface := range interfaces {
		mtInterface, err := i.mikrotikAdaptor.FetchWgInterface(context.Background(), iface.InterfaceID)
		if err != nil {
			i.logger.Warn("failed to fetch wireguard interface from Mikrotik; using local data",
				zap.String("interfaceID", iface.InterfaceID),
				zap.Error(err),
			)
			wgInterfaces = append(wgInterfaces, i.transformInterfaceToResponse(iface, "", "false"))
			continue
		}
		wgInterfaces = append(wgInterfaces, i.transformInterfaceToResponse(iface, mtInterface.MTU, *mtInterface.Running))
	}

	return &wgInterfaces, nil
}

func (i *WgInterface) CreateInterface(req *schema.CreateInterfaceRequest) (*schema.InterfaceResponse, error) {
	wgInterface := &mikrotik.WireGuardInterface{
		Name:       req.Name,
		Comment:    req.Comment,
		ListenPort: req.ListenPort,
	}

	mtInterface, err := i.mikrotikAdaptor.CreateWgInterface(context.Background(), *wgInterface)
	if err != nil {
		i.logger.Error("failed to create wireguard interface", zap.Error(err))
		return nil, err
	}

	dbInterface := model.Interface{
		InterfaceID: mtInterface.ID,
		Comment:     wgInterface.Comment,
		Name:        wgInterface.Name,
		PrivateKey:  mtInterface.PrivateKey,
		PublicKey:   mtInterface.PublicKey,
		ListenPort:  wgInterface.ListenPort,
	}

	if err := i.db.Create(&dbInterface).Error; err != nil {
		i.logger.Error("failed to save wireguard interface to database", zap.Error(err))
		if delErr := i.mikrotikAdaptor.DeleteWgInterface(context.Background(), mtInterface.ID); delErr != nil {
			i.logger.Warn("failed to rollback mikrotik interface after db create failure", zap.Error(delErr))
		}
		return nil, err
	}

	transformedInterface := i.transformInterfaceToResponse(dbInterface, mtInterface.MTU, *mtInterface.Running)
	return &transformedInterface, nil
}

func (i *WgInterface) ToggleInterfaceStatus(id uint) error {
	var iface model.Interface
	if err := i.db.First(&iface, id).Error; err != nil {
		i.logger.Error("failed to find wireguard interface in database", zap.Error(err))
		return fmt.Errorf("failed to find wireguard interface in database: %w", err)
	}

	newDisabled := !iface.Disabled
	disabled := strconv.FormatBool(newDisabled)

	wgInterface := mikrotik.WireGuardInterface{
		Disabled: disabled,
	}

	if _, err := i.mikrotikAdaptor.UpdateWgInterface(context.Background(), iface.InterfaceID, wgInterface); err != nil {
		i.logger.Error("failed to update wireguard interface status", zap.Error(err))
		return fmt.Errorf("failed to update wireguard interface status: %w", err)
	}

	if err := i.db.Model(&iface).Update("disabled", newDisabled).Error; err != nil {
		i.logger.Error("failed to update interface status in database", zap.Error(err))
		return fmt.Errorf("failed to update interface status in database: %w", err)
	}

	return nil
}

func (i *WgInterface) UpdateInterface(id uint, req *schema.UpdateInterfaceRequest) (*schema.InterfaceResponse, error) {
	var iface model.Interface
	if err := i.db.First(&iface, id).Error; err != nil {
		i.logger.Error("failed to get interface from database", zap.Error(err))
		return nil, err
	}

	oldName := iface.Name
	wgInterface := mikrotik.WireGuardInterface{}

	if req.Disabled != nil {
		disabledStr := strconv.FormatBool(*req.Disabled)
		wgInterface.Disabled = disabledStr
		iface.Disabled = *req.Disabled
	}
	if req.Comment != nil {
		wgInterface.Comment = req.Comment
		iface.Comment = req.Comment
	}

	if req.Name != "" {
		wgInterface.Name = req.Name
		iface.Name = req.Name
	}

	mtInterface, err := i.mikrotikAdaptor.UpdateWgInterface(context.Background(), iface.InterfaceID, wgInterface)
	if err != nil {
		i.logger.Error("failed to update wireguard interface", zap.Error(err))
		return nil, fmt.Errorf("failed to update wireguard interface: %w", err)
	}

	if mtInterface.ListenPort != "" {
		iface.ListenPort = mtInterface.ListenPort
	}

	if err := i.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&iface).Error; err != nil {
			return err
		}
		if oldName != iface.Name {
			if err := tx.Model(&model.Peer{}).
				Where("interface = ?", oldName).
				Update("interface", iface.Name).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		i.logger.Error("failed to update wireguard interface in database", zap.Error(err))
		return nil, fmt.Errorf("failed to update wireguard interface in database")
	}

	transformedInterface := i.transformInterfaceToResponse(iface, mtInterface.MTU, *mtInterface.Running)
	return &transformedInterface, nil
}

func (i *WgInterface) DeleteInterface(id uint) error {
	var iface model.Interface
	if err := i.db.First(&iface, id).Error; err != nil {
		i.logger.Error("failed to find wireguard interface in database", zap.Error(err))
		return fmt.Errorf("failed to find wireguard interface in database: %w", err)
	}

	return i.deleteInterfaceRecord(&iface)
}

func (i *WgInterface) GetInterfacesData() (*schema.InterfaceStatsResponse, error) {
	var totalInterfaces int64
	if err := i.db.Model(&model.Interface{}).Count(&totalInterfaces).Error; err != nil {
		i.logger.Error("failed to count total interfaces", zap.Error(err))
		return nil, fmt.Errorf("failed to count total interfaces: %w", err)
	}

	var activeInterfaces int64
	if err := i.db.Model(&model.Interface{}).Where("disabled = ?", false).Count(&activeInterfaces).Error; err != nil {
		i.logger.Error("failed to count active interfaces", zap.Error(err))
		return nil, fmt.Errorf("failed to count active interfaces: %w", err)
	}

	return &schema.InterfaceStatsResponse{
		TotalInterfaces:  int(totalInterfaces),
		ActiveInterfaces: int(activeInterfaces),
	}, nil
}

func (i *WgInterface) deleteInterfaceRecord(iface *model.Interface) error {
	if iface == nil {
		return fmt.Errorf("interface is required")
	}

	if i.peerService != nil {
		if err := i.peerService.DeletePeersByInterfaceName(iface.Name); err != nil {
			return fmt.Errorf("failed to delete peers for interface: %w", err)
		}
	}

	if err := i.db.Unscoped().Where("interface_id = ?", iface.ID).Delete(&model.Traffic{}).Error; err != nil {
		i.logger.Error("failed to delete interface traffic records", zap.Uint("interfaceID", iface.ID), zap.Error(err))
		return fmt.Errorf("failed to delete interface traffic records: %w", err)
	}

	if err := i.db.Unscoped().Where("interface_id = ?", iface.ID).Delete(&model.IPPool{}).Error; err != nil {
		i.logger.Error("failed to delete interface IP pool", zap.Uint("interfaceID", iface.ID), zap.Error(err))
		return fmt.Errorf("failed to delete interface IP pool: %w", err)
	}

	if err := i.mikrotikAdaptor.DeleteWgInterface(context.Background(), iface.InterfaceID); err != nil {
		i.logger.Error("failed to delete wireguard interface from Mikrotik", zap.Error(err))
		return fmt.Errorf("failed to delete wireguard interface from Mikrotik: %w", err)
	}

	if err := i.db.Unscoped().Delete(iface).Error; err != nil {
		i.logger.Error("failed to delete wireguard interface from database", zap.Error(err))
		return fmt.Errorf("failed to delete wireguard interface from database: %w", err)
	}

	return nil
}

func (i *WgInterface) transformInterfaceToResponse(wgInterface model.Interface, mtu, status string) schema.InterfaceResponse {
	return schema.InterfaceResponse{
		Id:          wgInterface.ID,
		InterfaceID: wgInterface.InterfaceID,
		Disabled:    wgInterface.Disabled,
		Comment:     wgInterface.Comment,
		Name:        wgInterface.Name,
		ListenPort:  wgInterface.ListenPort,
		MTU:         mtu,
		IsRunning:   status == "true",
	}
}
