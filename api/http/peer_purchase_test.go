package http

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
)

func TestPurchaseHTTPScopesOwnershipAndCannotForgePrice(t *testing.T) {
	app, db, auth, tenant := authorizationAPI(t)
	member, err := auth.Login("reseller", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	root, err := auth.Login("admin", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	billing := service.NewBilling(db)
	admin := service.Principal{AccountID: root.UserID, Role: model.RoleSuperadmin}
	if err = billing.CreateAccount(admin, tenant.ID, model.BillingPrepaid); err != nil {
		t.Fatal(err)
	}
	if err = billing.Adjust(admin, tenant.ID, "credit", 1_000_000, "test"); err != nil {
		t.Fatal(err)
	}
	server := model.Server{Name: "offline", IPAddress: "127.0.0.1", APIPort: 1}
	db.Create(&server)
	iface := model.Interface{ServerID: server.ID, InterfaceID: "*1", Name: "wg0"}
	db.Create(&iface)
	db.Create(&model.ResellerResource{ResellerID: tenant.ID, InterfaceID: iface.ID})
	quota := int64(10 * service.BillingGiB)
	expiry := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	peer := model.Peer{ResellerID: &tenant.ID, ServerID: server.ID, UUID: "owned-purchase", PeerID: "*1", Name: "owned", Interface: "wg0", AllowedAddress: "10.0.0.2/32", TrafficLimit: &quota, ExpireTime: &expiry, RestrictionsInitialized: true}
	db.Create(&peer)
	other := peer
	other.ID = 0
	other.UUID = "foreign-purchase"
	other.ResellerID = nil
	other.PeerID = "*2"
	other.AllowedAddress = "10.0.0.3/32"
	db.Create(&other)
	path := fmt.Sprintf("/api/reseller/me/peers/%d/purchases", peer.ID)
	foreign := fmt.Sprintf("/api/reseller/me/peers/%d/purchases", other.ID)
	if r := authRequest(app, http.MethodGet, foreign, member.AccessToken, ""); r.Code != 404 {
		t.Fatalf("foreign history %d", r.Code)
	}
	// authRequest has no idempotency header: the endpoint must reject, not charge.
	if r := authRequest(app, http.MethodPost, path, member.AccessToken, `{"operation":"add_traffic","bytes":"1073741824"}`); r.Code != 400 {
		t.Fatalf("missing key %d %s", r.Code, r.Body.String())
	}
	// Use an actual request with a forged cheap price and forged ownership.
	req := newPurchaseRequest(path+"?reseller_id=999", member.AccessToken, `{"operation":"add_traffic","bytes":"1073741824","amount_rial":"1","reseller_id":999}`, "purchase")
	result := servePurchase(app, req)
	if result.Code != 202 {
		t.Fatalf("purchase %d %s", result.Code, result.Body.String())
	}
	var charge model.BillingCharge
	db.First(&charge)
	if charge.ResellerID != tenant.ID || charge.AmountRial != 100000 {
		t.Fatal("forged owner or price accepted")
	}
	if r := authRequest(app, http.MethodGet, path, member.AccessToken, ""); r.Code != 200 {
		t.Fatal("own history failed", r.Code)
	}
	result = servePurchase(app, newPurchaseRequest(foreign, member.AccessToken, `{"operation":"add_traffic","bytes":"1073741824"}`, "foreign"))
	if result.Code != 404 {
		t.Fatalf("foreign purchase %d", result.Code)
	}
}

func newPurchaseRequest(path, token, body, key string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", key)
	return req
}
func servePurchase(app *echo.Echo, req *http.Request) *httptest.ResponseRecorder {
	result := httptest.NewRecorder()
	app.ServeHTTP(result, req)
	return result
}
