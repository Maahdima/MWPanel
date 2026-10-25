package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

type ResellerPeerCreate struct {
	Name         string `json:"name"`
	InterfaceID  uint   `json:"interface_id"`
	QuotaBytes   string `json:"quota_bytes"`
	DurationDays string `json:"duration_days"`
}

func (s *PeerPurchases) Create(actor Principal, resellerID uint, key string, req ResellerPeerCreate) (PeerPurchaseView, error) {
	var empty PeerPurchaseView
	if !billingAccess(actor, resellerID) {
		return empty, ErrRoleForbidden
	}
	if !validBillingKey(key) || strings.TrimSpace(req.Name) == "" || len(req.Name) > 128 {
		return empty, ErrInvalidBillingQuantity
	}
	encoded, _ := json.Marshal(req)
	digest := sha256.Sum256(encoded)
	hash := hex.EncodeToString(digest[:])
	var op model.PeerPurchase
	err := s.billing.db.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&op).Error
	if err == nil {
		if op.RequestHash != hash {
			return empty, ErrIdempotencyConflict
		}
		return s.view(op)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return empty, err
	}
	bytes, err := parsePurchaseQuantity(req.QuotaBytes)
	if err != nil || bytes <= 0 {
		return empty, ErrInvalidBillingQuantity
	}
	days, err := parsePurchaseQuantity(req.DurationDays)
	if err != nil || days <= 0 || days > 36500 {
		return empty, ErrInvalidBillingQuantity
	}
	var reseller model.Reseller
	if err = s.billing.db.Where("is_active = ?", true).First(&reseller, resellerID).Error; err != nil {
		return empty, err
	}
	interfaceID := req.InterfaceID
	if interfaceID == 0 && reseller.DefaultInterfaceID != nil {
		interfaceID = *reseller.DefaultInterfaceID
	}
	var assignment model.ResellerResource
	if err = s.billing.db.Where("reseller_id = ? AND interface_id = ?", resellerID, interfaceID).First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return empty, ErrResourceForbidden
		}
		return empty, err
	}
	var iface model.Interface
	if err = s.billing.db.First(&iface, interfaceID).Error; err != nil {
		return empty, err
	}
	var server model.Server
	if err = s.billing.db.Where("is_active = ?", true).First(&server, iface.ServerID).Error; err != nil {
		return empty, err
	}
	// Obtain a candidate from the existing allocator. The database's composite
	// address index resolves any concurrent allocation before router side effects.
	address, err := s.peers.GetNewPeerAllowedAddress(iface.ID)
	if err != nil {
		return empty, err
	}
	if strings.TrimSpace(address.AllowedAddress) == "" {
		return empty, ErrResellerSettings
	}
	credentials, err := s.peers.GetPeerCredentials()
	if err != nil {
		return empty, err
	}
	expiry := s.billing.now().UTC().AddDate(0, 0, int(days)).Format("2006-01-02")
	op = model.PeerPurchase{ID: uuid.NewString(), ResellerID: resellerID, IdempotencyKey: key, RequestHash: hash, ActorID: actor.AccountID, Kind: string(BillingCreate), State: "pending", QuotaBytes: bytes, ExpireTime: expiry}
	peer := model.Peer{ResellerID: &resellerID, ServerID: iface.ServerID, PurchaseID: &op.ID, ConfiguredDurationDays: days, ProvisioningBlocked: true, RestrictionsInitialized: true, Disabled: true, UUID: uuid.NewString(), PeerID: "pending:" + op.ID, Name: strings.TrimSpace(req.Name), PrivateKey: credentials.PrivateKey, PublicKey: credentials.PublicKey, Interface: iface.Name, AllowedAddress: address.AllowedAddress, Endpoint: server.IPAddress, EndpointPort: iface.ListenPort, PersistentKeepalive: common.DefaultKeepalive, TrafficLimit: &bytes, ExpireTime: &expiry}
	_, err = s.billing.reserveCharge(actor, resellerID, key, nil, BillingCreate, bytes, days, func(tx *gorm.DB, charge model.BillingCharge) error {
		// Re-read assignments under the account lock; a simultaneous default change
		// cannot silently reroute this accepted operation.
		var current model.Interface
		if err := tx.First(&current, iface.ID).Error; err != nil {
			return err
		}
		if current.ServerID != iface.ServerID || current.Name != iface.Name || current.PublicKey != iface.PublicKey {
			return ErrPendingPeerOperation
		}
		lock := tx.Model(&model.Interface{}).Where("id = ?", iface.ID).UpdateColumn("updated_at", gorm.Expr("updated_at"))
		if lock.Error != nil {
			return lock.Error
		}
		if lock.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}
		if err := s.validateAddress(tx, peer, peer.AllowedAddress); err != nil {
			return err
		}
		if err := tx.Create(&peer).Error; err != nil {
			return err
		}
		candidate := peer
		candidate.Disabled = false
		if err := ValidatePeerResource(tx, resellerID, candidate, peer.ID); err != nil {
			return err
		}
		op.PeerID = peer.ID
		op.ChargeID = &charge.ID
		return tx.Create(&op).Error
	})
	if err != nil {
		return empty, err
	}
	op = model.PeerPurchase{}
	if err = s.billing.db.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&op).Error; err != nil {
		return empty, err
	}
	if op.RequestHash != hash {
		return empty, ErrIdempotencyConflict
	}
	return s.view(op)
}

// applyCreate discovers objects by the persisted public key and stable UUID
// names before each create. Ambiguous timeouts are resolved by inventory on the
// next attempt. Destination peers stay disabled until the database and ledger
// commit together and the restriction reconciler authorizes activation.
func (s *PeerPurchases) applyCreate(op model.PeerPurchase, peer model.Peer) error {
	router, err := s.peers.mikrotikAdaptor.ForServer(peer.ServerID)
	if err != nil {
		return err
	}
	ctx := context.Background()
	iface, err := s.peers.getInterfaceByName(peer.ServerID, peer.Interface)
	if err != nil {
		return err
	}
	observedInterface, err := router.FetchWgInterface(ctx, iface.InterfaceID)
	if err != nil {
		return err
	}
	if observedInterface.Name != iface.Name || observedInterface.PublicKey != iface.PublicKey {
		return ErrRouterPeerIdentity
	}
	inventory, err := router.FetchWgPeers(ctx)
	if err != nil {
		return err
	}
	var remote *mikrotik.WireGuardPeer
	for _, candidate := range inventory {
		if candidate.PublicKey == peer.PublicKey {
			if remote != nil {
				return ErrRouterPeerIdentity
			}
			copyPeer := candidate
			remote = &copyPeer
		}
	}
	if remote == nil {
		// Reject an addressing conflict rather than touching an unrelated peer.
		for _, candidate := range inventory {
			if candidate.AllowedAddress == peer.AllowedAddress {
				return ErrRouterPeerIdentity
			}
		}
		remote, err = router.CreateWgPeer(ctx, mikrotik.WireGuardPeer{Name: peer.Name, PublicKey: peer.PublicKey, PrivateKey: &peer.PrivateKey, Interface: peer.Interface, AllowedAddress: peer.AllowedAddress, PersistentKeepAlive: &peer.PersistentKeepalive, Disabled: "true"})
		if err != nil {
			return err
		}
	}
	verified, err := router.FetchWgPeer(ctx, remote.ID)
	if err != nil {
		return err
	}
	if verified.PublicKey != peer.PublicKey || verified.Interface != peer.Interface || verified.AllowedAddress != peer.AllowedAddress {
		return ErrRouterPeerIdentity
	}
	if verified.Disabled != "true" {
		if _, err = router.UpdateWgPeer(ctx, remote.ID, mikrotik.WireGuardPeer{Disabled: "true"}); err != nil {
			return err
		}
	}
	queues, err := router.FetchSimpleQueues(ctx)
	if err != nil {
		return err
	}
	queueName := common.QueueName + peer.UUID
	var queueID *string
	for _, q := range queues {
		if q.Name == queueName {
			if queueID != nil || q.Target == nil || *q.Target != peer.AllowedAddress {
				return ErrRouterPeerIdentity
			}
			id := q.ID
			queueID = &id
		}
	}
	bound, err := s.peers.ForServer(peer.ServerID)
	if err != nil {
		return err
	}
	if queueID == nil {
		queueID, err = bound.queue.createQueue(peer.UUID, peer.AllowedAddress, peer.DownloadBandwidth, peer.UploadBandwidth)
		if err != nil {
			return err
		}
	}
	queue, err := router.FetchSimpleQueue(ctx, *queueID)
	if err != nil {
		return err
	}
	if queue.Name != queueName || queue.Target == nil || *queue.Target != peer.AllowedAddress {
		return ErrRouterPeerIdentity
	}
	schedules, err := router.FetchSchedulers(ctx)
	if err != nil {
		return err
	}
	scheduleName := common.SchedulerName + peer.UUID
	var schedulerID *string
	for _, schedule := range schedules {
		if schedule.Name == scheduleName {
			if schedulerID != nil || schedule.OnEvent == nil || *schedule.OnEvent != schedulerOnEvent(remote.ID) {
				return ErrRouterPeerIdentity
			}
			id := schedule.ID
			schedulerID = &id
		}
	}
	if schedulerID == nil {
		schedulerID, err = bound.scheduler.createScheduler(remote.ID, peer.UUID, &op.ExpireTime)
		if err != nil {
			return err
		}
	}
	scheduler, err := router.FetchScheduler(ctx, *schedulerID)
	if err != nil {
		return err
	}
	date, err := toMikrotikDate(&op.ExpireTime)
	if err != nil {
		return err
	}
	if scheduler.OnEvent == nil || *scheduler.OnEvent != schedulerOnEvent(remote.ID) || scheduler.StartDate == nil || *scheduler.StartDate != *date {
		return ErrRouterPeerIdentity
	}
	final := peer
	final.PeerID = remote.ID
	final.QueueID = queueID
	final.SchedulerID = schedulerID
	if err = bound.generatePeerAssets(peer.PrivateKey, final, iface.PublicKey); err != nil {
		return err
	}
	actor := Principal{AccountID: op.ActorID, Role: model.RoleReseller, ResellerID: &op.ResellerID}
	return s.billing.CommitCharge(actor, op.ResellerID, op.IdempotencyKey, func(tx *gorm.DB) error {
		result := tx.Model(&model.Peer{}).Where("id = ? AND reseller_id = ? AND purchase_id = ?", peer.ID, op.ResellerID, op.ID).Updates(map[string]any{"peer_id": remote.ID, "queue_id": queueID, "scheduler_id": schedulerID, "provisioning_blocked": false, "purchase_id": nil, "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}
		if err := tx.Model(&model.BillingCharge{}).Where("id = ?", *op.ChargeID).Update("peer_id", peer.ID).Error; err != nil {
			return err
		}
		return tx.Model(&model.PeerPurchase{}).Where("id = ? AND state = ?", op.ID, "pending").Updates(map[string]any{"state": "applied", "lease_until": 0, "last_error": ""}).Error
	})
}

func (s *Resellers) CreatePeer(actor Principal, resellerID uint, key string, req ResellerPeerCreate) (PeerPurchaseView, error) {
	return NewPeerPurchases(NewBilling(s.db), s.peers).Create(actor, resellerID, key, req)
}

func (s *Resellers) MyResources(actor Principal) ([]ResellerResourceView, error) {
	id, err := myReseller(actor)
	if err != nil {
		return nil, err
	}
	var resources []model.ResellerResource
	if err = s.db.Preload("Interface.IPPool").Where("reseller_id = ?", id).Order("interface_id ASC").Find(&resources).Error; err != nil {
		return nil, err
	}
	result := make([]ResellerResourceView, 0, len(resources))
	for _, resource := range resources {
		row := ResellerResourceView{InterfaceID: resource.InterfaceID, Name: resource.Interface.Name, ServerID: resource.Interface.ServerID}
		if resource.Interface.IPPool != nil {
			row.PoolID = &resource.Interface.IPPool.ID
		}
		result = append(result, row)
	}
	return result, nil
}
