package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	jobs "github.com/maahdima/mwp/api/cmd/jobs"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
	"github.com/maahdima/mwp/api/utils/validate"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func authorizationAPI(t *testing.T) (*echo.Echo, *gorm.DB, *service.Authentication, model.Reseller) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "auth.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := dataservice.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	tenant := model.Reseller{Name: "tenant"}
	if err := db.Create(&tenant).Error; err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	for _, account := range []model.Admin{{Username: "admin", Password: string(hash), Role: model.RoleSuperadmin}, {Username: "reseller", Password: string(hash), Role: model.RoleReseller, ResellerID: &tenant.ID}} {
		if err := db.Create(&account).Error; err != nil {
			t.Fatal(err)
		}
	}
	registry := common.NewMwpClient(db)
	adaptor := mikrotik.NewAdaptor(registry)
	auth := service.NewAuthentication(db)
	scheduler := service.NewScheduler(adaptor)
	queue := service.NewQueue(adaptor)
	files := service.NewConfigGenerator(db)
	qr := service.NewQRCodeGenerator(db)
	excel := service.NewExcelGenerator(db)
	peers := service.NewWGPeer(db, adaptor, scheduler, queue, files, qr)
	servers := service.NewServerService(db, registry, adaptor)
	interfaces := service.NewWgInterface(db, adaptor, peers)
	pools := service.NewIPPool(db, adaptor)
	device := service.NewDeviceData(db, adaptor, servers, interfaces, peers)
	calculator := jobs.NewTrafficCalculator(db, adaptor, nil)
	syncer := service.NewSyncService(db, adaptor, files, qr, peers, interfaces)
	app := echo.New()
	app.Validator = &validate.CustomValidator{Validator: validator.New()}
	SetupMwpAPI(app, registry, auth, servers, interfaces, pools, peers, files, qr, excel, device, calculator, syncer, nil, service.NewResellers(db, peers), service.NewBilling(db), service.NewPayments(service.NewBilling(db), t.TempDir()))
	return app, db, auth, tenant
}

func authRequest(app *echo.Echo, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}

func TestLegacyAdministrativeRoutesRejectResellers(t *testing.T) {
	app, _, auth, _ := authorizationAPI(t)
	login, err := auth.Login("reseller", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	tested := 0
	for _, route := range app.Routes() {
		if strings.HasPrefix(route.Path, "/api/reseller/me/") || strings.HasPrefix(route.Path, "/api/billing/me") || (route.Path == "/api/billing/instructions" && route.Method == http.MethodGet) {
			continue
		}
		if !strings.HasPrefix(route.Path, "/api/") || strings.HasPrefix(route.Path, "/api/auth/") || strings.HasPrefix(route.Path, "/api/user/") || strings.HasPrefix(route.Path, "/api/service/") || route.Path == "/api/telegram/status" || route.Path == "/api/bale/webhook" || route.Method == "echo_route_not_found" {
			continue
		}
		path := strings.ReplaceAll(route.Path, ":id", "1")
		t.Run(route.Method+" "+path, func(t *testing.T) {
			rec := authRequest(app, route.Method, path, login.AccessToken, `{"reseller_id":2,"disabled":false,"traffic_limit":"1000"}`)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
			}
			unauth := authRequest(app, route.Method, path, "", `{}`)
			if unauth.Code != http.StatusUnauthorized {
				t.Fatalf("missing token: %d", unauth.Code)
			}
		})
		tested++
	}
	if tested < 30 {
		t.Fatalf("insufficient route coverage: %d", tested)
	}
}

func TestRoleRevocationAndIdentityAreDatabaseBacked(t *testing.T) {
	app, db, auth, tenant := authorizationAPI(t)
	login, err := auth.Login("admin", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	rec := authRequest(app, http.MethodGet, "/api/auth/me", login.AccessToken, "")
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	if err := db.Model(&model.Admin{}).Where("username = ?", "admin").Updates(map[string]any{"role": model.RoleReseller, "reseller_id": tenant.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if rec := authRequest(app, http.MethodGet, "/api/server", login.AccessToken, ""); rec.Code != 403 {
		t.Fatalf("stale token kept admin role: %d %s", rec.Code, rec.Body.String())
	}
	forged, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "reseller", "exp": time.Now().Add(time.Hour).Unix(), "role": "superadmin", "reseller_id": 999}).SignedString(auth.AccessSecret)
	if err != nil {
		t.Fatal(err)
	}
	if rec := authRequest(app, http.MethodGet, "/api/server", forged, ""); rec.Code != 403 {
		t.Fatalf("trusted role claims: %d", rec.Code)
	}
	if err := db.Model(&model.Admin{}).Where("username = ?", "admin").Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	if rec := authRequest(app, http.MethodGet, "/api/auth/me", login.AccessToken, ""); rec.Code != 401 {
		t.Fatalf("disabled account token accepted: %d", rec.Code)
	}
	if _, err := auth.Login("admin", "test-password"); err == nil {
		t.Fatal("disabled account logged in")
	}
}

func TestResellerIdentityAndProfileCannotTargetAnotherUser(t *testing.T) {
	app, _, auth, _ := authorizationAPI(t)
	login, err := auth.Login("reseller", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	rec := authRequest(app, http.MethodGet, "/api/auth/me", login.AccessToken, "")
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var result struct {
		Data service.Principal `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.Role != model.RoleReseller || result.Data.ResellerID == nil {
		t.Fatalf("missing role: %s", rec.Body.String())
	}
	rec = authRequest(app, http.MethodPut, "/api/auth/profile", login.AccessToken, `{"old_username":"admin","old_password":"test-password","new_username":"hijacked"}`)
	if rec.Code != 403 {
		t.Fatalf("profile identity bypass: %d %s", rec.Code, rec.Body.String())
	}
}

func TestInvalidRolesAndUnassignedResellerCannotLogin(t *testing.T) {
	_, db, auth, tenant := authorizationAPI(t)
	if err := db.Model(&model.Admin{}).Where("username = ?", "reseller").Update("reseller_id", nil).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Login("reseller", "test-password"); err == nil {
		t.Fatal("unassigned reseller logged in")
	}
	if err := db.Model(&model.Admin{}).Where("username = ?", "reseller").Updates(map[string]any{"reseller_id": tenant.ID, "role": "unknown"}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Login("reseller", "test-password"); err == nil {
		t.Fatal("unknown role logged in")
	}
}
