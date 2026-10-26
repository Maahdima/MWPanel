package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"
	"github.com/maahdima/mwp/api/utils/httphelper"
	"gorm.io/gorm"
)

var ErrMigrationConflict = errors.New("migration conflicts with current router resources or the current migration phase")

type RouterMigrations struct {
	billing *Billing
	peers   *WgPeer
	runMu   sync.Mutex
}

func NewRouterMigrations(billing *Billing, peers *WgPeer) *RouterMigrations {
	return &RouterMigrations{billing: billing, peers: peers}
}

type RouterMigrationRequest struct {
	InterfaceID uint   `json:"interface_id"`
	Endpoint    string `json:"endpoint"`
	Reason      string `json:"reason"`
}

type RouterMigrationView struct {
	ID                     string `json:"id"`
	State                  string `json:"state"`
	DestinationInterfaceID uint   `json:"destination_interface_id"`
	TotalPeers             int64  `json:"total_peers"`
	CompletedPeers         int64  `json:"completed_peers"`
	Attempts               int64  `json:"attempts"`
	NeedsAttention         bool   `json:"needs_attention"`
}

func (s *RouterMigrations) view(job model.RouterMigration) (RouterMigrationView, error) {
	v := RouterMigrationView{ID: job.ID, State: job.State, DestinationInterfaceID: job.DestinationInterfaceID, Attempts: job.Attempts, NeedsAttention: job.LastError != ""}

	if err := s.billing.db.Model(&model.RouterMigrationPeer{}).Where("migration_id = ?", job.ID).Count(&v.TotalPeers).Error; err != nil {
		return v, err
	}

	err := s.billing.db.Model(&model.RouterMigrationPeer{}).Where("migration_id = ? AND phase IN ?", job.ID, []string{"complete", "cancelled"}).Count(&v.CompletedPeers).Error

	return v, err
}

// Begin performs read-only preflight before taking any durable locks. It then
// checks the inventory again under the account lock and records every affected
// peer, source state and destination address before any external mutation.
func (s *RouterMigrations) Begin(actor Principal, resellerID uint, key string, req RouterMigrationRequest) (RouterMigrationView, error) {
	var empty RouterMigrationView

	if !billingAdmin(actor) {
		return empty, ErrRoleForbidden
	}

	if !validBillingKey(key) || req.InterfaceID == 0 || strings.TrimSpace(req.Reason) == "" || len(req.Reason) > 1000 {
		return empty, ErrResellerSettings
	}

	payload, _ := json.Marshal(req)

	digest := sha256.Sum256(payload)

	hash := hex.EncodeToString(digest[:])

	var existing model.RouterMigration

	err := s.billing.db.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&existing).Error

	if err == nil {
		if existing.RequestHash != hash {
			return empty, ErrIdempotencyConflict
		}

		return s.view(existing)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return empty, err
	}

	var assignment model.ResellerResource

	if err = s.billing.db.Where("reseller_id = ? AND interface_id = ?", resellerID, req.InterfaceID).First(&assignment).Error; err != nil {
		return empty, ErrResourceForbidden
	}

	var reseller model.Reseller

	if err = s.billing.db.First(&reseller, resellerID).Error; err != nil {
		return empty, err
	}

	if reseller.MigrationID != nil {
		return empty, ErrPendingPeerOperation
	}

	var iface model.Interface

	if err = s.billing.db.Preload("IPPool").First(&iface, req.InterfaceID).Error; err != nil {
		return empty, err
	}

	var server model.Server

	if err = s.billing.db.Where("is_active = ?", true).First(&server, iface.ServerID).Error; err != nil {
		return empty, err
	}

	endpoint := req.Endpoint

	if endpoint == "" {
		endpoint = server.IPAddress
	}

	if !validMigrationEndpoint(endpoint) {
		return empty, ErrResellerSettings
	}

	if ip := net.ParseIP(strings.Trim(endpoint, "[]")); ip != nil && ip.To4() == nil {
		endpoint = "[" + ip.String() + "]"
	}

	if iface.Disabled || iface.IPPool == nil {
		return empty, ErrMigrationConflict
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	router, err := s.peers.mikrotikAdaptor.ForServer(iface.ServerID)

	if err != nil {
		return empty, err
	}

	wg, err := router.FetchWgInterface(ctx, iface.InterfaceID)

	if err != nil {
		return empty, err
	}

	if wg.Name != iface.Name || wg.PublicKey != iface.PublicKey || wg.Disabled == "true" {
		return empty, ErrMigrationConflict
	}

	remote, err := router.FetchWgPeers(ctx)

	if err != nil {
		return empty, err
	}

	addresses, err := router.FetchIPv4Addresses(ctx)

	if err != nil {
		return empty, err
	}

	used := []net.IP{}

	for _, p := range remote {
		if ip := net.ParseIP(strings.Split(p.AllowedAddress, "/")[0]).To4(); ip != nil {
			used = append(used, ip)
		}
	}

	for _, address := range addresses {
		if ip := net.ParseIP(strings.Split(address.Address, "/")[0]).To4(); ip != nil {
			used = append(used, ip)
		}
	}

	var destinationPeers []model.Peer

	if err = s.billing.db.Where("server_id = ?", iface.ServerID).Find(&destinationPeers).Error; err != nil {
		return empty, err
	}

	for _, p := range destinationPeers {
		if ip := net.ParseIP(strings.Split(p.AllowedAddress, "/")[0]).To4(); ip != nil {
			used = append(used, ip)
		}
	}

	var owned []model.Peer

	if err = s.billing.db.Where("reseller_id = ?", resellerID).Order("id ASC").Find(&owned).Error; err != nil {
		return empty, err
	}

	job := model.RouterMigration{ID: uuid.NewString(), ResellerID: resellerID, IdempotencyKey: key, RequestHash: hash, DestinationInterfaceID: iface.ID, ActorID: actor.AccountID, Reason: req.Reason, State: "pending"}

	rows := make([]model.RouterMigrationPeer, 0, len(owned))

	start := net.ParseIP(strings.Split(iface.IPPool.StartIP, "/")[0]).To4()

	end := net.ParseIP(strings.Split(iface.IPPool.EndIP, "/")[0]).To4()

	if start == nil || end == nil {
		return empty, ErrMigrationConflict
	}

	seenKeys := map[string]bool{}

	for _, peer := range owned {
		if peer.PurchaseID != nil || peer.MigrationID != nil || peer.SyncLeaseUntil > s.billing.now().Unix() {
			return empty, ErrPendingPeerOperation
		}

		if peer.ServerID == iface.ServerID && peer.Interface == iface.Name {
			continue
		}

		// A router-local interface move needs separate key/address conflict handling;
		// refusing it is safer than treating the existing source as a destination.
		if peer.ServerID == iface.ServerID {
			return empty, ErrMigrationConflict
		}

		if !validWireGuardKey(peer.PublicKey) || !validWireGuardKey(peer.PrivateKey) || seenKeys[peer.PublicKey] {
			return empty, ErrMigrationConflict
		}

		seenKeys[peer.PublicKey] = true

		for _, other := range remote {
			if other.PublicKey == peer.PublicKey {
				return empty, ErrMigrationConflict
			}
		}

		sourceRouter, e := s.peers.mikrotikAdaptor.ForServer(peer.ServerID)

		if e != nil {
			return empty, e
		}

		actual, e := sourceRouter.FetchWgPeer(ctx, peer.PeerID)

		if e != nil {
			return empty, e
		}

		peer.PresharedKey = actual.PresharedKey

		if peer.PresharedKey != nil && (*peer.PresharedKey == "" || *peer.PresharedKey == "none") {
			peer.PresharedKey = nil
		}

		if peer.PresharedKey != nil && !validWireGuardKey(*peer.PresharedKey) {
			return empty, ErrRouterPeerIdentity
		}

		if peer.QueueID != nil {
			queue, e := sourceRouter.FetchSimpleQueue(ctx, *peer.QueueID)

			if e != nil {
				return empty, e
			}

			if queue.MaxLimit != nil {
				parts := strings.Split(*queue.MaxLimit, "/")

				if len(parts) != 2 {
					return empty, ErrRouterPeerIdentity
				}

				peer.DownloadBandwidth = &parts[0]

				peer.UploadBandwidth = &parts[1]
			}
		}

		if e = verifyMigrationSource(ctx, sourceRouter, peer); e != nil {
			return empty, e
		}

		sourceIface, e := s.peers.getInterfaceByName(peer.ServerID, peer.Interface)

		if e != nil {
			return empty, e
		}

		address := net.ParseIP(strings.Split(peer.AllowedAddress, "/")[0]).To4()

		conflict := address == nil || bytes.Compare(address, start) < 0 || bytes.Compare(address, end) > 0

		for _, ip := range used {
			if ip.Equal(address) {
				conflict = true
			}
		}

		if conflict {
			address, e = utils.NextAvailableIPv4(start, end, used)

			if e != nil {
				return empty, ErrMigrationConflict
			}
		}

		used = append(used, address)

		target := peer

		target.PeerID = ""

		target.QueueID = nil

		target.SchedulerID = nil

		target.ServerID = iface.ServerID

		target.Interface = iface.Name

		target.AllowedAddress = address.String() + "/32"

		target.Endpoint = endpoint

		target.EndpointPort = iface.ListenPort

		sourceJSON, _ := json.Marshal(peer)

		targetJSON, _ := json.Marshal(target)

		rows = append(rows, model.RouterMigrationPeer{ID: uuid.NewString(), MigrationID: job.ID, PeerID: peer.ID, PublicKey: peer.PublicKey, SourceServerID: peer.ServerID, SourceInterfaceID: sourceIface.ID, SourcePeerID: peer.PeerID, DestinationServerID: iface.ServerID, SourceSnapshot: string(sourceJSON), DestinationSnapshot: string(targetJSON), Phase: "pending"})
	}

	err = s.billing.accountTransaction(resellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var previous model.RouterMigration

		if e := tx.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&previous).Error; e == nil {
			if previous.RequestHash != hash {
				return ErrIdempotencyConflict
			}

			job = previous

			return nil
		} else if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}

		var current model.Reseller

		if e := tx.First(&current, resellerID).Error; e != nil {
			return e
		}

		if current.MigrationID != nil {
			return ErrPendingPeerOperation
		}

		var count int64

		if e := tx.Model(&model.Peer{}).Where("reseller_id = ?", resellerID).Count(&count).Error; e != nil {
			return e
		}

		if count != int64(len(owned)) {
			return ErrPendingPeerOperation
		}

		// Lock the destination pool against pending address reservations.
		if e := tx.Model(&model.Interface{}).Where("id = ?", iface.ID).UpdateColumn("updated_at", gorm.Expr("updated_at")).Error; e != nil {
			return e
		}

		for _, row := range rows {
			var original model.Peer

			json.Unmarshal([]byte(row.SourceSnapshot), &original)

			locked := tx.Model(&model.Peer{}).Where("id = ? AND reseller_id = ? AND purchase_id IS NULL AND migration_id IS NULL AND restriction_revision = ? AND updated_at = ? AND sync_lease_until <= ?", row.PeerID, resellerID, original.RestrictionRevision, original.UpdatedAt, s.billing.now().Unix()).UpdateColumn("migration_id", job.ID)

			if locked.Error != nil {
				return locked.Error
			}

			if locked.RowsAffected != 1 {
				return ErrPendingPeerOperation
			}

			var target model.Peer

			json.Unmarshal([]byte(row.DestinationSnapshot), &target)

			if e := NewPeerPurchases(s.billing, s.peers).validateAddress(tx, target, target.AllowedAddress); e != nil {
				return e
			}
		}

		if len(rows) == 0 {
			job.State = "complete"
		}

		if e := tx.Create(&job).Error; e != nil {
			return e
		}

		if len(rows) > 0 {
			if e := tx.Create(&rows).Error; e != nil {
				return e
			}
		}

		updates := map[string]any{"migration_id": job.ID}

		if len(rows) == 0 {
			updates["migration_id"] = nil

			updates["default_interface_id"] = iface.ID
		}

		if e := tx.Model(&current).Updates(updates).Error; e != nil {
			return e
		}

		return billingAudit(tx, resellerID, actor.AccountID, nil, "router_migration_queued", job.ID+":"+req.Reason)
	})

	if err != nil {
		return empty, err
	}

	return s.view(job)
}

func verifyMigrationSource(ctx context.Context, router *mikrotik.Adaptor, peer model.Peer) error {
	remote, err := router.FetchWgPeer(ctx, peer.PeerID)

	if err != nil {
		return err
	}

	if remote.PublicKey != peer.PublicKey || remote.Interface != peer.Interface || remote.AllowedAddress != peer.AllowedAddress || normalizedPresharedKey(remote.PresharedKey) != normalizedPresharedKey(peer.PresharedKey) {
		return ErrRouterPeerIdentity
	}

	if peer.QueueID != nil {
		queue, e := router.FetchSimpleQueue(ctx, *peer.QueueID)

		if e != nil {
			return e
		}

		if queue.Target == nil || *queue.Target != peer.AllowedAddress {
			return ErrRouterPeerIdentity
		}
	}

	if peer.SchedulerID != nil {
		schedule, e := router.FetchScheduler(ctx, *peer.SchedulerID)

		if e != nil {
			return e
		}

		if schedule.OnEvent == nil || *schedule.OnEvent != schedulerOnEvent(peer.PeerID) {
			return ErrRouterPeerIdentity
		}
	}

	return nil
}

// One peer per job/tick keeps external work bounded. A database lease also
// prevents another process from running the same step after a service restart.
func (s *RouterMigrations) Run() {
	if !s.runMu.TryLock() {
		return
	}

	defer s.runMu.Unlock()

	var jobs []model.RouterMigration

	now := s.billing.now().Unix()

	if s.billing.db.Where("state IN ? AND lease_until <= ? AND next_attempt_at <= ?", []string{"pending", "cancelling"}, now, now).Order("created_at ASC").Limit(5).Find(&jobs).Error != nil {
		return
	}

	for _, job := range jobs {
		_ = s.Reconcile(job.ID)
	}
}

func (s *RouterMigrations) Reconcile(id string) error {
	now := s.billing.now().Unix()

	claim := s.billing.db.Model(&model.RouterMigration{}).Where("id = ? AND state IN ? AND lease_until <= ? AND next_attempt_at <= ?", id, []string{"pending", "cancelling"}, now, now).UpdateColumn("lease_until", now+120)

	if claim.Error != nil {
		return claim.Error
	}

	if claim.RowsAffected == 0 {
		return nil
	}

	var job model.RouterMigration

	if err := s.billing.db.First(&job, "id = ?", id).Error; err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)

	defer cancel()

	worker := &RouterMigrations{billing: s.billing.withContext(ctx), peers: s.peers}

	var err error

	if job.State == "cancelling" {
		err = worker.cancelStep(ctx, job)
	} else {
		err = worker.step(ctx, job)
	}

	updates := map[string]any{"lease_until": 0, "next_attempt_at": 0, "last_error": ""}

	if err != nil {
		updates["attempts"] = gorm.Expr("attempts + 1")

		updates["next_attempt_at"] = now + min(int64(300), int64(1)<<min(job.Attempts+1, int64(8)))

		updates["last_error"] = "router or storage verification requires retry"
	}

	updateErr := s.billing.db.Model(&model.RouterMigration{}).Where("id = ?", id).Updates(updates).Error

	if err != nil {
		return err
	}

	return updateErr
}

func (s *RouterMigrations) step(ctx context.Context, job model.RouterMigration) error {
	var row model.RouterMigrationPeer

	err := s.billing.db.Where("migration_id = ? AND phase <> ?", job.ID, "complete").Order("peer_id ASC").First(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.complete(job)
	}

	if err != nil {
		return err
	}

	var source, target model.Peer

	if err = json.Unmarshal([]byte(row.SourceSnapshot), &source); err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(row.DestinationSnapshot), &target); err != nil {
		return err
	}

	if row.Phase == "pending" {
		// Detect changed interface credentials and addressing before preparing objects.
		var destination model.Interface

		if err = s.billing.db.Preload("IPPool").First(&destination, job.DestinationInterfaceID).Error; err != nil {
			return err
		}

		if destination.ServerID != target.ServerID || destination.Name != target.Interface || destination.IPPool == nil || destination.Disabled {
			return ErrMigrationConflict
		}

		if err = NewPeerPurchases(s.billing, s.peers).validateAddress(s.billing.db, target, target.AllowedAddress); err != nil {
			return err
		}

		prepared, e := NewPeerPurchases(s.billing, s.peers).provision(ctx, target)

		if e != nil {
			return e
		}

		preparedJSON, _ := json.Marshal(prepared)

		if e = s.billing.db.Model(&row).Where("phase = ?", "pending").Update("destination_snapshot", string(preparedJSON)).Error; e != nil {
			return e
		}

		leaseUntil := s.billing.now().Unix() + 120

		claim := s.billing.db.Model(&model.Peer{}).Where("id = ? AND server_id = ? AND peer_id = ? AND migration_id = ? AND sync_lease_until <= ?", source.ID, source.ServerID, source.PeerID, job.ID, s.billing.now().Unix()).UpdateColumn("sync_lease_until", leaseUntil)

		if claim.Error != nil {
			return claim.Error
		}

		if claim.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}

		defer s.peers.db.Model(&model.Peer{}).Where("id = ? AND server_id = ? AND peer_id = ? AND sync_lease_until = ?", source.ID, source.ServerID, source.PeerID, leaseUntil).UpdateColumn("sync_lease_until", 0)

		// A successfully provisioned destination remains disabled through the switch.
		src, e := s.peers.mikrotikAdaptor.ForServer(source.ServerID)

		if e != nil {
			return e
		}

		if e = verifyMigrationSource(ctx, src, source); e != nil {
			return e
		}

		if _, e = src.UpdateWgPeer(ctx, source.PeerID, mikrotik.WireGuardPeer{Disabled: "true"}); e != nil {
			return e
		}

		observed, e := src.FetchWgPeer(ctx, source.PeerID)

		if e != nil {
			return e
		}

		if observed.PublicKey != source.PublicKey || observed.Disabled != "true" {
			return ErrRouterPeerIdentity
		}

		sourceTx, e := strconv.ParseInt(observed.TransferTx, 10, 64)

		if e != nil || sourceTx < 0 {
			return ErrInvalidBillingQuantity
		}

		sourceRx, e := strconv.ParseInt(observed.TransferRx, 10, 64)

		if e != nil || sourceRx < 0 {
			return ErrInvalidBillingQuantity
		}

		dest, e := s.peers.mikrotikAdaptor.ForServer(target.ServerID)

		if e != nil {
			return e
		}

		ready, e := dest.FetchWgPeer(ctx, prepared.PeerID)

		if e != nil {
			return e
		}

		if ready.PublicKey != source.PublicKey || ready.Interface != target.Interface || ready.AllowedAddress != target.AllowedAddress || ready.Disabled != "true" {
			return ErrRouterPeerIdentity
		}

		destTx, e := strconv.ParseInt(ready.TransferTx, 10, 64)

		if e != nil || destTx < 0 {
			return ErrInvalidBillingQuantity
		}

		destRx, e := strconv.ParseInt(ready.TransferRx, 10, 64)

		if e != nil || destRx < 0 {
			return ErrInvalidBillingQuantity
		}

		encoded, _ := json.Marshal(prepared)

		err = s.billing.accountTransaction(job.ResellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
			var current model.Peer

			if e := tx.Where("reseller_id = ? AND migration_id = ? AND server_id = ? AND peer_id = ?", job.ResellerID, job.ID, source.ServerID, source.PeerID).First(&current, row.PeerID).Error; e != nil {
				return e
			}

			deltaTx, _ := utils.PeerCounterDelta(current.LastTx, sourceTx, 1<<32)

			deltaRx, _ := utils.PeerCounterDelta(current.LastRx, sourceRx, 1<<32)

			download, e := checkedRialAdd(current.DownloadUsage, deltaTx)

			if e != nil {
				return e
			}

			upload, e := checkedRialAdd(current.UploadUsage, deltaRx)

			if e != nil {
				return e
			}

			if _, e = checkedRialAdd(download, upload); e != nil {
				return e
			}

			updates := map[string]any{"server_id": prepared.ServerID, "peer_id": prepared.PeerID, "interface": prepared.Interface, "allowed_address": prepared.AllowedAddress, "endpoint": prepared.Endpoint, "endpoint_port": prepared.EndpointPort, "preshared_key": prepared.PresharedKey, "download_bandwidth": prepared.DownloadBandwidth, "upload_bandwidth": prepared.UploadBandwidth, "queue_id": prepared.QueueID, "scheduler_id": prepared.SchedulerID, "last_tx": destTx, "last_rx": destRx, "download_usage": download, "upload_usage": upload, "restriction_revision": gorm.Expr("restriction_revision + 1"), "synced_disabled": nil, "sync_next_attempt_at": 0, "sync_lease_until": 0}

			delta, e := checkedRialAdd(deltaTx, deltaRx)

			if e != nil {
				return e
			}

			total := model.TotalTrafficUsage{Model: model.Model{ID: model.TotalTrafficUsageSingletonID}}

			if e = tx.FirstOrCreate(&total, model.TotalTrafficUsage{Model: model.Model{ID: model.TotalTrafficUsageSingletonID}}).Error; e != nil {
				return e
			}

			if _, e = checkedRialAdd(total.TotalUsage, delta); e != nil {
				return e
			}

			if e = tx.Model(&total).UpdateColumn("total_usage", gorm.Expr("total_usage + ?", delta)).Error; e != nil {
				return e
			}

			if e = tx.Model(&current).Updates(updates).Error; e != nil {
				return e
			}

			return tx.Model(&row).Updates(map[string]any{"phase": "switched", "destination_snapshot": string(encoded)}).Error
		})

		if err != nil {
			return err
		}

		row.Phase = "switched"

		target = prepared
	}

	if row.Phase == "switched" {
		var current model.Peer

		if err = s.billing.db.Where("reseller_id = ? AND server_id = ? AND peer_id = ?", job.ResellerID, target.ServerID, target.PeerID).First(&current, row.PeerID).Error; err != nil {
			return err
		}

		iface, e := s.peers.getInterfaceByName(target.ServerID, target.Interface)

		if e != nil {
			return e
		}

		if e = s.peers.generatePeerAssets(current.PrivateKey, current, iface.PublicKey); e != nil {
			return e
		}

		// Assets and destination state now agree with durable routing metadata.
		// Clear only the migration lock; all financial/manual restrictions survive.
		if e = s.billing.db.Model(&model.Peer{}).Where("id = ? AND migration_id = ?", current.ID, job.ID).Updates(map[string]any{"migration_id": nil, "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0}).Error; e != nil {
			return e
		}

		if e = NewPeerReconciler(s.billing.db, s.peers.mikrotikAdaptor, s.billing).reconcilePeerContext(ctx, current.ID); e != nil {
			return e
		}

		s.billing.db.First(&current, current.ID)

		dest, e := s.peers.mikrotikAdaptor.ForServer(current.ServerID)

		if e != nil {
			return e
		}

		verified, e := dest.FetchWgPeer(ctx, current.PeerID)

		if e != nil {
			return e
		}

		if verified.PublicKey != current.PublicKey || verified.Interface != current.Interface || verified.AllowedAddress != current.AllowedAddress || verified.Disabled != strconv.FormatBool(current.Disabled) {
			return ErrRouterPeerIdentity
		}

		if e = s.cleanupSource(ctx, source); e != nil {
			return e
		}

		if err = s.billing.db.Model(&row).Update("phase", "complete").Error; err != nil {
			return err
		}
	}

	return s.complete(job)
}

func (s *RouterMigrations) cleanupSource(ctx context.Context, source model.Peer) error {
	router, err := s.peers.mikrotikAdaptor.ForServer(source.ServerID)

	if err != nil {
		return err
	}

	remote, err := router.FetchWgPeer(ctx, source.PeerID)

	if err != nil && !httphelper.IsNotFound(err) {
		return err
	}

	if err == nil {
		if remote.PublicKey != source.PublicKey || remote.Interface != source.Interface || remote.AllowedAddress != source.AllowedAddress || remote.Disabled != "true" {
			return ErrRouterPeerIdentity
		}
	}

	if source.SchedulerID != nil {
		schedule, e := router.FetchScheduler(ctx, *source.SchedulerID)

		if e == nil {
			if schedule.OnEvent == nil || *schedule.OnEvent != schedulerOnEvent(source.PeerID) {
				return ErrRouterPeerIdentity
			}

			if e = router.DeleteScheduler(ctx, *source.SchedulerID); e != nil && !httphelper.IsNotFound(e) {
				return e
			}
		} else if !httphelper.IsNotFound(e) {
			return e
		}
	}

	if source.QueueID != nil {
		queue, e := router.FetchSimpleQueue(ctx, *source.QueueID)

		if e == nil {
			if queue.Target == nil || *queue.Target != source.AllowedAddress {
				return ErrRouterPeerIdentity
			}

			if e = router.DeleteSimpleQueue(ctx, *source.QueueID); e != nil && !httphelper.IsNotFound(e) {
				return e
			}
		} else if !httphelper.IsNotFound(e) {
			return e
		}
	}

	if err == nil {
		if err = router.DeleteWgPeer(ctx, source.PeerID); err != nil && !httphelper.IsNotFound(err) {
			return err
		}
	}

	if _, err = router.FetchWgPeer(ctx, source.PeerID); !httphelper.IsNotFound(err) {
		if err != nil {
			return err
		}

		return ErrRouterSyncPending
	}

	return nil
}

func (s *RouterMigrations) complete(job model.RouterMigration) error {
	return s.billing.accountTransaction(job.ResellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var pending int64

		if err := tx.Model(&model.RouterMigrationPeer{}).Where("migration_id = ? AND phase <> ?", job.ID, "complete").Count(&pending).Error; err != nil {
			return err
		}

		if pending > 0 {
			return nil
		}

		result := tx.Model(&model.RouterMigration{}).Where("id = ? AND state = ?", job.ID, "pending").Update("state", "complete")

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return nil
		}

		if err := tx.Model(&model.Reseller{}).Where("id = ? AND migration_id = ?", job.ResellerID, job.ID).Updates(map[string]any{"migration_id": nil, "default_interface_id": job.DestinationInterfaceID}).Error; err != nil {
			return err
		}

		return billingAudit(tx, job.ResellerID, job.ActorID, nil, "router_migration_completed", job.ID)
	})
}

func (s *RouterMigrations) List(actor Principal, resellerID uint) ([]RouterMigrationView, error) {
	if !billingAdmin(actor) {
		return nil, ErrRoleForbidden
	}

	var jobs []model.RouterMigration

	if err := s.billing.db.Where("reseller_id = ?", resellerID).Order("created_at DESC").Limit(50).Find(&jobs).Error; err != nil {
		return nil, err
	}

	result := make([]RouterMigrationView, 0, len(jobs))

	for _, job := range jobs {
		view, err := s.view(job)

		if err != nil {
			return nil, err
		}

		result = append(result, view)
	}

	return result, nil
}

func migrationUnlocked(tx *gorm.DB, resellerID uint) error {
	var reseller model.Reseller

	if err := tx.First(&reseller, resellerID).Error; err != nil {
		return err
	}

	if reseller.MigrationID != nil {
		return ErrPendingPeerOperation
	}

	return nil
}

func (s *Resellers) BeginMigration(actor Principal, id uint, key string, req RouterMigrationRequest) (RouterMigrationView, error) {
	return NewRouterMigrations(NewBilling(s.db), s.peers).Begin(actor, id, key, req)
}

func (s *Resellers) MigrationHistory(actor Principal, id uint) ([]RouterMigrationView, error) {
	return NewRouterMigrations(NewBilling(s.db), s.peers).List(actor, id)
}

// Cancellation is compensating work, not deletion of migration history. It is
// permitted only before any peer's routing metadata switches to the destination.
func (s *RouterMigrations) Cancel(actor Principal, resellerID uint, id, reason string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}

	if strings.TrimSpace(reason) == "" || len(reason) > 1000 {
		return ErrResellerSettings
	}

	return s.billing.accountTransaction(resellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var job model.RouterMigration

		if err := tx.Where("reseller_id = ?", resellerID).First(&job, "id = ?", id).Error; err != nil {
			return err
		}

		if job.State == "cancelled" || job.State == "cancelling" {
			return nil
		}

		if job.State != "pending" || job.LeaseUntil > s.billing.now().Unix() {
			return ErrPendingPeerOperation
		}

		var switched int64

		if err := tx.Model(&model.RouterMigrationPeer{}).Where("migration_id = ? AND phase <> ?", id, "pending").Count(&switched).Error; err != nil {
			return err
		}

		if switched > 0 {
			return ErrMigrationConflict
		}

		result := tx.Model(&job).Where("state = ? AND lease_until <= ?", "pending", s.billing.now().Unix()).Updates(map[string]any{"state": "cancelling", "next_attempt_at": 0, "last_error": ""})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}

		return billingAudit(tx, resellerID, actor.AccountID, nil, "router_migration_cancel", id+":"+reason)
	})
}

func (s *RouterMigrations) cancelStep(ctx context.Context, job model.RouterMigration) error {
	var row model.RouterMigrationPeer

	err := s.billing.db.Where("migration_id = ? AND phase = ?", job.ID, "pending").Order("peer_id ASC").First(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.finishCancellation(job)
	}

	if err != nil {
		return err
	}

	var source, target model.Peer

	if err = json.Unmarshal([]byte(row.SourceSnapshot), &source); err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(row.DestinationSnapshot), &target); err != nil {
		return err
	}

	src, err := s.peers.mikrotikAdaptor.ForServer(source.ServerID)

	if err != nil {
		return err
	}

	if err = verifyMigrationSource(ctx, src, source); err != nil {
		return err
	}

	dest, err := s.peers.mikrotikAdaptor.ForServer(target.ServerID)

	if err != nil {
		return err
	}

	inventory, err := dest.FetchWgPeers(ctx)

	if err != nil {
		return err
	}

	target.QueueID = nil

	target.SchedulerID = nil

	found := false

	for _, p := range inventory {
		if p.PublicKey == source.PublicKey {
			if found || p.Interface != target.Interface || p.AllowedAddress != target.AllowedAddress {
				return ErrRouterPeerIdentity
			}

			target.PeerID = p.ID

			found = true
		}
	}

	queues, err := dest.FetchSimpleQueues(ctx)

	if err != nil {
		return err
	}

	for _, q := range queues {
		if q.Name == common.QueueName+source.UUID {
			if target.QueueID != nil || q.Target == nil || *q.Target != target.AllowedAddress {
				return ErrRouterPeerIdentity
			}

			id := q.ID

			target.QueueID = &id
		}
	}

	schedules, err := dest.FetchSchedulers(ctx)

	if err != nil {
		return err
	}

	for _, schedule := range schedules {
		if schedule.Name == common.SchedulerName+source.UUID {
			if target.SchedulerID != nil || target.PeerID == "" || schedule.OnEvent == nil || *schedule.OnEvent != schedulerOnEvent(target.PeerID) {
				return ErrRouterPeerIdentity
			}

			id := schedule.ID

			target.SchedulerID = &id
		}
	}

	if found {
		if _, err = dest.UpdateWgPeer(ctx, target.PeerID, mikrotik.WireGuardPeer{Disabled: "true"}); err != nil {
			return err
		}
	}

	if found || target.QueueID != nil || target.SchedulerID != nil {
		// cleanupSource's identity checks are reused for verified destination objects.
		if target.PeerID == "" {
			target.PeerID = "uncreated:" + job.ID
		}

		if err = s.cleanupSource(ctx, target); err != nil {
			return err
		}
	}

	if err = s.billing.accountTransaction(job.ResellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		result := tx.Model(&model.Peer{}).Where("id = ? AND reseller_id = ? AND server_id = ? AND peer_id = ? AND migration_id = ?", source.ID, job.ResellerID, source.ServerID, source.PeerID, job.ID).Updates(map[string]any{"migration_id": nil, "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}

		return tx.Model(&row).Update("phase", "cancelled").Error
	}); err != nil {
		return err
	}

	if err = NewPeerReconciler(s.billing.db, s.peers.mikrotikAdaptor, s.billing).reconcilePeerContext(ctx, source.ID); err != nil {
		return err
	}

	return s.finishCancellation(job)
}

func (s *RouterMigrations) finishCancellation(job model.RouterMigration) error {
	return s.billing.accountTransaction(job.ResellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var pending int64

		if err := tx.Model(&model.RouterMigrationPeer{}).Where("migration_id = ? AND phase <> ?", job.ID, "cancelled").Count(&pending).Error; err != nil {
			return err
		}

		if pending > 0 {
			return nil
		}

		result := tx.Model(&model.RouterMigration{}).Where("id = ? AND state = ?", job.ID, "cancelling").Update("state", "cancelled")

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return nil
		}

		if err := tx.Model(&model.Reseller{}).Where("id = ? AND migration_id = ?", job.ResellerID, job.ID).Update("migration_id", nil).Error; err != nil {
			return err
		}

		return billingAudit(tx, job.ResellerID, job.ActorID, nil, "router_migration_cancelled", job.ID)
	})
}

func (s *Resellers) CancelMigration(actor Principal, id uint, job, reason string) error {
	return NewRouterMigrations(NewBilling(s.db), s.peers).Cancel(actor, id, job, reason)
}

func validMigrationEndpoint(value string) bool {
	if value == "" || len(value) > 253 || strings.TrimSpace(value) != value {
		return false
	}

	if net.ParseIP(strings.Trim(value, "[]")) != nil {
		return true
	}

	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}

		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}

	return true
}
