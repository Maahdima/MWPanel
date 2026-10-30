package http

import (
	"fmt"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
	"net/http"
	"testing"
)

func TestResellerRetryHTTPEnforcesRoleOwnershipAndLease(t *testing.T) {
	app, db, auth, tenant := authorizationAPI(t)

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	if err := service.NewBilling(db).CreateAccount(admin, tenant.ID, model.BillingPrepaid); err != nil {
		t.Fatal(err)
	}

	peer := model.Peer{ResellerID: &tenant.ID, UUID: "retry-peer", PeerID: "pending:retry", PublicKey: "retry-public", AllowedAddress: "10.0.0.2/32", ProvisioningBlocked: true}

	opID := "retry-operation"

	peer.PurchaseID = &opID

	if err := db.Create(&peer).Error; err != nil {
		t.Fatal(err)
	}

	op := model.PeerPurchase{ID: opID, ResellerID: tenant.ID, PeerID: peer.ID, State: "pending", Kind: "create", IdempotencyKey: "retry", RequestHash: "retry", LastError: "interface: verification failed", NextAttemptAt: 9999999999}

	if err := db.Create(&op).Error; err != nil {
		t.Fatal(err)
	}

	adminToken, err := auth.Login("admin", "test-password")

	if err != nil {
		t.Fatal(err)
	}

	resellerToken, err := auth.Login("reseller", "test-password")

	if err != nil {
		t.Fatal(err)
	}

	path := fmt.Sprintf("/api/reseller/%d/peers/%d/retry", tenant.ID, peer.ID)

	if rec := authRequest(app, http.MethodPost, path, "", `{}`); rec.Code != http.StatusUnauthorized {
		t.Fatal(rec.Code)
	}

	if rec := authRequest(app, http.MethodPost, path, resellerToken.AccessToken, `{}`); rec.Code != http.StatusForbidden {
		t.Fatal(rec.Code)
	}

	wrong := fmt.Sprintf("/api/reseller/%d/peers/%d/retry", tenant.ID, peer.ID+100)

	if rec := authRequest(app, http.MethodPost, wrong, adminToken.AccessToken, `{}`); rec.Code != http.StatusNotFound {
		t.Fatal("ownership lookup", rec.Code)
	}

	if rec := authRequest(app, http.MethodPost, path, adminToken.AccessToken, `{}`); rec.Code != http.StatusAccepted {
		t.Fatal(rec.Code, rec.Body.String())
	}

	db.First(&op, "id = ?", opID)

	if op.NextAttemptAt != 0 || op.LastError == "" || op.State != "pending" {
		t.Fatal("retry changed verification outcome")
	}
}
