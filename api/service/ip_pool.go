package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"
)

type IPPool struct {
	db       *gorm.DB
	mikrotik *mikrotik.Adaptor
	logger   *zap.Logger
}

func NewIPPool(db *gorm.DB, adaptor *mikrotik.Adaptor) *IPPool {
	return &IPPool{
		db:       db,
		mikrotik: adaptor,
		logger:   zap.L().Named("IPPoolService"),
	}
}

func (s *IPPool) GetIPPools() (*[]schema.IPPoolResponse, error) {
	var dbPools []model.IPPool

	if err := s.db.Find(&dbPools).Error; err != nil {
		s.logger.Error("failed to get IP pools", zap.Error(err))

		return nil, err
	}

	var pools []schema.IPPoolResponse

	for _, dbPool := range dbPools {
		pool := s.transformPoolToResponse(dbPool)

		pools = append(pools, pool)
	}

	return &pools, nil
}

func (s *IPPool) UpdateIPPool(id uint, req *schema.UpdateIPPoolRequest) (*schema.IPPoolResponse, error) {
	var pool model.IPPool

	if err := s.db.First(&pool, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("IP pool not found", zap.Uint("id", id))

			return nil, gorm.ErrRecordNotFound
		}

		s.logger.Error("failed to find IP pool", zap.Uint("id", id), zap.Error(err))

		return nil, err
	}

	_, _, err := net.ParseCIDR(req.StartIP)

	if err != nil {
		s.logger.Error("failed to parse start IP", zap.String("start_ip", pool.StartIP), zap.Error(err))

		return nil, err
	}

	_, _, err = net.ParseCIDR(req.EndIP)

	if err != nil {
		s.logger.Error("failed to parse end IP", zap.String("end_ip", pool.EndIP), zap.Error(err))

		return nil, err
	}

	pool.Name = req.Name

	pool.StartIP = req.StartIP

	pool.EndIP = req.EndIP

	if err := s.db.Save(&pool).Error; err != nil {
		s.logger.Error("failed to update IP pool", zap.Uint("id", id), zap.Error(err))

		return nil, err
	}

	resp := s.transformPoolToResponse(pool)

	return &resp, nil
}

func (s *IPPool) CreateIPPool(req *schema.CreateIPPoolRequest) (*schema.IPPoolResponse, error) {
	_, _, err := net.ParseCIDR(req.StartIP)

	if err != nil {
		s.logger.Error("failed to parse start IP", zap.String("start_ip", req.StartIP), zap.Error(err))

		return nil, err
	}

	_, _, err = net.ParseCIDR(req.EndIP)

	if err != nil {
		s.logger.Error("failed to parse end IP", zap.String("end_ip", req.EndIP), zap.Error(err))

		return nil, err
	}

	var iface model.Interface

	if err := s.db.First(&iface, req.InterfaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("interface not found", zap.Uint("id", req.InterfaceID))

			return nil, gorm.ErrRecordNotFound
		}

		s.logger.Error("failed to find interface", zap.Uint("id", req.InterfaceID), zap.Error(err))

		return nil, err
	}

	pool := model.IPPool{
		ServerID:    iface.ServerID,
		Name:        req.Name,
		StartIP:     req.StartIP,
		EndIP:       req.EndIP,
		InterfaceID: iface.ID,
	}

	if err := s.db.Create(&pool).Error; err != nil {
		s.logger.Error("failed to create IP pool", zap.Error(err))

		return nil, err
	}

	resp := s.transformPoolToResponse(pool)

	return &resp, nil
}

func (s *IPPool) DeleteIPPool(id uint) error {
	var pool model.IPPool

	if err := s.db.First(&pool, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("IP pool not found", zap.Uint("id", id))

			return gorm.ErrRecordNotFound
		}

		s.logger.Error("failed to find IP pool", zap.Uint("id", id), zap.Error(err))

		return err
	}

	if err := s.db.Unscoped().Delete(&pool).Error; err != nil {
		s.logger.Error("failed to delete IP pool", zap.Uint("id", id), zap.Error(err))

		return err
	}

	return nil
}

func (s *IPPool) transformPoolToResponse(pool model.IPPool) schema.IPPoolResponse {
	startIP, _, err := net.ParseCIDR(pool.StartIP)

	if err != nil {
		s.logger.Error("failed to parse start IP", zap.String("start_ip", pool.StartIP), zap.Error(err))

		return schema.IPPoolResponse{}
	}

	endIP, _, err := net.ParseCIDR(pool.EndIP)

	if err != nil {
		s.logger.Error("failed to parse end IP", zap.String("end_ip", pool.EndIP), zap.Error(err))

		return schema.IPPoolResponse{}
	}

	totalIPs, usedIPs, remainingIPs, err := s.getIPCounts(pool, startIP, endIP)

	if err != nil {
		s.logger.Error("failed to get IP counts", zap.Error(err))

		return schema.IPPoolResponse{}
	}

	return schema.IPPoolResponse{
		Id:          pool.ID,
		Name:        pool.Name,
		StartIP:     startIP.String(),
		EndIP:       endIP.String(),
		TotalIP:     totalIPs,
		UsedIP:      usedIPs,
		RemainingIP: remainingIPs,
	}
}

func (s *IPPool) getIPCounts(pool model.IPPool, startIP, endIP net.IP) (totalIPs, usedIPs, remainingIPs int, err error) {
	start := utils.IPToUint32(startIP)

	end := utils.IPToUint32(endIP)

	if end < start {
		s.logger.Error("end IP is less than start IP", zap.String("start_ip", startIP.String()), zap.String("end_ip", endIP.String()))

		return 0, 0, 0, fmt.Errorf("end IP cannot be before start IP")
	}

	var peers []model.Peer

	if err := s.db.Model(&model.Peer{}).Where("server_id = ?", pool.ServerID).
		Find(&peers).Error; err != nil {
		s.logger.Error("failed to fetch peers", zap.Error(err))

		return 0, 0, 0, err
	}

	usedSet := map[uint32]struct{}{}

	for _, peer := range peers {
		ipStr := strings.Split(peer.AllowedAddress, "/")[0]

		if ip := net.ParseIP(ipStr); ip != nil {
			ipUint32 := utils.IPToUint32(ip)

			if ipUint32 >= start && ipUint32 <= end {
				usedSet[ipUint32] = struct{}{}
			}
		}
	}

	for n := range s.reservedInPool(pool, start, end) {
		usedSet[n] = struct{}{}
	}

	totalIPs = int(end - start + 1)

	usedIPs = len(usedSet)

	remainingIPs = totalIPs - usedIPs

	if remainingIPs < 0 {
		s.logger.Warn("remaining IPs is negative; resetting to zero", zap.Int("used", usedIPs), zap.Int("total", totalIPs))

		remainingIPs = 0
	}

	return totalIPs, usedIPs, remainingIPs, nil
}

func (s *IPPool) reservedInPool(pool model.IPPool, start, end uint32) map[uint32]struct{} {
	reserved := map[uint32]struct{}{}

	if s.mikrotik == nil || pool.InterfaceID == 0 {
		return reserved
	}

	var iface model.Interface

	if err := s.db.First(&iface, pool.InterfaceID).Error; err != nil {
		return reserved
	}

	adaptor, err := s.mikrotik.ForServer(iface.ServerID)

	if err != nil {
		return reserved
	}

	addrs, err := adaptor.FetchIPv4Addresses(context.Background())

	if err != nil {
		s.logger.Warn("failed to read interface addresses for IP pool counts",
			zap.String("interface", iface.Name),
			zap.Error(err))

		return reserved
	}

	for _, addr := range addrs {
		if !strings.EqualFold(addr.Interface, iface.Name) && !strings.EqualFold(addr.ActualInterface, iface.Name) {
			continue
		}

		for _, ip := range utils.ReservedIPv4s(addr.Address) {
			n := utils.IPToUint32(ip)

			if n >= start && n <= end {
				reserved[n] = struct{}{}
			}
		}
	}

	return reserved
}
