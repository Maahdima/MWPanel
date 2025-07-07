package service

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/dataservice/model"
	"strconv"
)

type Server struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
}

func NewServerService(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor) *Server {
	return &Server{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		logger:          zap.L().Named("ServerService"),
	}
}

func (s *Server) ToggleServerStatus(id uint) (*schema.ServerResponse, error) {
	var server model.Server
	if err := s.db.First(&server, id).Error; err != nil {
		s.logger.Error("failed to find server by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	server.IsActive = !server.IsActive

	if err := s.db.Save(&server).Error; err != nil {
		s.logger.Error("failed to update server status", zap.Error(err))
		return nil, err
	}

	return &schema.ServerResponse{
		Id:        server.ID,
		Comment:   &server.Comment,
		Name:      server.Name,
		IPAddress: server.IPAddress,
		APIPort:   strconv.Itoa(server.APIPort),
		IsActive:  server.IsActive,
		Status:    schema.AvailableServer,
	}, nil
}

func (s *Server) CreateServer(req *schema.CreateServerRequest) (*schema.ServerResponse, error) {
	apiPort, err := strconv.Atoi(req.APIPort)
	if err != nil {
		s.logger.Error("invalid API port", zap.Error(err))
		return nil, err
	}

	_, err = s.mikrotikAdaptor.FetchDeviceIdentity(context.Background())
	if err != nil {
		s.logger.Error("failed to connect to device when creating a new server", zap.Error(err))
		return nil, err
	}

	server := model.Server{
		Comment:   *req.Comment,
		Name:      req.Name,
		IPAddress: req.IPAddress,
		APIPort:   apiPort,
		Username:  req.Username,
		Password:  req.Password,
	}

	if err := s.db.Create(&server).Error; err != nil {
		s.logger.Error("failed to create server record", zap.Error(err))
		return nil, err
	}

	return &schema.ServerResponse{
		Id:        server.ID,
		Comment:   &server.Comment,
		Name:      server.Name,
		IPAddress: server.IPAddress,
		APIPort:   strconv.Itoa(server.APIPort),
		IsActive:  server.IsActive,
		Status:    schema.AvailableServer,
	}, nil
}

func (s *Server) GetServers() (*[]schema.ServerResponse, error) {
	var servers []model.Server
	if err := s.db.Find(&servers).Error; err != nil {
		s.logger.Error("failed to fetch servers from database", zap.Error(err))
		return nil, err
	}

	var serverResponses []schema.ServerResponse
	for _, server := range servers {
		var serverStatus schema.ServerStatus

		// TODO : call the specific server's API to check its status
		_, err := s.mikrotikAdaptor.FetchDeviceIdentity(context.Background())
		if err != nil {
			serverStatus = schema.NotAvailableServer
			continue
		} else {
			serverStatus = schema.AvailableServer
		}

		serverResponses = append(serverResponses, schema.ServerResponse{
			Id:        server.ID,
			Comment:   &server.Comment,
			Name:      server.Name,
			IPAddress: server.IPAddress,
			APIPort:   strconv.Itoa(server.APIPort),
			IsActive:  server.IsActive,
			Status:    serverStatus,
		})
	}

	return &serverResponses, nil
}

func (s *Server) UpdateServer(id uint, req *schema.UpdateServerRequest) (*schema.ServerResponse, error) {
	var server model.Server
	if err := s.db.First(&server, id).Error; err != nil {
		s.logger.Error("failed to find server by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	server.Comment = *req.Comment
	server.Name = *req.Name
	server.IPAddress = *req.IPAddress

	apiPort, err := strconv.Atoi(*req.APIPort)
	if err != nil {
		s.logger.Error("invalid API port", zap.Error(err))
		return nil, err
	}
	server.APIPort = apiPort

	server.Username = *req.Username
	server.Password = *req.Password

	if err := s.db.Save(&server).Error; err != nil {
		s.logger.Error("failed to update server record", zap.Error(err))
		return nil, err
	}

	return &schema.ServerResponse{
		Id:        server.ID,
		Comment:   &server.Comment,
		Name:      server.Name,
		IPAddress: server.IPAddress,
		APIPort:   strconv.Itoa(server.APIPort),
		IsActive:  server.IsActive,
		Status:    schema.AvailableServer,
	}, nil
}

func (s *Server) DeleteServer(id uint) error {
	var server model.Server
	if err := s.db.First(&server, id).Error; err != nil {
		s.logger.Error("failed to find server by ID", zap.Uint("id", id), zap.Error(err))
		return err
	}

	if err := s.db.Unscoped().Delete(&server).Error; err != nil {
		s.logger.Error("failed to delete server record", zap.Error(err))
		return err
	}

	return nil
}

func (s *Server) GetServersData() (*schema.ServerStatsResponse, error) {
	var totalServers int64
	if err := s.db.Model(&model.Server{}).Count(&totalServers).Error; err != nil {
		s.logger.Error("failed to count total servers", zap.Error(err))
		return nil, err
	}

	var activeServers int64
	if err := s.db.Model(&model.Server{}).Where("is_active = ?", true).Count(&activeServers).Error; err != nil {
		s.logger.Error("failed to count active servers", zap.Error(err))
		return nil, err
	}

	return &schema.ServerStatsResponse{
		TotalServers:  int(totalServers),
		ActiveServers: int(activeServers),
	}, nil
}
