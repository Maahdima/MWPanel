package http

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
)

func TestBillingHTTPRejectsCrossTenantFilesAndFinancialBypasses(t *testing.T) {
	app, db, auth, tenant := authorizationAPI(t)
	adminLogin, err := auth.Login("admin", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	member, err := auth.Login("reseller", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	actor := service.Principal{AccountID: adminLogin.UserID, Role: model.RoleSuperadmin}
	billing := service.NewBilling(db)
	if err := billing.CreateAccount(actor, tenant.ID, model.BillingPrepaid); err != nil {
		t.Fatal(err)
	}
	other := model.Reseller{Name: "other"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	var root model.Admin
	db.Where("username = ?", "admin").First(&root)
	if err := db.Create(&model.Admin{Username: "other", Password: root.Password, Role: model.RoleReseller, ResellerID: &other.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := billing.CreateAccount(actor, other.ID, model.BillingPrepaid); err != nil {
		t.Fatal(err)
	}
	second, err := auth.Login("other", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	var imageData bytes.Buffer
	png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("amount_rial", "123456789012345")
	writer.WriteField("paid_at", time.Now().UTC().Format(time.RFC3339))
	writer.WriteField("transaction_reference", "http-bank-reference")
	part, err := writer.CreateFormFile("image", "receipt.png")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(imageData.Bytes())
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/billing/me/receipts?reseller_id="+strconv.Itoa(int(other.ID)), &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+member.AccessToken)
	req.Header.Set("Idempotency-Key", "http-upload")
	result := httptest.NewRecorder()
	app.ServeHTTP(result, req)
	if result.Code != 201 {
		t.Fatalf("upload: %d %s", result.Code, result.Body.String())
	}
	var receipt struct {
		Data service.ReceiptView `json:"data"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Data.ResellerID != tenant.ID {
		t.Fatal("forged reseller_id selected receipt owner")
	}
	account, _ := billing.Account(actor, tenant.ID)
	if account.BalanceRial != 0 {
		t.Fatal("HTTP upload credited unapproved payment")
	}
	imagePath := "/api/billing/me/receipts/" + strconv.Itoa(int(receipt.Data.ID)) + "/image"
	if rec := authRequest(app, http.MethodGet, imagePath, second.AccessToken, ""); rec.Code != 404 {
		t.Fatalf("receipt IDOR: %d", rec.Code)
	}
	if rec := authRequest(app, http.MethodGet, imagePath, "", ""); rec.Code != 401 {
		t.Fatalf("unprotected image: %d", rec.Code)
	}
	if rec := authRequest(app, http.MethodGet, imagePath, member.AccessToken, ""); rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("protected image headers/status: %d", rec.Code)
	}
	approve := "/api/billing/receipts/" + strconv.Itoa(int(receipt.Data.ID)) + "/approve"
	if rec := authRequest(app, http.MethodPost, approve, member.AccessToken, `{}`); rec.Code != 403 {
		t.Fatal("reseller approved payment")
	}
	for i := 0; i < 2; i++ {
		if rec := authRequest(app, http.MethodPost, approve, adminLogin.AccessToken, `{}`); rec.Code != 200 {
			t.Fatalf("approval: %d %s", rec.Code, rec.Body.String())
		}
	}
	account, _ = billing.Account(actor, tenant.ID)
	if account.BalanceRial != 123456789012345 {
		t.Fatalf("payment duplicated or lost precision: %d", account.BalanceRial)
	}
	rec := authRequest(app, http.MethodGet, "/api/billing/me?reseller_id="+strconv.Itoa(int(other.ID)), member.AccessToken, "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"balance_rial":"123456789012345"`) {
		t.Fatalf("own book/scoping: %d %s", rec.Code, rec.Body.String())
	}
	if rec := authRequest(app, http.MethodPut, "/api/billing/instructions", member.AccessToken, `{"rate_rial":"0"}`); rec.Code != 403 {
		t.Fatal("reseller edited payment settings")
	}
}

func TestReadonlyResellerHTTPScopesPeersAndSessions(t *testing.T) {
	app, db, auth, tenant := authorizationAPI(t)
	member, err := auth.Login("reseller", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	own := model.Peer{ResellerID: &tenant.ID, UUID: "own-uuid", PeerID: "*1", Name: "owned", AllowedAddress: "10.0.0.1"}
	foreign := model.Peer{UUID: "foreign-uuid", PeerID: "*2", Name: "foreign", AllowedAddress: "10.0.0.2"}
	if err := db.Create(&own).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&foreign).Error; err != nil {
		t.Fatal(err)
	}
	rec := authRequest(app, http.MethodGet, "/api/reseller/me/peers", member.AccessToken, "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "own-uuid") || strings.Contains(rec.Body.String(), "foreign-uuid") {
		t.Fatalf("peer scope: %d %s", rec.Code, rec.Body.String())
	}
	for _, suffix := range []string{"config", "qrcode", "sessions"} {
		path := "/api/reseller/me/peers/" + strconv.Itoa(int(foreign.ID)) + "/" + suffix
		if rec := authRequest(app, http.MethodGet, path, member.AccessToken, ""); rec.Code != 404 {
			t.Fatalf("%s IDOR: %d", suffix, rec.Code)
		}
	}
	if rec := authRequest(app, http.MethodGet, "/api/reseller/me/peers/"+strconv.Itoa(int(own.ID))+"/sessions", member.AccessToken, ""); rec.Code != 200 {
		t.Fatalf("own session read: %d", rec.Code)
	}
}
