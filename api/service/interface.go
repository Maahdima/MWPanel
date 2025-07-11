package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/dataservice/model"
	"mikrotik-wg-go/http/schema"
	"mikrotik-wg-go/utils"
	"strconv"
)

type WgInterface struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
}

func NewWgInterface(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor) *WgInterface {
	return &WgInterface{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
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
			i.logger.Error("failed to fetch wireguard interface from Mikrotik", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))
			return nil, fmt.Errorf("failed to fetch wireguard interface from Mikrotik: %w", err)
		}
		wgInterface := i.transformInterfaceToResponse(iface, *mtInterface.MTU, *mtInterface.Running)
		wgInterfaces = append(wgInterfaces, wgInterface)
	}

	return &wgInterfaces, nil
}

func (i *WgInterface) CreateInterface(req *schema.CreateInterfaceRequest) (*schema.InterfaceResponse, error) {
	wgInterface := &mikrotik.WireGuardInterface{
		Name:       &req.Name,
		Comment:    req.Comment,
		ListenPort: &req.ListenPort,
	}

	mtInterface, err := i.mikrotikAdaptor.CreateWgInterface(context.Background(), *wgInterface)
	if err != nil {
		i.logger.Error("failed to create wireguard interface", zap.Error(err))
		return nil, err
	}

	dbInterface := model.Interface{
		InterfaceID: *mtInterface.ID,
		Comment:     req.Comment,
		Name:        *wgInterface.Name,
		ListenPort:  *wgInterface.ListenPort,
	}

	if err := i.db.Create(&dbInterface).Error; err != nil {
		i.logger.Error("failed to save wireguard interface to database", zap.Error(err))
		return nil, err
	}

	transformedInterface := i.transformInterfaceToResponse(dbInterface, *mtInterface.MTU, *mtInterface.Running)
	return &transformedInterface, nil
}

func (i *WgInterface) ToggleInterfaceStatus(id uint) error {
	var iface model.Interface
	if err := i.db.First(&iface, id).Error; err != nil {
		i.logger.Error("failed to find wireguard interface in database", zap.Error(err))
		return fmt.Errorf("failed to find wireguard interface in database: %w", err)
	}

	disabled := utils.Ptr(strconv.FormatBool(!iface.Disabled))

	wgInterface := mikrotik.WireGuardInterface{
		Disabled: disabled,
	}

	if _, err := i.mikrotikAdaptor.UpdateWgInterface(context.Background(), iface.InterfaceID, wgInterface); err != nil {
		i.logger.Error("failed to update wireguard interface status", zap.Error(err))
		return fmt.Errorf("failed to update wireguard interface status: %w", err)
	}

	if err := i.db.Model(&iface).Update("disabled", disabled).Error; err != nil {
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

	wgInterface := mikrotik.WireGuardInterface{}

	if req.Disabled != nil {
		disabledStr := strconv.FormatBool(*req.Disabled)
		wgInterface.Disabled = &disabledStr
	}
	if req.Comment != nil {
		wgInterface.Comment = req.Comment
	}
	if req.Name != nil {
		wgInterface.Name = req.Name
	}

	mtInterface, err := i.mikrotikAdaptor.UpdateWgInterface(context.Background(), iface.InterfaceID, wgInterface)
	if err != nil {
		i.logger.Error("failed to update wireguard interface", zap.Error(err))
		return nil, fmt.Errorf("failed to update wireguard interface: %w", err)
	}

	iface.Comment = req.Comment
	iface.Name = *wgInterface.Name
	iface.ListenPort = *wgInterface.ListenPort

	if err := i.db.Save(&iface).Error; err != nil {
		i.logger.Error("failed to update wireguard interface in database", zap.Error(err))
		return nil, fmt.Errorf("failed to update wireguard interface in database")
	}

	transformedInterface := i.transformInterfaceToResponse(iface, *mtInterface.MTU, *mtInterface.Running)
	return &transformedInterface, nil
}

func (i *WgInterface) DeleteInterface(id uint) error {
	var iface model.Interface
	if err := i.db.First(&iface, id).Error; err != nil {
		i.logger.Error("failed to find wireguard interface in database", zap.Error(err))
		return fmt.Errorf("failed to find wireguard interface in database: %w", err)
	}

	if err := i.mikrotikAdaptor.DeleteWgInterface(context.Background(), iface.InterfaceID); err != nil {
		i.logger.Error("failed to delete wireguard interface from Mikrotik", zap.Error(err))
		return fmt.Errorf("failed to delete wireguard interface from Mikrotik: %w", err)
	}

	if err := i.db.Unscoped().Delete(&iface).Error; err != nil {
		i.logger.Error("failed to delete wireguard interface from database", zap.Error(err))
		return fmt.Errorf("failed to delete wireguard interface from database: %w", err)
	}

	return nil
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
