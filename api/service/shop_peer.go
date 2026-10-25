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
	"github.com/maahdima/mwp/api/utils"
)

func (w *WgPeer) ShopInterface(interfaceID uint) (model.Interface, error) {
	var iface model.Interface
	query := w.db.Preload("IPPool")
	if interfaceID == 0 {
		serverID, err := resolveServerID(w.db, w.serverID)
		if err != nil {
			return iface, err
		}
		query = query.Where("server_id = ?", serverID)
	}
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
	downloadBandwidth := "16M"
	uploadBandwidth := "16M"
	req := &schema.CreatePeerRequest{
		Comment:           &comment,
		Name:              strings.TrimSpace(name),
		InterfaceId:       iface.ID,
		PrivateKey:        creds.PrivateKey,
		PublicKey:         creds.PublicKey,
		AllowedAddress:    address.AllowedAddress,
		Endpoint:          strings.TrimSpace(endpoint),
		ExpireTime:        &expire,
		TrafficLimit:      &traffic,
		DownloadBandwidth: &downloadBandwidth,
		UploadBandwidth:   &uploadBandwidth,
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
	if peer.ResellerID != nil {
		return model.Peer{}, ErrBilledOperationRequired
	}

	bound, err := w.ForServer(peer.ServerID)
	if err != nil {
		return model.Peer{}, err
	}
	w = bound

	if addBytes <= 0 || addDays <= 0 {
		return model.Peer{}, errors.New("invalid vpn charge")
	}

	now := time.Now()
	expired := utils.IsPeerExpired(peer.ExpireTime, now)
	used := peer.DownloadUsage + peer.UploadUsage
	newLimit := renewedTrafficLimit(peer.TrafficLimit, used, addBytes, expired)
	expire := renewalExpireDate(addDays, now)
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
	if expired {
		updates["download_usage"] = int64(0)
		updates["upload_usage"] = int64(0)
		if w.mikrotikAdaptor != nil {
			if wgPeer, err := w.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID); err != nil {
				w.logger.Warn("failed to read peer counters before traffic reset", zap.Error(err), zap.String("peerID", peer.PeerID))
			} else {
				updates["last_tx"] = utils.ParseStringToInt(wgPeer.TransferTx)
				updates["last_rx"] = utils.ParseStringToInt(wgPeer.TransferRx)
			}
		}
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

func renewedTrafficLimit(current *int64, used, addBytes int64, expired bool) int64 {
	if expired {
		if addBytes < 0 {
			return 0
		}
		return addBytes
	}
	return chargedTrafficLimit(current, used, addBytes)
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

// renewalExpireDate counts the purchased duration from the purchase day.
// It does not extend a previous expire date.
func renewalExpireDate(addDays int, purchasedAt time.Time) string {
	if addDays < 0 {
		addDays = 0
	}
	year, month, day := purchasedAt.Date()
	base := time.Date(year, month, day, 0, 0, 0, 0, purchasedAt.Location())
	return base.AddDate(0, 0, addDays).Format("2006-01-02")
}
