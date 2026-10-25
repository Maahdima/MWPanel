package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils/httphelper"
	"gorm.io/gorm"
)

type PeerMutationRequest struct {
	Action  string `json:"action"`
	Value   string `json:"value"`
	Enabled *bool  `json:"enabled"`
}
type peerMutationPayload struct {
	Request    PeerMutationRequest
	PrivateKey string
	PublicKey  string
}

func (s *PeerPurchases) Mutate(actor Principal, resellerID, peerID uint, key string, req PeerMutationRequest) (PeerPurchaseView, error) {
	var empty PeerPurchaseView
	if !billingAccess(actor, resellerID) {
		return empty, ErrRoleForbidden
	}
	if actor.Role == model.RoleReseller {
		state, err := s.billing.State(actor, resellerID)
		if err != nil {
			return empty, err
		}
		if state.BillingSuspended || state.AdministrativeSuspended {
			return empty, ErrPeerRestricted
		}
	}
	if !validBillingKey(key) {
		return empty, ErrInvalidBillingQuantity
	}
	encoded, _ := json.Marshal(struct {
		PeerID  uint
		Request PeerMutationRequest
	}{peerID, req})
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
	var peer model.Peer
	if err = s.billing.db.Where("reseller_id = ?", resellerID).First(&peer, peerID).Error; err != nil {
		return empty, err
	}
	if peer.PurchaseID != nil || peer.ProvisioningBlocked {
		return empty, ErrPendingPeerOperation
	}
	if peer.TrafficLimit == nil || peer.ExpireTime == nil {
		return empty, ErrInvalidBillingQuantity
	}
	payload := peerMutationPayload{Request: req}
	quota := *peer.TrafficLimit
	switch req.Action {
	case "rename":
		if strings.TrimSpace(req.Value) == "" || len(req.Value) > 128 {
			return empty, ErrInvalidBillingQuantity
		}
	case "set_manual", "set_share":
		if req.Enabled == nil || req.Value != "" {
			return empty, ErrInvalidBillingQuantity
		}
	case "regenerate_keys":
		if req.Value != "" || req.Enabled != nil {
			return empty, ErrInvalidBillingQuantity
		}
		keys, err := s.peers.GetPeerCredentials()
		if err != nil {
			return empty, err
		}
		payload.PrivateKey = keys.PrivateKey
		payload.PublicKey = keys.PublicKey
	case "set_address":
		if err = s.validateAddress(s.billing.db, peer, req.Value); err != nil {
			return empty, err
		}
	case "reduce_quota":
		quota, err = parsePurchaseQuantity(req.Value)
		if err != nil || quota <= 0 || quota > *peer.TrafficLimit {
			return empty, ErrBilledOperationRequired
		}
	case "delete":
		if req.Value != "" || req.Enabled != nil {
			return empty, ErrInvalidBillingQuantity
		}
	default:
		return empty, ErrInvalidBillingQuantity
	}
	body, _ := json.Marshal(payload)
	op = model.PeerPurchase{ID: uuid.NewString(), ResellerID: resellerID, IdempotencyKey: key, RequestHash: hash, PeerID: peerID, ActorID: actor.AccountID, Kind: req.Action, Payload: string(body), QuotaBytes: quota, ExpireTime: *peer.ExpireTime, State: "pending"}
	err = s.billing.accountTransaction(resellerID, func(tx *gorm.DB, account *model.BillingAccount) error {
		if actor.Role == model.RoleReseller {
			state, err := s.billing.accountState(tx, account, s.billing.now().Unix())
			if err != nil {
				return err
			}
			if state.BillingSuspended || state.AdministrativeSuspended {
				return ErrPeerRestricted
			}
		}

		var existing model.PeerPurchase
		if e := tx.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&existing).Error; e == nil {
			if existing.RequestHash != hash {
				return ErrIdempotencyConflict
			}
			op = existing
			return nil
		} else if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if req.Action == "set_address" {
			lock := tx.Model(&model.Interface{}).Where("server_id = ? AND name = ?", peer.ServerID, peer.Interface).UpdateColumn("updated_at", gorm.Expr("updated_at"))
			if lock.Error != nil {
				return lock.Error
			}
			if lock.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		lock := tx.Model(&model.Peer{}).Where("id = ? AND reseller_id = ? AND purchase_id IS NULL AND restriction_revision = ? AND sync_lease_until <= ?", peerID, resellerID, peer.RestrictionRevision, s.billing.now().Unix()).UpdateColumn("purchase_id", op.ID)
		if lock.Error != nil {
			return lock.Error
		}
		if lock.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}
		if req.Action == "set_address" {
			if e := s.validateAddress(tx, peer, req.Value); e != nil {
				return e
			}
		}
		if req.Action == "set_manual" && *req.Enabled {
			blocked, e := peerOtherRestriction(tx, peer, s.billing.now())
			if e != nil {
				return e
			}
			if blocked {
				return ErrPeerRestricted
			}
			candidate := peer
			candidate.Disabled = false
			if e = ValidatePeerResource(tx, resellerID, candidate, peerID); e != nil {
				return e
			}
		}
		if err := tx.Create(&op).Error; err != nil {
			return err
		}
		return billingAudit(tx, resellerID, actor.AccountID, nil, "peer_mutation_queued", op.ID+":"+req.Action)
	})
	if err != nil {
		return empty, err
	}
	return s.view(op)
}

func (s *PeerPurchases) validateAddress(tx *gorm.DB, peer model.Peer, address string) error {
	ip, network, err := net.ParseCIDR(address)
	if err != nil || ip.To4() == nil {
		return ErrInvalidBillingQuantity
	}
	ones, bits := network.Mask.Size()
	if ones != 32 || bits != 32 || address != ip.String()+"/32" {
		return ErrInvalidBillingQuantity
	}
	var iface model.Interface
	if err = tx.Preload("IPPool").Where("server_id = ? AND name = ?", peer.ServerID, peer.Interface).First(&iface).Error; err != nil {
		return err
	}
	if iface.IPPool == nil {
		return ErrResourceForbidden
	}
	start := net.ParseIP(strings.Split(iface.IPPool.StartIP, "/")[0]).To4()
	end := net.ParseIP(strings.Split(iface.IPPool.EndIP, "/")[0]).To4()
	if start == nil || end == nil || bytes.Compare(ip.To4(), start) < 0 || bytes.Compare(ip.To4(), end) > 0 {
		return ErrResourceForbidden
	}
	var count int64
	if err = tx.Model(&model.Peer{}).Where("server_id = ? AND id <> ? AND allowed_address = ?", peer.ServerID, peer.ID, address).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrResellerLimit
	}
	var pending []model.PeerPurchase
	if err = tx.Joins("JOIN peers ON peers.id = peer_purchases.peer_id").Where("peers.server_id = ? AND peer_purchases.kind = ? AND peer_purchases.state = ? AND peer_purchases.peer_id <> ?", peer.ServerID, "set_address", "pending", peer.ID).Find(&pending).Error; err != nil {
		return err
	}
	for _, other := range pending {
		var p peerMutationPayload
		if json.Unmarshal([]byte(other.Payload), &p) != nil {
			return ErrPendingPeerOperation
		}
		if p.Request.Value == address {
			return ErrResellerLimit
		}
	}
	return nil
}

func (s *PeerPurchases) applyFree(op model.PeerPurchase, peer model.Peer) error {
	var payload peerMutationPayload
	if err := json.Unmarshal([]byte(op.Payload), &payload); err != nil {
		return err
	}
	req := payload.Request
	if req.Action != op.Kind {
		return ErrIdempotencyConflict
	}
	router, err := s.peers.mikrotikAdaptor.ForServer(peer.ServerID)
	if err != nil {
		return err
	}
	ctx := context.Background()
	remote, err := router.FetchWgPeer(ctx, peer.PeerID)
	missing := httphelper.IsNotFound(err)
	if err != nil && !(op.Kind == "delete" && missing) {
		return err
	}
	if !missing && (remote.Interface != peer.Interface || (remote.PublicKey != peer.PublicKey && !(op.Kind == "regenerate_keys" && remote.PublicKey == payload.PublicKey))) {
		return ErrRouterPeerIdentity
	}
	final := peer
	updates := map[string]any{"purchase_id": nil, "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0}
	switch op.Kind {
	case "rename":
		final.Name = strings.TrimSpace(req.Value)
		if _, err = router.UpdateWgPeer(ctx, peer.PeerID, mikrotik.WireGuardPeer{Name: final.Name}); err != nil {
			return err
		}
		updates["name"] = final.Name
	case "regenerate_keys":
		final.PublicKey = payload.PublicKey
		final.PrivateKey = payload.PrivateKey
		if _, err = router.UpdateWgPeer(ctx, peer.PeerID, mikrotik.WireGuardPeer{PublicKey: final.PublicKey, PrivateKey: &final.PrivateKey}); err != nil {
			return err
		}
		updates["public_key"] = final.PublicKey
		updates["private_key"] = final.PrivateKey
	case "set_address":
		if err = s.validateAddress(s.billing.db, peer, req.Value); err != nil {
			return err
		}
		bound, e := s.peers.ForServer(peer.ServerID)
		if e != nil {
			return e
		}
		addressIP := net.ParseIP(strings.Split(req.Value, "/")[0])
		for _, reserved := range bound.interfaceReservedIPs(peer.Interface) {
			if reserved.Equal(addressIP) {
				return ErrResourceForbidden
			}
		}
		inventory, e := router.FetchWgPeers(ctx)
		if e != nil {
			return e
		}
		for _, other := range inventory {
			if other.ID != peer.PeerID && other.AllowedAddress == req.Value {
				return ErrResourceForbidden
			}
		}
		if peer.QueueID != nil {
			queue, e := router.FetchSimpleQueue(ctx, *peer.QueueID)
			if e != nil {
				return e
			}
			if queue.Target == nil || (*queue.Target != peer.AllowedAddress && *queue.Target != req.Value) {
				return ErrRouterPeerIdentity
			}
			if _, e = router.UpdateSimpleQueue(ctx, *peer.QueueID, mikrotik.Queue{Target: &req.Value}); e != nil {
				return e
			}
		}
		if _, err = router.UpdateWgPeer(ctx, peer.PeerID, mikrotik.WireGuardPeer{AllowedAddress: req.Value}); err != nil {
			return err
		}
		final.AllowedAddress = req.Value
		updates["allowed_address"] = req.Value
	case "set_manual":
		updates["manual_disabled"] = !*req.Enabled
		updates["disabled"] = gorm.Expr("provisioning_blocked OR expiry_blocked OR quota_blocked OR billing_blocked OR admin_blocked OR ?", !*req.Enabled)
	case "set_share":
		updates["is_shared"] = *req.Enabled
	case "reduce_quota":
		updates["traffic_limit"] = op.QuotaBytes
	case "delete":
		// Disable first; remove only objects whose stored identities still match.
		if !missing {
			if _, err = router.UpdateWgPeer(ctx, peer.PeerID, mikrotik.WireGuardPeer{Disabled: "true"}); err != nil {
				return err
			}
		}
		if peer.SchedulerID != nil {
			scheduler, e := router.FetchScheduler(ctx, *peer.SchedulerID)
			if e == nil {
				if scheduler.OnEvent == nil || *scheduler.OnEvent != schedulerOnEvent(peer.PeerID) {
					return ErrRouterPeerIdentity
				}
				if e = router.DeleteScheduler(ctx, *peer.SchedulerID); e != nil && !httphelper.IsNotFound(e) {
					return e
				}
			} else if !httphelper.IsNotFound(e) {
				return e
			}
		}
		if peer.QueueID != nil {
			queue, e := router.FetchSimpleQueue(ctx, *peer.QueueID)
			if e == nil {
				if queue.Target == nil || *queue.Target != peer.AllowedAddress {
					return ErrRouterPeerIdentity
				}
				if e = router.DeleteSimpleQueue(ctx, *peer.QueueID); e != nil && !httphelper.IsNotFound(e) {
					return e
				}
			} else if !httphelper.IsNotFound(e) {
				return e
			}
		}
		if !missing {
			if err = router.DeleteWgPeer(ctx, peer.PeerID); err != nil && !httphelper.IsNotFound(err) {
				return err
			}
		}
		if _, e := router.FetchWgPeer(ctx, peer.PeerID); !httphelper.IsNotFound(e) {
			if e != nil {
				return e
			}
			return ErrRouterSyncPending
		}
		if err = s.peers.configGenerator.RemovePeerConfigByUUID(peer.UUID); err != nil {
			return err
		}
		if err = s.peers.qrCodeGenerator.RemovePeerQRCodeByUUID(peer.UUID); err != nil {
			return err
		}
	default:
		return ErrInvalidBillingQuantity
	}
	if op.Kind != "delete" {
		verified, e := router.FetchWgPeer(ctx, peer.PeerID)
		if e != nil {
			return e
		}
		if verified.PublicKey != final.PublicKey || verified.Interface != final.Interface || verified.AllowedAddress != final.AllowedAddress || (op.Kind == "rename" && verified.Name != final.Name) {
			return ErrRouterPeerIdentity
		}
		if op.Kind == "rename" || op.Kind == "regenerate_keys" || op.Kind == "set_address" {
			iface, e := s.peers.getInterfaceByName(peer.ServerID, peer.Interface)
			if e != nil {
				return e
			}
			if e = s.peers.generatePeerAssets(final.PrivateKey, final, iface.PublicKey); e != nil {
				return e
			}
		}
	}
	return s.billing.accountTransaction(op.ResellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var current model.Peer
		if err := tx.Where("reseller_id = ? AND purchase_id = ?", op.ResellerID, op.ID).First(&current, peer.ID).Error; err != nil {
			return err
		}
		if op.Kind == "delete" {
			if err := s.peers.deletePeerSessions(tx, peer.ID); err != nil {
				return err
			}
			if err := tx.Where("peer_uuid = ?", peer.UUID).Delete(&model.TelegramChat{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Delete(&current).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&current).Updates(updates).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.PeerPurchase{}).Where("id = ? AND state = ?", op.ID, "pending").Updates(map[string]any{"state": "applied", "lease_until": 0, "last_error": ""}).Error; err != nil {
			return err
		}
		return billingAudit(tx, op.ResellerID, op.ActorID, nil, "peer_mutation_applied", op.ID+":"+op.Kind)
	})
}

func (s *Resellers) MutatePeer(actor Principal, resellerID, peerID uint, key string, req PeerMutationRequest) (PeerPurchaseView, error) {
	return NewPeerPurchases(NewBilling(s.db), s.peers).Mutate(actor, resellerID, peerID, key, req)
}
