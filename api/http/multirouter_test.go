package http

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/internal/mockrouteros"
)

func TestHTTPUsesPeerRouterAndPreservesAuthenticationGuards(t *testing.T) {
	app, db, auth, _ := authorizationAPI(t)
	login, err := auth.Login("admin", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	createServer := func(name, hostOverride string) uint {
		mock := mockrouteros.New(name, "u", "p")
		server := httptest.NewServer(mock)
		t.Cleanup(server.Close)
		parsed, _ := url.Parse(server.URL)
		host, port, _ := net.SplitHostPort(parsed.Host)
		if hostOverride != "" {
			host = hostOverride
		}
		input, _ := json.Marshal(map[string]any{"name": name, "ip_address": host, "api_port": port, "username": "u", "password": "p", "is_ssl": false})
		rec := authRequest(app, http.MethodPost, "/api/server", login.AccessToken, string(input))
		if rec.Code != 200 {
			t.Fatalf("server create: %d %s", rec.Code, rec.Body.String())
		}
		var body struct {
			Data struct {
				ID uint `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.Data.ID
	}
	a := createServer("a", "")
	b := createServer("b", "localhost")
	createInterface := func(serverID uint) uint {
		rec := authRequest(app, http.MethodPost, "/api/interface", login.AccessToken, fmt.Sprintf(`{"server_id":%d,"name":"wg0","listen_port":"51820"}`, serverID))
		if rec.Code != http.StatusCreated {
			t.Fatalf("interface create: %d %s", rec.Code, rec.Body.String())
		}
		var body struct {
			Data struct {
				ID uint `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.Data.ID
	}
	ia, ib := createInterface(a), createInterface(b)
	createPeer := func(iface uint, name string) uint {
		rec := authRequest(app, http.MethodGet, "/api/peer/credentials", login.AccessToken, "")
		if rec.Code != 200 {
			t.Fatal(rec.Body.String())
		}
		var credentials struct {
			Data map[string]string `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &credentials); err != nil {
			t.Fatal(err)
		}
		input := map[string]any{"interface_id": iface, "name": name, "private_key": credentials.Data["private_key"], "public_key": credentials.Data["public_key"], "allowed_address": "10.0.0.2/32", "endpoint": "vpn.example.test", "traffic_limit": "10", "expire_time": time.Now().AddDate(0, 0, 30).Format("2006-01-02")}
		encoded, _ := json.Marshal(input)
		rec = authRequest(app, http.MethodPost, "/api/peer", login.AccessToken, string(encoded))
		if rec.Code != 201 {
			t.Fatalf("peer body was lost or routed incorrectly: %d %s", rec.Code, rec.Body.String())
		}
		var body struct {
			Data struct {
				ID uint `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.Data.ID
	}
	pa, pb := createPeer(ia, "a"), createPeer(ib, "b")
	if err := db.Model(&model.Server{}).Where("id = ?", a).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	// A forged query selector cannot redirect a request away from the persisted peer router.
	rec := authRequest(app, http.MethodPatch, fmt.Sprintf("/api/peer/%d/status?server_id=%d", pb, a), login.AccessToken, "")
	if rec.Code != 200 {
		t.Fatalf("healthy peer blocked by inactive default: %d %s", rec.Code, rec.Body.String())
	}
	var first, second model.Peer
	db.First(&first, pa)
	db.First(&second, pb)
	if first.Disabled || !second.Disabled {
		t.Fatal("HTTP operation changed another router's peer")
	}
	rec = authRequest(app, http.MethodPatch, "/api/peer/"+strconv.Itoa(int(pa))+"/status", login.AccessToken, "")
	if rec.Code != 503 {
		t.Fatalf("inactive target did not retain connection guard: %d", rec.Code)
	}
	rec = authRequest(app, http.MethodGet, "/api/peer", login.AccessToken, "")
	if rec.Code != 200 {
		t.Fatalf("healthy router's peers inaccessible: %d", rec.Code)
	}
	if rec := authRequest(app, http.MethodGet, "/api/peer", "", ""); rec.Code != 401 {
		t.Fatalf("JWT guard bypassed: %d", rec.Code)
	}
	reseller, err := auth.Login("reseller", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	if rec := authRequest(app, http.MethodPatch, fmt.Sprintf("/api/peer/%d/status", pb), reseller.AccessToken, ""); rec.Code != 403 {
		t.Fatalf("role guard bypassed: %d", rec.Code)
	}
}
