package service

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrResellerSettings        = errors.New("invalid reseller settings")
	ErrResourceForbidden       = errors.New("interface is not assigned to this reseller")
	ErrResellerLimit           = errors.New("reseller resource limit exceeded")
	ErrRouterMigrationRequired = errors.New("existing peers require a verified router migration before changing the default interface")
	ErrPendingPeerOperation    = errors.New("peer has an unresolved financial operation")
)

type ResellerSettings struct {
	IntervalDays         *int64 `json:"interval_days"`
	GraceDays            *int64 `json:"grace_days"`
	RecurringFeeRial     *int64 `json:"recurring_fee_rial"`
	Name                 string `json:"name"`
	Mode                 string `json:"billing_mode"`
	RateRial             *int64 `json:"rate_rial"`
	InterfaceIDs         []uint `json:"interface_ids"`
	DefaultInterfaceID   *uint  `json:"default_interface_id"`
	MaxActivePeers       *int64 `json:"max_active_peers"`
	MaxPeerTrafficBytes  *int64 `json:"max_peer_traffic_bytes"`
	MaxTotalTrafficBytes *int64 `json:"max_total_traffic_bytes"`
	MaxDurationDays      *int64 `json:"max_duration_days"`
}

type CreateReseller struct {
	ResellerSettings
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResellerResourceView struct {
	InterfaceID uint   `json:"interface_id"`
	Name        string `json:"name"`
	ServerID    uint   `json:"server_id"`
	PoolID      *uint  `json:"pool_id"`
}

type ResellerView struct {
	MigrationID          *string                `json:"migration_id"`
	IntervalDays         int64                  `json:"interval_days"`
	GraceDays            int64                  `json:"grace_days"`
	RecurringFeeRial     string                 `json:"recurring_fee_rial"`
	AdminSuspended       bool                   `json:"admin_suspended"`
	BillingSuspended     bool                   `json:"billing_suspended"`
	NextInvoiceAt        int64                  `json:"next_invoice_at"`
	ID                   uint                   `json:"id"`
	Username             string                 `json:"username"`
	Name                 string                 `json:"name"`
	IsActive             bool                   `json:"is_active"`
	Mode                 string                 `json:"billing_mode"`
	RateRial             *string                `json:"rate_rial"`
	BalanceRial          string                 `json:"balance_rial"`
	CreditRial           string                 `json:"credit_rial"`
	DebtRial             string                 `json:"debt_rial"`
	ReservedRial         string                 `json:"reserved_rial"`
	ActivePeers          int64                  `json:"active_peers"`
	TotalPeers           int64                  `json:"total_peers"`
	DefaultInterfaceID   *uint                  `json:"default_interface_id"`
	MaxActivePeers       *string                `json:"max_active_peers"`
	MaxPeerTrafficBytes  *string                `json:"max_peer_traffic_bytes"`
	MaxTotalTrafficBytes *string                `json:"max_total_traffic_bytes"`
	MaxDurationDays      *string                `json:"max_duration_days"`
	Resources            []ResellerResourceView `json:"resources"`
}

type Resellers struct {
	db    *gorm.DB
	peers *WgPeer
}

func NewResellers(db *gorm.DB, peers *WgPeer) *Resellers { return &Resellers{db: db, peers: peers} }

func validateResellerSettings(tx *gorm.DB, settings ResellerSettings) error {
	if strings.TrimSpace(settings.Name) == "" || len(settings.Name) > 255 || (settings.Mode != model.BillingPrepaid && settings.Mode != model.BillingPostpaid) {
		return ErrResellerSettings
	}
	for _, limit := range []*int64{settings.RateRial, settings.RecurringFeeRial, settings.GraceDays, settings.MaxActivePeers, settings.MaxPeerTrafficBytes, settings.MaxTotalTrafficBytes, settings.MaxDurationDays} {
		if limit != nil && *limit < 0 {
			return ErrResellerSettings
		}
	}
	if settings.IntervalDays != nil {
		seconds, err := billingDays(*settings.IntervalDays)
		if err != nil || seconds == 0 {
			return ErrResellerSettings
		}
	}
	if settings.GraceDays != nil {
		if _, err := billingDays(*settings.GraceDays); err != nil {
			return ErrResellerSettings
		}
	}
	seen := make(map[uint]bool)
	for _, id := range settings.InterfaceIDs {
		if id == 0 || seen[id] {
			return ErrResellerSettings
		}
		seen[id] = true
		var iface model.Interface
		if err := tx.First(&iface, id).Error; err != nil {
			return err
		}
		var server model.Server
		if err := tx.Where("is_active = ?", true).First(&server, iface.ServerID).Error; err != nil {
			return err
		}
	}
	if settings.DefaultInterfaceID != nil && !seen[*settings.DefaultInterfaceID] {
		return ErrResourceForbidden
	}
	return nil
}

func writeResellerResources(tx *gorm.DB, id uint, interfaces []uint) error {
	if err := tx.Unscoped().Where("reseller_id = ?", id).Delete(&model.ResellerResource{}).Error; err != nil {
		return err
	}
	for _, interfaceID := range interfaces {
		if err := tx.Create(&model.ResellerResource{ResellerID: id, InterfaceID: interfaceID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Resellers) Create(actor Principal, input CreateReseller) (ResellerView, error) {
	if !billingAdmin(actor) {
		return ResellerView{}, ErrRoleForbidden
	}
	if strings.TrimSpace(input.Username) != input.Username || input.Username == "" || len(input.Username) > 64 || len(input.Password) < 12 || len(input.Password) > 72 {
		return ResellerView{}, ErrResellerSettings
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return ResellerView{}, err
	}
	var id uint
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := validateResellerSettings(tx, input.ResellerSettings); err != nil {
			return err
		}
		reseller := model.Reseller{Name: strings.TrimSpace(input.Name), DefaultInterfaceID: input.DefaultInterfaceID, MaxActivePeers: input.MaxActivePeers, MaxPeerTrafficBytes: input.MaxPeerTrafficBytes, MaxTotalTrafficBytes: input.MaxTotalTrafficBytes, MaxDurationDays: input.MaxDurationDays}
		if err := tx.Create(&reseller).Error; err != nil {
			return err
		}
		id = reseller.ID
		if err := tx.Create(&model.Admin{Username: input.Username, Password: string(hashed), Role: model.RoleReseller, ResellerID: &id}).Error; err != nil {
			return err
		}
		if err := NewBilling(tx).CreateAccount(actor, id, input.Mode); err != nil {
			return err
		}
		var account model.BillingAccount
		if err := tx.Where("reseller_id = ?", id).First(&account).Error; err != nil {
			return err
		}
		if err := applyFinancialSettings(tx, &account, input.ResellerSettings); err != nil {
			return err
		}
		if err := writeResellerResources(tx, id, input.InterfaceIDs); err != nil {
			return err
		}
		encoded, _ := json.Marshal(input.ResellerSettings)
		return billingAudit(tx, id, actor.AccountID, nil, "reseller_created", string(encoded))
	})
	if err != nil {
		return ResellerView{}, err
	}
	return s.Get(actor, id)
}

func (s *Resellers) Update(actor Principal, id uint, settings ResellerSettings) (ResellerView, error) {
	if !billingAdmin(actor) {
		return ResellerView{}, ErrRoleForbidden
	}
	if _, err := NewBilling(s.db).ReconcileAccount(id); err != nil {
		return ResellerView{}, err
	}

	err := NewBilling(s.db).accountTransaction(id, func(tx *gorm.DB, account *model.BillingAccount) error {
		if err := validateResellerSettings(tx, settings); err != nil {
			return err
		}
		var reseller model.Reseller
		if err := tx.First(&reseller, id).Error; err != nil {
			return err
		}
		if reseller.MigrationID != nil {
			if !uintPointersEqual(reseller.DefaultInterfaceID, settings.DefaultInterfaceID) {
				return ErrPendingPeerOperation
			}
			var existing []model.ResellerResource
			if e := tx.Where("reseller_id = ?", id).Find(&existing).Error; e != nil {
				return e
			}
			assigned := map[uint]bool{}
			for _, resource := range existing {
				assigned[resource.InterfaceID] = true
			}
			if len(assigned) != len(settings.InterfaceIDs) {
				return ErrPendingPeerOperation
			}
			for _, interfaceID := range settings.InterfaceIDs {
				if !assigned[interfaceID] {
					return ErrPendingPeerOperation
				}
			}
		}
		if !uintPointersEqual(reseller.DefaultInterfaceID, settings.DefaultInterfaceID) && settings.DefaultInterfaceID != nil {
			var iface model.Interface
			if err := tx.First(&iface, *settings.DefaultInterfaceID).Error; err != nil {
				return err
			}
			var incompatible int64
			if err := tx.Model(&model.Peer{}).Where("reseller_id = ? AND (server_id <> ? OR interface <> ?)", id, iface.ServerID, iface.Name).Count(&incompatible).Error; err != nil {
				return err
			}
			if incompatible > 0 {
				return ErrRouterMigrationRequired
			}
		}
		// A resource in use cannot be silently revoked underneath an existing peer.
		var owned []model.Peer
		if err := tx.Where("reseller_id = ?", id).Find(&owned).Error; err != nil {
			return err
		}
		for _, peer := range owned {
			var iface model.Interface
			if err := tx.Where("server_id = ? AND name = ?", peer.ServerID, peer.Interface).First(&iface).Error; err != nil {
				return err
			}
			permitted := false
			for _, allowed := range settings.InterfaceIDs {
				if allowed == iface.ID {
					permitted = true
				}
			}
			if !permitted {
				return ErrResourceForbidden
			}
		}
		data := map[string]any{"name": strings.TrimSpace(settings.Name), "default_interface_id": settings.DefaultInterfaceID, "max_active_peers": settings.MaxActivePeers, "max_peer_traffic_bytes": settings.MaxPeerTrafficBytes, "max_total_traffic_bytes": settings.MaxTotalTrafficBytes, "max_duration_days": settings.MaxDurationDays}
		if err := tx.Model(&reseller).Updates(data).Error; err != nil {
			return err
		}
		if err := writeResellerResources(tx, id, settings.InterfaceIDs); err != nil {
			return err
		}
		if err := applyFinancialSettings(tx, account, settings); err != nil {
			return err
		}
		encoded, _ := json.Marshal(settings)
		return billingAudit(tx, id, actor.AccountID, nil, "reseller_settings_changed", string(encoded))
	})
	if err != nil {
		return ResellerView{}, err
	}
	return s.Get(actor, id)
}

func uintPointersEqual(a, b *uint) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}

func (s *Resellers) Get(actor Principal, id uint) (ResellerView, error) {
	var view ResellerView
	if !billingAdmin(actor) {
		return view, ErrRoleForbidden
	}
	var reseller model.Reseller
	if err := s.db.First(&reseller, id).Error; err != nil {
		return view, err
	}
	var login model.Admin
	if err := s.db.Where("reseller_id = ? AND role = ?", id, model.RoleReseller).First(&login).Error; err != nil {
		return view, err
	}
	account, err := NewBilling(s.db).Account(actor, id)
	if err != nil {
		return view, err
	}
	view = ResellerView{MigrationID: reseller.MigrationID, IntervalDays: account.IntervalDays, GraceDays: account.GraceDays, RecurringFeeRial: strconv.FormatInt(account.RecurringFeeRial, 10), AdminSuspended: reseller.AdminSuspended, BillingSuspended: account.BillingSuspended, NextInvoiceAt: account.NextInvoiceAt, ID: id, Username: login.Username, Name: reseller.Name, IsActive: reseller.IsActive, Mode: account.Mode, BalanceRial: strconv.FormatInt(account.BalanceRial, 10), CreditRial: "0", DebtRial: "0", ReservedRial: strconv.FormatInt(account.ReservedRial, 10), DefaultInterfaceID: reseller.DefaultInterfaceID, MaxActivePeers: int64Text(reseller.MaxActivePeers), MaxPeerTrafficBytes: int64Text(reseller.MaxPeerTrafficBytes), MaxTotalTrafficBytes: int64Text(reseller.MaxTotalTrafficBytes), MaxDurationDays: int64Text(reseller.MaxDurationDays), Resources: []ResellerResourceView{}}
	if account.RateRial != nil {
		rate := strconv.FormatInt(*account.RateRial, 10)
		view.RateRial = &rate
	}
	if account.BalanceRial >= 0 {
		view.CreditRial = strconv.FormatInt(account.BalanceRial, 10)
	} else {
		view.DebtRial = strings.TrimPrefix(view.BalanceRial, "-")
	}
	var peers []model.Peer
	if err := s.db.Where("reseller_id = ?", id).Find(&peers).Error; err != nil {
		return view, err
	}
	view.TotalPeers = int64(len(peers))
	for _, peer := range peers {
		if !peer.Disabled && !utils.IsPeerExpired(peer.ExpireTime, time.Now()) && (peer.TrafficLimit == nil || peer.DownloadUsage+peer.UploadUsage < *peer.TrafficLimit) {
			view.ActivePeers++
		}
	}
	var resources []model.ResellerResource
	if err := s.db.Preload("Interface.IPPool").Where("reseller_id = ?", id).Order("interface_id ASC").Find(&resources).Error; err != nil {
		return view, err
	}
	for _, resource := range resources {
		item := ResellerResourceView{InterfaceID: resource.InterfaceID, Name: resource.Interface.Name, ServerID: resource.Interface.ServerID}
		if resource.Interface.IPPool != nil {
			poolID := resource.Interface.IPPool.ID
			item.PoolID = &poolID
		}
		view.Resources = append(view.Resources, item)
	}
	return view, nil
}

func (s *Resellers) List(actor Principal) ([]ResellerView, error) {
	if !billingAdmin(actor) {
		return nil, ErrRoleForbidden
	}
	var resellers []model.Reseller
	if err := s.db.Order("id ASC").Find(&resellers).Error; err != nil {
		return nil, err
	}
	result := make([]ResellerView, 0, len(resellers))
	for _, reseller := range resellers {
		view, err := s.Get(actor, reseller.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	return result, nil
}

// ValidatePeerResource is shared by ownership transfer and future billed peer
// operations. Financial commits must revalidate under the account transaction.
func ValidatePeerResource(tx *gorm.DB, resellerID uint, peer model.Peer, excludePeerID uint) error {
	var reseller model.Reseller
	if err := tx.Where("is_active = ?", true).First(&reseller, resellerID).Error; err != nil {
		return err
	}
	var iface model.Interface
	if err := tx.Where("server_id = ? AND name = ?", peer.ServerID, peer.Interface).First(&iface).Error; err != nil {
		return err
	}
	var resource model.ResellerResource
	if err := tx.Where("reseller_id = ? AND interface_id = ?", resellerID, iface.ID).First(&resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResourceForbidden
		}
		return err
	}
	if peer.TrafficLimit == nil || *peer.TrafficLimit <= 0 || peer.ExpireTime == nil {
		return ErrResellerSettings
	}
	expiry, err := time.Parse("2006-01-02", *peer.ExpireTime)
	if err != nil {
		return ErrResellerSettings
	}
	if reseller.MaxPeerTrafficBytes != nil && *peer.TrafficLimit > *reseller.MaxPeerTrafficBytes {
		return ErrResellerLimit
	}
	remainingSeconds := max(int64(0), expiry.Unix()-time.Now().UTC().Unix())
	remainingDays := (remainingSeconds + 86400 - 1) / 86400
	if reseller.MaxDurationDays != nil && remainingDays > *reseller.MaxDurationDays {
		return ErrResellerLimit
	}
	var existing []model.Peer
	if err := tx.Where("reseller_id = ? AND id <> ?", resellerID, excludePeerID).Find(&existing).Error; err != nil {
		return err
	}
	active, total := int64(0), int64(0)
	for _, other := range existing {
		if other.PurchaseID != nil {
			var pending model.PeerPurchase
			if err := tx.First(&pending, "id = ?", *other.PurchaseID).Error; err != nil {
				return err
			}
			other.TrafficLimit = &pending.QuotaBytes
			if pending.Kind == string(BillingCreate) {
				other.Disabled = other.ManualDisabled || other.AdminBlocked || other.BillingBlocked
			}
		}
		if !other.Disabled && !utils.IsPeerExpired(other.ExpireTime, time.Now()) && (other.TrafficLimit == nil || other.DownloadUsage+other.UploadUsage < *other.TrafficLimit) {
			active++
		}
		if other.TrafficLimit != nil {
			next, err := checkedRialAdd(total, *other.TrafficLimit)
			if err != nil {
				return ErrResellerLimit
			}
			total = next
		}
	}
	if !peer.Disabled && !utils.IsPeerExpired(peer.ExpireTime, time.Now()) && (peer.TrafficLimit == nil || peer.DownloadUsage+peer.UploadUsage < *peer.TrafficLimit) {
		active++
	}
	if reseller.MaxActivePeers != nil && active > *reseller.MaxActivePeers {
		return ErrResellerLimit
	}
	total, err = checkedRialAdd(total, *peer.TrafficLimit)
	if err != nil {
		return ErrResellerLimit
	}
	if reseller.MaxTotalTrafficBytes != nil && total > *reseller.MaxTotalTrafficBytes {
		return ErrResellerLimit
	}
	return nil
}

func (s *Resellers) TransferPeer(actor Principal, peerID uint, target *uint, reason string) error {
	return s.TransferPeerPlan(actor, peerID, target, reason, nil)
}

func (s *Resellers) TransferPeerPlan(actor Principal, peerID uint, target *uint, reason string, duration *int64) error {
	if duration != nil && (*duration <= 0 || *duration > 36500) {
		return ErrResellerSettings
	}

	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if strings.TrimSpace(reason) == "" || len(reason) > 1000 || (target != nil && *target == 0) {
		return ErrResellerSettings
	}
	var initial model.Peer
	if err := s.db.First(&initial, peerID).Error; err != nil {
		return err
	}
	accounts := []uint{}
	if initial.ResellerID != nil {
		accounts = append(accounts, *initial.ResellerID)
	}
	if target != nil && (initial.ResellerID == nil || *target != *initial.ResellerID) {
		accounts = append(accounts, *target)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i] < accounts[j] })
	transfer := func(tx *gorm.DB) error {
		// Account locks precede peer locks, matching reservation/posting order.
		for _, id := range accounts {
			if err := migrationUnlocked(tx, id); err != nil {
				return err
			}
			locked := tx.Model(&model.BillingAccount{}).Where("reseller_id = ?", id).UpdateColumn("revision", gorm.Expr("revision + 1"))
			if locked.Error != nil {
				return locked.Error
			}
			if locked.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		locked := tx.Model(&model.Peer{}).Where("id = ?", peerID).UpdateColumn("updated_at", time.Now().Unix())
		if locked.Error != nil {
			return locked.Error
		}
		if locked.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		var peer model.Peer
		if err := tx.First(&peer, peerID).Error; err != nil {
			return err
		}
		if peer.PurchaseID != nil || !uintPointersEqual(peer.ResellerID, initial.ResellerID) {
			return ErrPendingPeerOperation
		}
		if uintPointersEqual(peer.ResellerID, target) && duration == nil {
			return nil
		}
		var pending int64
		if err := tx.Model(&model.BillingCharge{}).Where("peer_id = ? AND state = ?", peerID, model.ChargeReserved).Count(&pending).Error; err != nil {
			return err
		}
		if pending > 0 {
			return ErrPendingPeerOperation
		}
		if target != nil {
			if err := ValidatePeerResource(tx, *target, peer, peer.ID); err != nil {
				return err
			}
		}
		previous := peer.ResellerID
		if previous != nil {
			copyID := *previous
			previous = &copyID
		}
		updates := map[string]any{"reseller_id": target, "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0}
		if duration != nil {
			updates["configured_duration_days"] = *duration
		}
		if err := tx.Model(&peer).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Create(&model.PeerOwnershipEvent{ID: uuid.NewString(), PeerID: peerID, PreviousResellerID: previous, ResellerID: target, ActorID: actor.AccountID, Reason: reason, ConfiguredDurationDays: duration}).Error
	}
	if len(accounts) > 0 {
		return NewBilling(s.db).accountTransaction(accounts[0], func(tx *gorm.DB, _ *model.BillingAccount) error { return transfer(tx) })
	}
	return s.db.Transaction(transfer)
}

func (s *Resellers) PeerList(actor Principal, id uint) (*[]schema.PeerResponse, error) {
	if !billingAdmin(actor) {
		return nil, ErrRoleForbidden
	}
	var reseller model.Reseller
	if err := s.db.First(&reseller, id).Error; err != nil {
		return nil, err
	}
	selected := *s.peers
	selected.tenantID = &id
	return selected.GetPeers()
}

func (s *Resellers) SetSuspension(actor Principal, id uint, suspended bool, reason string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if strings.TrimSpace(reason) == "" || len(reason) > 1000 {
		return ErrResellerSettings
	}
	return NewBilling(s.db).accountTransaction(id, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var reseller model.Reseller
		if err := tx.First(&reseller, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&reseller).Updates(map[string]any{"admin_suspended": suspended, "admin_suspend_reason": reason}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Peer{}).Where("reseller_id = ? AND restrictions_initialized = ?", id, false).Updates(map[string]any{"manual_disabled": gorm.Expr("disabled"), "restrictions_initialized": true}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Peer{}).Where("reseller_id = ?", id).Updates(map[string]any{"admin_blocked": suspended, "disabled": gorm.Expr("provisioning_blocked OR manual_disabled OR quota_blocked OR expiry_blocked OR billing_blocked OR ?", suspended), "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0}).Error; err != nil {
			return err
		}
		event := "administrative_suspension"
		if !suspended {
			event = "administrative_reactivation"
		}
		return billingAudit(tx, id, actor.AccountID, nil, event, reason)
	})
}

func myReseller(actor Principal) (uint, error) {
	if actor.AccountID == 0 || actor.Role != model.RoleReseller || actor.ResellerID == nil || *actor.ResellerID == 0 {
		return 0, ErrRoleForbidden
	}
	return *actor.ResellerID, nil
}
func (s *Resellers) MyPeers(actor Principal) (*[]schema.PeerResponse, error) {
	id, err := myReseller(actor)
	if err != nil {
		return nil, err
	}
	selected := *s.peers
	selected.tenantID = &id
	return selected.GetPeers()
}
func (s *Resellers) MyConfig(actor Principal, id uint) (string, string, error) {
	if _, err := myReseller(actor); err != nil {
		return "", "", err
	}
	if _, err := actor.FindOwnedPeer(s.db, id); err != nil {
		return "", "", err
	}
	return s.peers.configGenerator.GetPeerConfig(id)
}
func (s *Resellers) MyQRCode(actor Principal, id uint) (string, error) {
	if _, err := myReseller(actor); err != nil {
		return "", err
	}
	if _, err := actor.FindOwnedPeer(s.db, id); err != nil {
		return "", err
	}
	return s.peers.qrCodeGenerator.GetPeerQRCode(id)
}
func (s *Resellers) MySessions(actor Principal, id uint) (*schema.PeerSessionsResponse, error) {
	if _, err := myReseller(actor); err != nil {
		return nil, err
	}
	if _, err := actor.FindOwnedPeer(s.db, id); err != nil {
		return nil, err
	}
	return s.peers.GetPeerSessions(id)
}
func (s *Resellers) MyExport(actor Principal) (string, error) {
	if _, err := myReseller(actor); err != nil {
		return "", err
	}
	return NewExcelGenerator(actor.PeerScope(s.db)).GetTrafficUsageReport()
}

func int64Text(value *int64) *string {
	if value == nil {
		return nil
	}
	text := strconv.FormatInt(*value, 10)
	return &text
}

func applyFinancialSettings(tx *gorm.DB, a *model.BillingAccount, settings ResellerSettings) error {
	interval, grace, fee := a.IntervalDays, a.GraceDays, a.RecurringFeeRial
	if settings.IntervalDays != nil {
		interval = *settings.IntervalDays
	}
	if settings.GraceDays != nil {
		grace = *settings.GraceDays
	}
	if settings.RecurringFeeRial != nil {
		fee = *settings.RecurringFeeRial
	}
	seconds, err := billingDays(interval)
	if err != nil || seconds == 0 {
		return ErrResellerSettings
	}
	next := a.NextInvoiceAt
	if interval != a.IntervalDays {
		next, err = checkedRialAdd(a.PeriodStart, seconds)
		if err != nil {
			return err
		}
	}
	if err := tx.Model(a).Updates(map[string]any{"mode": settings.Mode, "rate_rial": settings.RateRial, "interval_days": interval, "grace_days": grace, "recurring_fee_rial": fee, "next_invoice_at": next}).Error; err != nil {
		return err
	}
	return NewBilling(tx).advancePeriods(tx, a, time.Now().UTC().Unix())
}
