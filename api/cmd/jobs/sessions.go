package traffic

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
)

const mikrotikPeerCounterMax int64 = 4294967296

type sessionChangeKind int

const (
	sessionNoop sessionChangeKind = iota
	sessionOpen
	sessionClose
	sessionRefresh
)

type sessionObservation struct {
	Online        bool
	HandshakeAge  time.Duration
	Endpoint      string
	EndpointPort  string
	DownloadBytes int64
	UploadBytes   int64
	HasHandshake  bool
}

type sessionChange struct {
	Kind           sessionChangeKind
	ConnectedAt    time.Time
	DisconnectedAt *time.Time
	Endpoint       *string
	EndpointPort   *string
	StartDownload  int64
	StartUpload    int64
	DownloadUsage  int64
	UploadUsage    int64
}

func (c *Calculator) TrackPeerSessions() {
	peers, err := c.fetchPeers()
	if err != nil {
		return
	}
	if len(peers) == 0 {
		c.logger.Debug("No peers found, skipping session tracking")
		return
	}

	wgPeers, err := c.mikrotikAdaptor.FetchWgPeers(context.Background())
	if err != nil {
		c.logger.Error("Failed to fetch WireGuard peers for session tracking", zap.Error(err))
		return
	}

	wgPeerMap := make(map[string]mikrotik.WireGuardPeer, len(wgPeers))
	for _, wgPeer := range wgPeers {
		wgPeerMap[wgPeer.ID] = wgPeer
	}

	var openSessions []model.PeerSession
	if err := c.db.Where("disconnected_at IS NULL").Find(&openSessions).Error; err != nil {
		c.logger.Error("Failed to fetch open peer sessions", zap.Error(err))
		return
	}

	openByPeer := make(map[uint][]model.PeerSession, len(openSessions))
	for _, session := range openSessions {
		openByPeer[session.PeerID] = append(openByPeer[session.PeerID], session)
	}

	now := time.Now()
	opened := 0
	closed := 0

	for _, peer := range peers {
		var mtPeer *mikrotik.WireGuardPeer
		if found, ok := wgPeerMap[peer.PeerID]; ok {
			peerCopy := found
			mtPeer = &peerCopy
		}

		openedCount, closedCount := c.syncPeerSession(peer, mtPeer, openByPeer[peer.ID], now)
		opened += openedCount
		closed += closedCount
	}

	c.logger.Info("Peer session tracking job completed",
		zap.Int("opened", opened),
		zap.Int("closed", closed),
	)
}

func (c *Calculator) syncPeerSession(peer model.Peer, mtPeer *mikrotik.WireGuardPeer, opens []model.PeerSession, now time.Time) (opened, closed int) {
	sort.Slice(opens, func(i, j int) bool {
		return opens[i].ConnectedAt > opens[j].ConnectedAt
	})

	if len(opens) > 1 {
		for i := 1; i < len(opens); i++ {
			if err := c.closeSession(&opens[i], now, sessionObservation{}); err != nil {
				continue
			}
			closed++
		}
		opens = opens[:1]
	}

	var current *model.PeerSession
	if len(opens) > 0 {
		current = &opens[0]
	}

	change := decideSessionChange(observePeerSession(mtPeer), current, now)
	switch change.Kind {
	case sessionOpen:
		if err := c.openSession(peer, change); err != nil {
			return opened, closed
		}
		opened++
	case sessionClose:
		if current == nil {
			return opened, closed
		}
		if err := c.applySessionClose(current, change); err != nil {
			return opened, closed
		}
		closed++
	case sessionRefresh:
		if current != nil {
			_ = c.applySessionRefresh(current, change)
		}
	}

	return opened, closed
}

func observePeerSession(mtPeer *mikrotik.WireGuardPeer) sessionObservation {
	if mtPeer == nil {
		return sessionObservation{}
	}

	age, online, err := utils.HandshakeStatus(mtPeer.Disabled, mtPeer.LastHandshake, common.PeerOnlineHandshakeTimeout)
	if err != nil {
		return sessionObservation{}
	}

	return sessionObservation{
		Online:        online,
		HandshakeAge:  age,
		Endpoint:      strings.TrimSpace(utils.DerefString(mtPeer.CurrentEndpointAddress)),
		EndpointPort:  strings.TrimSpace(utils.DerefString(mtPeer.CurrentEndpointPort)),
		DownloadBytes: utils.ParseStringToInt(mtPeer.TransferTx),
		UploadBytes:   utils.ParseStringToInt(mtPeer.TransferRx),
		HasHandshake:  mtPeer.LastHandshake != nil && strings.TrimSpace(*mtPeer.LastHandshake) != "",
	}
}

func decideSessionChange(obs sessionObservation, open *model.PeerSession, now time.Time) sessionChange {
	if obs.Online {
		if open == nil {
			connectedAt := now.Add(-obs.HandshakeAge)
			if connectedAt.After(now) {
				connectedAt = now
			}
			return sessionChange{
				Kind:          sessionOpen,
				ConnectedAt:   connectedAt,
				Endpoint:      optionalEndpoint(obs.Endpoint),
				EndpointPort:  optionalEndpoint(obs.EndpointPort),
				StartDownload: obs.DownloadBytes,
				StartUpload:   obs.UploadBytes,
			}
		}

		downloadUsage, _ := calculateDelta(open.StartDownload, obs.DownloadBytes, mikrotikPeerCounterMax)
		uploadUsage, _ := calculateDelta(open.StartUpload, obs.UploadBytes, mikrotikPeerCounterMax)
		return sessionChange{
			Kind:          sessionRefresh,
			Endpoint:      optionalEndpoint(obs.Endpoint),
			EndpointPort:  optionalEndpoint(obs.EndpointPort),
			DownloadUsage: downloadUsage,
			UploadUsage:   uploadUsage,
		}
	}

	if open == nil {
		return sessionChange{Kind: sessionNoop}
	}

	disconnectedAt := estimateDisconnectAt(now, obs, open.ConnectedAt)
	downloadUsage, _ := calculateDelta(open.StartDownload, obs.DownloadBytes, mikrotikPeerCounterMax)
	uploadUsage, _ := calculateDelta(open.StartUpload, obs.UploadBytes, mikrotikPeerCounterMax)
	return sessionChange{
		Kind:           sessionClose,
		DisconnectedAt: &disconnectedAt,
		Endpoint:       firstPresent(optionalEndpoint(obs.Endpoint), open.Endpoint),
		EndpointPort:   firstPresent(optionalEndpoint(obs.EndpointPort), open.EndpointPort),
		DownloadUsage:  downloadUsage,
		UploadUsage:    uploadUsage,
	}
}

func estimateDisconnectAt(now time.Time, obs sessionObservation, connectedAt int64) time.Time {
	disconnectedAt := now
	if obs.HasHandshake && obs.HandshakeAge > common.PeerOnlineHandshakeTimeout {
		disconnectedAt = now.Add(-(obs.HandshakeAge - common.PeerOnlineHandshakeTimeout))
	}
	if disconnectedAt.Unix() < connectedAt {
		disconnectedAt = time.Unix(connectedAt, 0)
	}
	if disconnectedAt.After(now) {
		disconnectedAt = now
	}
	return disconnectedAt
}

func (c *Calculator) openSession(peer model.Peer, change sessionChange) error {
	session := model.PeerSession{
		PeerID:        peer.ID,
		PeerUUID:      peer.UUID,
		ConnectedAt:   change.ConnectedAt.Unix(),
		Endpoint:      change.Endpoint,
		EndpointPort:  change.EndpointPort,
		StartDownload: change.StartDownload,
		StartUpload:   change.StartUpload,
	}
	if err := c.db.Create(&session).Error; err != nil {
		c.logger.Error("Failed to open peer session", zap.String("peerID", peer.PeerID), zap.Error(err))
		return err
	}
	return nil
}

func (c *Calculator) applySessionClose(session *model.PeerSession, change sessionChange) error {
	disconnectedAt := time.Now().Unix()
	if change.DisconnectedAt != nil {
		disconnectedAt = change.DisconnectedAt.Unix()
	}

	updates := map[string]interface{}{
		"disconnected_at": disconnectedAt,
		"download_usage":  change.DownloadUsage,
		"upload_usage":    change.UploadUsage,
	}
	if change.Endpoint != nil {
		updates["endpoint"] = *change.Endpoint
	}
	if change.EndpointPort != nil {
		updates["endpoint_port"] = *change.EndpointPort
	}

	if err := c.db.Model(&model.PeerSession{}).Where("id = ?", session.ID).Updates(updates).Error; err != nil {
		c.logger.Error("Failed to close peer session", zap.Uint("sessionID", session.ID), zap.Error(err))
		return err
	}
	return nil
}

func (c *Calculator) applySessionRefresh(session *model.PeerSession, change sessionChange) error {
	updates := map[string]interface{}{}
	if change.DownloadUsage != session.DownloadUsage {
		updates["download_usage"] = change.DownloadUsage
	}
	if change.UploadUsage != session.UploadUsage {
		updates["upload_usage"] = change.UploadUsage
	}
	if change.Endpoint != nil && utils.DerefString(session.Endpoint) != *change.Endpoint {
		updates["endpoint"] = *change.Endpoint
	}
	if change.EndpointPort != nil && utils.DerefString(session.EndpointPort) != *change.EndpointPort {
		updates["endpoint_port"] = *change.EndpointPort
	}
	if len(updates) == 0 {
		return nil
	}

	if err := c.db.Model(&model.PeerSession{}).Where("id = ?", session.ID).Updates(updates).Error; err != nil {
		c.logger.Error("Failed to refresh peer session", zap.Uint("sessionID", session.ID), zap.Error(err))
		return err
	}
	return nil
}

func (c *Calculator) closeSession(session *model.PeerSession, now time.Time, obs sessionObservation) error {
	change := decideSessionChange(obs, session, now)
	if change.Kind != sessionClose {
		disconnectedAt := now
		change = sessionChange{
			Kind:           sessionClose,
			DisconnectedAt: &disconnectedAt,
			DownloadUsage:  session.DownloadUsage,
			UploadUsage:    session.UploadUsage,
			Endpoint:       session.Endpoint,
			EndpointPort:   session.EndpointPort,
		}
	}
	return c.applySessionClose(session, change)
}

func optionalEndpoint(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return utils.Ptr(value)
}

func firstPresent(primary, fallback *string) *string {
	if primary != nil && strings.TrimSpace(*primary) != "" {
		return primary
	}
	return fallback
}
