package service_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
	"gorm.io/gorm"
)

func migrationFixture(t *testing.T) (multiFixture, model.Peer, service.Principal, *service.Billing, *service.RouterMigrations, model.Interface) {
	t.Helper()

	f, peer, actor, billing, _ := purchaseFixture(t)

	created, err := f.interfaces.CreateInterface(&schema.CreateInterfaceRequest{ServerID: f.b.ID, Name: "wg0", ListenPort: "51830"})

	if err != nil {
		t.Fatal(err)
	}

	if _, err = service.NewIPPool(f.db, f.adaptor).CreateIPPool(&schema.CreateIPPoolRequest{Name: "target pool", InterfaceID: created.Id, StartIP: "10.42.0.2/32", EndIP: "10.42.0.10/32"}); err != nil {
		t.Fatal(err)
	}

	var target model.Interface

	if err = f.db.Preload("IPPool").First(&target, created.Id).Error; err != nil {
		t.Fatal(err)
	}

	if err = f.db.Create(&model.ResellerResource{ResellerID: *actor.ResellerID, InterfaceID: target.ID}).Error; err != nil {
		t.Fatal(err)
	}

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	return f, peer, admin, billing, service.NewRouterMigrations(billing, f.peers), target
}

func retryMigration(t *testing.T, f multiFixture, id string) {
	t.Helper()

	if err := f.db.Model(&model.RouterMigration{}).Where("id = ?", id).Updates(map[string]any{"next_attempt_at": 0, "lease_until": 0}).Error; err != nil {
		t.Fatal(err)
	}
}

func TestRouterMigrationPreservesIdentityBillingUsageAndPublicConfiguration(t *testing.T) {
	f, peer, admin, billing, migrations, target := migrationFixture(t)

	original := peer

	if err := f.db.Model(&peer).Updates(map[string]any{"is_shared": true, "download_usage": 4 * service.BillingGiB, "upload_usage": 256, "last_tx": 100, "last_rx": 20}).Error; err != nil {
		t.Fatal(err)
	}

	src, _ := f.adaptor.ForServer(peer.ServerID)

	if _, err := src.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{TransferTx: "800", TransferRx: "55"}); err != nil {
		t.Fatal(err)
	}

	req := service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "move to second router"}

	job, err := migrations.Begin(admin, *peer.ResellerID, "migration", req)

	if err != nil {
		t.Fatal(err)
	}

	if job.TotalPeers != 1 || job.State != "pending" {
		t.Fatal("inventory missing", job)
	}

	if _, _, err = service.NewConfigGenerator(f.db).GetUserConfig(peer.UUID); !errors.Is(err, service.ErrRouterSyncPending) {
		t.Fatal("pending migration served an old config", err)
	}

	if err = migrations.Reconcile(job.ID); err != nil {
		t.Fatal(err)
	}

	f.db.First(&peer, peer.ID)

	if peer.ServerID != target.ServerID || peer.Interface != target.Name || peer.UUID != original.UUID || peer.PublicKey != original.PublicKey || peer.PrivateKey != original.PrivateKey || *peer.ResellerID != *original.ResellerID || *peer.ExpireTime != *original.ExpireTime || *peer.TrafficLimit != *original.TrafficLimit || peer.ConfiguredDurationDays != original.ConfiguredDurationDays {
		t.Fatal("identity, ownership or plan changed")
	}

	if peer.DownloadUsage != 4*service.BillingGiB+700 || peer.UploadUsage != 291 || peer.LastTx != 0 || peer.LastRx != 0 {
		t.Fatal("usage not preserved", peer.DownloadUsage, peer.UploadUsage)
	}

	dest, _ := f.adaptor.ForServer(target.ServerID)

	remote, err := dest.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil || remote.Disabled != "false" {
		t.Fatal("destination not restored", err)
	}

	peers, err := src.FetchWgPeers(context.Background())

	if err != nil || len(peers) != 0 {
		t.Fatal("source peer not cleaned")
	}

	queues, _ := src.FetchSimpleQueues(context.Background())

	schedules, _ := src.FetchSchedulers(context.Background())

	if len(queues) != 0 || len(schedules) != 0 {
		t.Fatal("source resources retained")
	}

	filename, _, err := service.NewConfigGenerator(f.db).GetUserConfig(original.UUID)

	if err != nil {
		t.Fatal(err)
	}

	contents, err := os.ReadFile(filename)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(contents), target.PublicKey) || !strings.Contains(string(contents), "localhost:51830") {
		t.Fatal("download has stale interface key or endpoint")
	}

	qr, err := service.NewQRCodeGenerator(f.db).GetUserQRCode(original.UUID)

	if err != nil {
		t.Fatal(err)
	}

	if _, err = os.Stat(qr); err != nil {
		t.Fatal(err)
	}

	account, _ := billing.Account(admin, *peer.ResellerID)

	if account.BalanceRial != 10_000_000 || account.ReservedRial != 0 {
		t.Fatal("migration changed accounting")
	}

	var reseller model.Reseller

	f.db.First(&reseller, *peer.ResellerID)

	if reseller.MigrationID != nil || reseller.DefaultInterfaceID == nil || *reseller.DefaultInterfaceID != target.ID {
		t.Fatal("default not finalized")
	}

	replay, err := migrations.Begin(admin, *peer.ResellerID, "migration", req)

	if err != nil || replay.ID != job.ID || replay.State != "complete" {
		t.Fatal("migration replay", err)
	}
}

func TestRouterMigrationAmbiguousDestinationCreationRestartsSafely(t *testing.T) {
	f, peer, admin, billing, migrations, target := migrationFixture(t)

	job, err := migrations.Begin(admin, *peer.ResellerID, "ambiguous", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "test recovery"})

	if err != nil {
		t.Fatal(err)
	}

	f.mockB.FailAfterNext(http.MethodPut, "/interface/wireguard/peers", 1)

	if err = migrations.Reconcile(job.ID); err == nil {
		t.Fatal("expected ambiguous response")
	}

	var current model.Peer

	f.db.First(&current, peer.ID)

	if current.ServerID != peer.ServerID {
		t.Fatal("routing switched before destination verification")
	}

	src, _ := f.adaptor.ForServer(peer.ServerID)

	live, _ := src.FetchWgPeer(context.Background(), peer.PeerID)

	if live.Disabled != "false" {
		t.Fatal("preparation failure disabled source")
	}

	retryMigration(t, f, job.ID)

	if err = service.NewRouterMigrations(billing, f.peers).Reconcile(job.ID); err != nil {
		t.Fatal(err)
	}

	dest, _ := f.adaptor.ForServer(target.ServerID)

	rows, err := dest.FetchWgPeers(context.Background())

	if err != nil || len(rows) != 1 {
		t.Fatal("destination duplicated")
	}
}

func TestRouterMigrationCleanupFailureDoesNotReimportSource(t *testing.T) {
	f, peer, admin, billing, migrations, target := migrationFixture(t)

	job, err := migrations.Begin(admin, *peer.ResellerID, "cleanup", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "cleanup recovery"})

	if err != nil {
		t.Fatal(err)
	}

	f.mockA.FailNext(http.MethodDelete, "/interface/wireguard/peers/"+peer.PeerID, 1)

	if err = migrations.Reconcile(job.ID); err == nil {
		t.Fatal("expected source cleanup failure")
	}

	var current model.Peer

	f.db.First(&current, peer.ID)

	if current.ServerID != target.ServerID {
		t.Fatal("verified routing reverted on cleanup failure")
	}

	sourceSync, err := f.sync.ForServer(peer.ServerID)

	if err != nil {
		t.Fatal(err)
	}

	if err = sourceSync.SyncPeers(); err != nil {
		t.Fatal(err)
	}

	var count int64

	f.db.Model(&model.Peer{}).Count(&count)

	if count != 1 {
		t.Fatal("old source was imported as another tenant/admin peer")
	}

	member := service.Principal{AccountID: 2, Role: model.RoleReseller, ResellerID: peer.ResellerID}

	if _, err = service.NewPeerPurchases(billing, f.peers).Begin(member, *peer.ResellerID, peer.ID, "blocked", service.PeerPurchaseRequest{Operation: service.BillingAddTraffic, Bytes: "1073741824"}); !errors.Is(err, service.ErrPendingPeerOperation) {
		t.Fatal("mutation allowed during unfinished cleanup", err)
	}

	retryMigration(t, f, job.ID)

	if err = service.NewRouterMigrations(billing, f.peers).Reconcile(job.ID); err != nil {
		t.Fatal(err)
	}
}

func TestRouterMigrationStorageFailureAfterSourceDisableCanResume(t *testing.T) {
	f, peer, admin, billing, migrations, target := migrationFixture(t)

	job, err := migrations.Begin(admin, *peer.ResellerID, "db-failure", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "storage recovery"})

	if err != nil {
		t.Fatal(err)
	}

	const callback = "fail_migration_switch"

	if err = f.db.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "peers" && tx.Statement.Changed("ServerID") {
			tx.AddError(errors.New("injected switch failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}

	if err = migrations.Reconcile(job.ID); err == nil {
		t.Fatal("expected storage failure")
	}

	f.db.Callback().Update().Remove(callback)

	var current model.Peer

	f.db.First(&current, peer.ID)

	if current.ServerID != peer.ServerID {
		t.Fatal("partial routing switch")
	}

	source, _ := f.adaptor.ForServer(peer.ServerID)

	remote, _ := source.FetchWgPeer(context.Background(), peer.PeerID)

	if remote.Disabled != "true" {
		t.Fatal("source disable did not precede storage failure")
	}

	retryMigration(t, f, job.ID)

	if err = service.NewRouterMigrations(billing, f.peers).Reconcile(job.ID); err != nil {
		t.Fatal(err)
	}

	f.db.First(&current, peer.ID)

	if current.ServerID != target.ServerID {
		t.Fatal("did not recover")
	}
}

func TestRouterMigrationPreservesMixedRestrictionsAndResolvesAddressConflict(t *testing.T) {
	f, peer, admin, billing, migrations, target := migrationFixture(t)

	var sourceInterface model.Interface

	f.db.Where("server_id = ?", peer.ServerID).First(&sourceInterface)

	manual := extraPeer(t, f, sourceInterface.ID, "manual migration", "10.42.0.3/32")

	expired := extraPeer(t, f, sourceInterface.ID, "expired migration", "10.42.0.4/32")

	quota := extraPeer(t, f, sourceInterface.ID, "quota migration", "10.42.0.5/32")

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	for _, p := range []model.Peer{manual, expired, quota} {
		f.db.Model(&p).Update("reseller_id", *peer.ResellerID)
	}

	f.db.Model(&manual).Update("manual_disabled", true)

	f.db.Model(&expired).Update("expire_time", yesterday)

	f.db.Model(&quota).Update("download_usage", *quota.TrafficLimit)

	// Another tenant/admin peer already occupies .2 on the destination.
	occupied := extraPeer(t, f, target.ID, "occupied destination", "10.42.0.2/32")

	job, err := migrations.Begin(admin, *peer.ResellerID, "mixed", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "mixed states"})

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		if err = migrations.Reconcile(job.ID); err != nil {
			t.Fatal(err)
		}
	}

	f.db.First(&peer, peer.ID)

	if peer.AllowedAddress == occupied.AllowedAddress {
		t.Fatal("destination address conflict retained")
	}

	dest, _ := f.adaptor.ForServer(target.ServerID)

	for _, p := range []model.Peer{manual, expired, quota} {
		f.db.First(&p, p.ID)

		remote, e := dest.FetchWgPeer(context.Background(), p.PeerID)

		if e != nil || remote.Disabled != "true" {
			t.Fatal("restricted peer enabled", p.Name, e)
		}
	}

	f.db.First(&manual, manual.ID)

	if !manual.ManualDisabled {
		t.Fatal("manual restriction lost")
	}

	account, _ := billing.Account(admin, *peer.ResellerID)

	if account.BalanceRial != 10_000_000 {
		t.Fatal("migration billed")
	}
}

func TestRouterMigrationPreflightRejectsUnauthorizedOrUnavailableDestination(t *testing.T) {
	f, peer, admin, _, migrations, target := migrationFixture(t)

	member := service.Principal{AccountID: 2, Role: model.RoleReseller, ResellerID: peer.ResellerID}

	req := service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "test"}

	if _, err := migrations.Begin(member, *peer.ResellerID, "forbidden", req); !errors.Is(err, service.ErrRoleForbidden) {
		t.Fatal(err)
	}

	f.mockB.FailNext(http.MethodGet, "/interface/wireguard/"+target.InterfaceID, 1)

	if _, err := migrations.Begin(admin, *peer.ResellerID, "unavailable", req); err == nil {
		t.Fatal("unavailable destination accepted")
	}

	var current model.Peer

	f.db.First(&current, peer.ID)

	if current.MigrationID != nil || current.ServerID != peer.ServerID {
		t.Fatal("failed preflight modified source")
	}

	var count int64

	f.db.Model(&model.RouterMigration{}).Count(&count)

	if count != 0 {
		t.Fatal("failed preflight queued changes")
	}
}

func TestRouterMigrationCancellationCompensatesBeforeSwitch(t *testing.T) {
	f, peer, admin, billing, migrations, target := migrationFixture(t)

	job, err := migrations.Begin(admin, *peer.ResellerID, "cancel", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "test compensation"})

	if err != nil {
		t.Fatal(err)
	}

	// Source is disabled, but the response is lost before routing metadata switches.
	f.mockA.FailAfterNext(http.MethodPatch, "/interface/wireguard/peers/"+peer.PeerID, 1)

	if err = migrations.Reconcile(job.ID); err == nil {
		t.Fatal("expected ambiguous source response")
	}

	if err = migrations.Cancel(admin, *peer.ResellerID, job.ID, "keep original router"); err != nil {
		t.Fatal(err)
	}

	if err = service.NewRouterMigrations(billing, f.peers).Reconcile(job.ID); err != nil {
		t.Fatal(err)
	}

	var current model.Peer

	f.db.First(&current, peer.ID)

	if current.ServerID != peer.ServerID || current.MigrationID != nil {
		t.Fatal("cancellation changed original ownership/routing")
	}

	source, _ := f.adaptor.ForServer(peer.ServerID)

	remote, err := source.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil || remote.Disabled != "false" {
		t.Fatal("eligible source was not restored")
	}

	dest, _ := f.adaptor.ForServer(target.ServerID)

	rows, _ := dest.FetchWgPeers(context.Background())

	queues, _ := dest.FetchSimpleQueues(context.Background())

	schedules, _ := dest.FetchSchedulers(context.Background())

	if len(rows)+len(queues)+len(schedules) != 0 {
		t.Fatal("prepared destination objects were not compensated")
	}

	history, err := migrations.List(admin, *peer.ResellerID)

	if err != nil || history[0].State != "cancelled" {
		t.Fatal("cancellation history missing", err)
	}

	account, _ := billing.Account(admin, *peer.ResellerID)

	if account.BalanceRial != 10_000_000 {
		t.Fatal("compensation modified billing")
	}
}

func TestRouterMigrationCannotCancelAfterRoutingSwitch(t *testing.T) {
	f, peer, admin, _, migrations, target := migrationFixture(t)

	job, err := migrations.Begin(admin, *peer.ResellerID, "no-cancel", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "forward recovery only"})

	if err != nil {
		t.Fatal(err)
	}

	f.mockA.FailNext(http.MethodDelete, "/interface/wireguard/peers/"+peer.PeerID, 1)

	if err = migrations.Reconcile(job.ID); err == nil {
		t.Fatal("expected cleanup failure")
	}

	if err = migrations.Cancel(admin, *peer.ResellerID, job.ID, "unsafe rollback"); !errors.Is(err, service.ErrMigrationConflict) {
		t.Fatal("unsafe rollback permitted", err)
	}
}

func TestRouterMigrationPreservesPresharedKeyAndActualQueueLimit(t *testing.T) {
	f, peer, admin, _, migrations, target := migrationFixture(t)

	source, _ := f.adaptor.ForServer(peer.ServerID)

	key := peer.PrivateKey

	if _, err := source.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{PresharedKey: &key}); err != nil {
		t.Fatal(err)
	}

	limit := "15M/7M"

	if _, err := source.UpdateSimpleQueue(context.Background(), *peer.QueueID, mikrotik.Queue{MaxLimit: &limit}); err != nil {
		t.Fatal(err)
	}

	job, err := migrations.Begin(admin, *peer.ResellerID, "preserve-key", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "preserve legacy router configuration"})

	if err != nil {
		t.Fatal(err)
	}

	if err = migrations.Reconcile(job.ID); err != nil {
		t.Fatal(err)
	}

	f.db.First(&peer, peer.ID)

	dest, _ := f.adaptor.ForServer(peer.ServerID)

	remote, err := dest.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil || remote.PresharedKey == nil || *remote.PresharedKey != key {
		t.Fatal("preshared key was not preserved")
	}

	queue, err := dest.FetchSimpleQueue(context.Background(), *peer.QueueID)

	if err != nil || queue.MaxLimit == nil || *queue.MaxLimit != limit {
		t.Fatal("actual queue limit was not preserved")
	}

	path, _, err := service.NewConfigGenerator(f.db).GetPeerConfig(peer.ID)

	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "PresharedKey = "+key) {
		t.Fatal("client configuration omitted the preshared key")
	}

	info, err := os.Stat(path)

	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("private config is not stored with owner-only permissions")
	}

	qr, err := service.NewQRCodeGenerator(f.db).GetPeerQRCode(peer.ID)

	if err != nil {
		t.Fatal(err)
	}

	info, err = os.Stat(qr)

	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("QR code is not stored privately")
	}
}

func TestRouterMigrationConcurrentWorkersAndExplicitPublicEndpoint(t *testing.T) {
	f, peer, admin, billing, migrations, target := migrationFixture(t)

	job, err := migrations.Begin(admin, *peer.ResellerID, "concurrent", service.RouterMigrationRequest{InterfaceID: target.ID, Endpoint: "vpn-new.example.test", Reason: "public endpoint differs from management address"})

	if err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 2)

	start := make(chan struct{})

	for i := 0; i < 2; i++ {
		go func() { <-start; results <- service.NewRouterMigrations(billing, f.peers).Reconcile(job.ID) }()
	}

	close(start)

	for i := 0; i < 2; i++ {
		if err = <-results; err != nil {
			t.Fatal(err)
		}
	}

	f.db.First(&peer, peer.ID)

	if peer.Endpoint != "vpn-new.example.test" {
		t.Fatal("public endpoint not applied")
	}

	router, _ := f.adaptor.ForServer(target.ServerID)

	rows, err := router.FetchWgPeers(context.Background())

	if err != nil || len(rows) != 1 {
		t.Fatal("concurrent workers duplicated objects")
	}
}

func TestRouterMigrationContinuesTrafficAndQuotaEnforcementWhilePreparing(t *testing.T) {
	f, peer, admin, _, migrations, target := migrationFixture(t)

	job, err := migrations.Begin(admin, *peer.ResellerID, "traffic", service.RouterMigrationRequest{InterfaceID: target.ID, Reason: "keep enforcing active source"})

	if err != nil {
		t.Fatal(err)
	}

	source, _ := f.adaptor.ForServer(peer.ServerID)

	usage := *peer.TrafficLimit + 1

	if _, err = source.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{TransferTx: strconv.FormatInt(usage, 10)}); err != nil {
		t.Fatal(err)
	}

	f.calculator.CalculatePeerTraffic()

	f.db.First(&peer, peer.ID)

	if peer.DownloadUsage != usage || !peer.QuotaBlocked {
		t.Fatal("migration paused source usage or quota enforcement")
	}

	live, err := source.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil || live.Disabled != "true" {
		t.Fatal("quota-exceeded source remained enabled")
	}

	if err = migrations.Reconcile(job.ID); err != nil {
		t.Fatal(err)
	}

	f.db.First(&peer, peer.ID)

	dest, _ := f.adaptor.ForServer(peer.ServerID)

	live, err = dest.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil || live.Disabled != "true" || peer.DownloadUsage != usage {
		t.Fatal("migration lost quota restriction or consumption")
	}
}
