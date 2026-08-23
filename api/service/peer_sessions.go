package service

import (
	"fmt"
	"net"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const peerSessionHistoryLimit = 100

func (w *WgPeer) GetPeerSessions(id uint) (*schema.PeerSessionsResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return nil, err
	}

	var sessions []model.PeerSession
	if err := w.db.
		Where("peer_id = ?", peer.ID).
		Order("connected_at DESC").
		Limit(peerSessionHistoryLimit).
		Find(&sessions).Error; err != nil {
		w.logger.Error("failed to fetch peer sessions", zap.Uint("peerID", peer.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch peer sessions: %w", err)
	}

	var totals []model.PeerSession
	if err := w.db.
		Select("connected_at", "disconnected_at", "download_usage", "upload_usage").
		Where("peer_id = ?", peer.ID).
		Find(&totals).Error; err != nil {
		w.logger.Error("failed to fetch peer session totals", zap.Uint("peerID", peer.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch peer session totals: %w", err)
	}

	now := time.Now()
	ongoing := 0
	var totalDuration time.Duration
	var totalDownload, totalUpload int64
	for _, session := range totals {
		if session.DisconnectedAt == nil {
			ongoing++
		}
		totalDuration += sessionDuration(session.ConnectedAt, unixOrNow(session.DisconnectedAt, now))
		totalDownload += session.DownloadUsage
		totalUpload += session.UploadUsage
	}

	items := make([]schema.PeerSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, transformPeerSession(session, now))
	}

	return &schema.PeerSessionsResponse{
		PeerID:             peer.ID,
		PeerName:           peer.Name,
		TotalSessions:      len(totals),
		OngoingSessions:    ongoing,
		TotalDuration:      int64(totalDuration.Seconds()),
		TotalDurationLabel: utils.FormatPrettyDuration(totalDuration),
		TotalDownloadUsage: utils.FormatDataSize(totalDownload),
		TotalUploadUsage:   utils.FormatDataSize(totalUpload),
		TotalUsage:         utils.FormatDataSize(totalDownload + totalUpload),
		Sessions:           items,
	}, nil
}

func transformPeerSession(session model.PeerSession, now time.Time) schema.PeerSessionResponse {
	status := schema.PeerSessionEnded
	var disconnectedAt *string
	endedAt := now
	if session.DisconnectedAt != nil {
		formatted := formatUnixTime(*session.DisconnectedAt)
		disconnectedAt = &formatted
		endedAt = time.Unix(*session.DisconnectedAt, 0)
	} else {
		status = schema.PeerSessionOngoing
	}

	duration := sessionDuration(session.ConnectedAt, endedAt.Unix())
	return schema.PeerSessionResponse{
		ID:             session.ID,
		Status:         status,
		ConnectedAt:    formatUnixTime(session.ConnectedAt),
		DisconnectedAt: disconnectedAt,
		Duration:       int64(duration.Seconds()),
		DurationLabel:  utils.FormatPrettyDuration(duration),
		Endpoint:       formatSessionEndpoint(session.Endpoint, session.EndpointPort),
		DownloadUsage:  utils.FormatDataSize(session.DownloadUsage),
		UploadUsage:    utils.FormatDataSize(session.UploadUsage),
		TotalUsage:     utils.FormatDataSize(session.DownloadUsage + session.UploadUsage),
	}
}

func sessionDuration(connectedAt, endedAt int64) time.Duration {
	if endedAt < connectedAt {
		return 0
	}
	return time.Duration(endedAt-connectedAt) * time.Second
}

func unixOrNow(value *int64, now time.Time) int64 {
	if value == nil {
		return now.Unix()
	}
	return *value
}

func formatUnixTime(value int64) string {
	return time.Unix(value, 0).UTC().Format(time.RFC3339)
}

func formatSessionEndpoint(address, port *string) *string {
	host := utils.DerefString(address)
	if host == "" {
		return nil
	}
	if portValue := utils.DerefString(port); portValue != "" {
		return utils.Ptr(net.JoinHostPort(host, portValue))
	}
	return utils.Ptr(host)
}

func deletePeerSessions(db *gorm.DB, peerID uint) error {
	return db.Unscoped().Where("peer_id = ?", peerID).Delete(&model.PeerSession{}).Error
}
