package http

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
)

func TestRouterMigrationHTTPAdminGuardsAndCancellationScope(t *testing.T) {
	app, db, auth, tenant := authorizationAPI(t)
	root, err := auth.Login("admin", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	member, err := auth.Login("reseller", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	billing := service.NewBilling(db)
	actor := service.Principal{AccountID: root.UserID, Role: model.RoleSuperadmin}
	if err = billing.CreateAccount(actor, tenant.ID, model.BillingPrepaid); err != nil {
		t.Fatal(err)
	}
	job := model.RouterMigration{ID: uuid.NewString(), ResellerID: tenant.ID, IdempotencyKey: "test", RequestHash: strings.Repeat("0", 64), DestinationInterfaceID: 1, ActorID: root.UserID, Reason: "test", State: "pending"}
	if err = db.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	// Reseller identity must not authorize migration, even with forged owner fields.
	for _, path := range []string{"/api/reseller/1/migrations", "/api/reseller/1/migrations/" + job.ID + "/cancel"} {
		if result := authRequest(app, http.MethodPost, path, member.AccessToken, `{"reseller_id":1,"interface_id":1,"reason":"forged"}`); result.Code != 403 {
			t.Fatalf("reseller migration access: %d", result.Code)
		}
	}
	if result := authRequest(app, http.MethodGet, "/api/reseller/1/migrations", root.AccessToken, ""); result.Code != 200 {
		t.Fatal("admin history failed", result.Code)
	}
	if result := authRequest(app, http.MethodPost, "/api/reseller/1/migrations/"+job.ID+"/cancel", root.AccessToken, `{"reason":"cancel before switch"}`); result.Code != 202 {
		t.Fatal("admin cancellation failed", result.Code, result.Body.String())
	}
	db.First(&job, "id = ?", job.ID)
	if job.State != "cancelling" {
		t.Fatal("cancellation was not durable")
	}
	if result := authRequest(app, http.MethodPost, "/api/reseller/1/migrations", root.AccessToken, `{"reason":"`+strings.Repeat("x", 70<<10)+`"}`); result.Code != 413 {
		t.Fatal("oversized management request accepted", result.Code)
	}
}
