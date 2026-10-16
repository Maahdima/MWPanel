package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
)

func (w *WgPeer) ShopInterface(interfaceID uint) (model.Interface, error) {
	var iface model.Interface
	query := w.db.Preload("IPPool")
	if interfaceID > 0 {
		err := query.First(&iface, "id = ?", interfaceID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return iface, fmt.Errorf("interface %d not found", interfaceID)
			}
			return iface, err
		}
		return iface, nil
	}

	err := query.Where("disabled = ?", false).Order("id asc").First(&iface).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = w.db.Preload("IPPool").Order("id asc").First(&iface).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return iface, errors.New("no wireguard interface is configured")
		}
		return iface, err
	}
	return iface, nil
}

func (w *WgPeer) CreateShopPeer(name, endpoint, comment string, interfaceID uint, trafficGB, durationDays int) (model.Peer, error) {
	iface, err := w.ShopInterface(interfaceID)
	if err != nil {
		return model.Peer{}, err
	}
	if iface.IPPool == nil {
		return model.Peer{}, errors.New("wireguard interface has no IP pool")
	}
	if strings.TrimSpace(endpoint) == "" {
		return model.Peer{}, errors.New("vpn endpoint is not configured")
	}

	creds, err := w.GetPeerCredentials()
	if err != nil {
		return model.Peer{}, err
	}
	address, err := w.GetNewPeerAllowedAddress(iface.ID)
	if err != nil {
		return model.Peer{}, err
	}
	if strings.TrimSpace(address.AllowedAddress) == "" {
		return model.Peer{}, errors.New("could not allocate an IP address")
	}

	expire := time.Now().AddDate(0, 0, durationDays).Format("2006-01-02")
	traffic := strconv.Itoa(trafficGB)
	comment = strings.TrimSpace(comment)
	req := &schema.CreatePeerRequest{
		Comment:        &comment,
		Name:           strings.TrimSpace(name),
		InterfaceId:    iface.ID,
		PrivateKey:     creds.PrivateKey,
		PublicKey:      creds.PublicKey,
		AllowedAddress: address.AllowedAddress,
		Endpoint:       strings.TrimSpace(endpoint),
		ExpireTime:     &expire,
		TrafficLimit:   &traffic,
	}

	created, err := w.CreatePeer(req)
	if err != nil {
		return model.Peer{}, err
	}

	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", created.Id).Error; err != nil {
		return model.Peer{}, err
	}
	if err := w.db.Model(&peer).Update("is_shared", true).Error; err != nil {
		w.logger.Warn("failed to mark shop peer as shared", zap.Error(err), zap.Uint("id", peer.ID))
	} else {
		peer.IsShared = true
	}
	return peer, nil
}

func (w *WgPeer) ApplyCharge(peerID uint, addBytes int64, addDays int) (model.Peer, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", peerID).Error; err != nil {
		return model.Peer{}, err
	}
	if addBytes <= 0 || addDays <= 0 {
		return model.Peer{}, errors.New("invalid vpn charge")
	}

	used := peer.DownloadUsage + peer.UploadUsage
	newLimit := chargedTrafficLimit(peer.TrafficLimit, used, addBytes)
	expire := extendedExpireDate(peer.ExpireTime, addDays, time.Now())
	expirePtr := expire

	schedulerID, err := w.handleScheduler(&peer, &schema.UpdatePeerRequest{ExpireTime: &expirePtr})
	if err != nil {
		return model.Peer{}, err
	}

	wasDisabled := peer.Disabled
	updates := map[string]interface{}{
		"traffic_limit":         newLimit,
		"expire_time":           expire,
		"scheduler_id":          schedulerID,
		"disabled":              false,
		"traffic_first_notify":  false,
		"traffic_second_notify": false,
		"traffic_third_notify":  false,
		"expire_first_notify":   false,
		"expire_second_notify":  false,
		"expire_third_notify":   false,
	}
	if err := w.db.Model(&peer).Updates(updates).Error; err != nil {
		w.logger.Error("failed to store vpn charge", zap.Error(err), zap.Uint("id", peer.ID))
		return model.Peer{}, err
	}

	if wasDisabled {
		if _, err := w.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{
			Disabled: strconv.FormatBool(false),
		}); err != nil {
			_ = w.db.Model(&peer).Update("disabled", true).Error
			w.logger.Error("failed to enable charged peer", zap.Error(err), zap.String("peerID", peer.PeerID))
			return model.Peer{}, err
		}
		if peer.QueueID != nil {
			if _, err := w.mikrotikAdaptor.UpdateSimpleQueue(context.Background(), *peer.QueueID, mikrotik.Queue{
				Disabled: strconv.FormatBool(false),
			}); err != nil {
				w.logger.Warn("failed to enable queue for charged peer", zap.Error(err), zap.String("peerID", peer.PeerID))
			}
		}
		if peer.SchedulerID != nil {
			if _, err := w.mikrotikAdaptor.UpdateScheduler(context.Background(), *peer.SchedulerID, mikrotik.Scheduler{
				Disabled: strconv.FormatBool(false),
			}); err != nil {
				w.logger.Warn("failed to enable scheduler for charged peer", zap.Error(err), zap.String("peerID", peer.PeerID))
			}
		}
	}

	if err := w.db.First(&peer, "id = ?", peerID).Error; err != nil {
		return model.Peer{}, err
	}
	return peer, nil
}

func chargedTrafficLimit(current *int64, used, addBytes int64) int64 {
	if addBytes < 0 {
		addBytes = 0
	}
	if used < 0 {
		used = 0
	}
	if current != nil && *current > used {
		return *current + addBytes
	}
	return used + addBytes
}

func extendedExpireDate(current *string, addDays int, now time.Time) string {
	if addDays < 0 {
		addDays = 0
	}
	year, month, day := now.Date()
	base := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	if current != nil {
		raw := strings.TrimSpace(*current)
		if parsed, err := time.ParseInLocation("2006-01-02", raw, now.Location()); err == nil && parsed.After(base) {
			base = parsed
		}
	}
	return base.AddDate(0, 0, addDays).Format("2006-01-02")
}
