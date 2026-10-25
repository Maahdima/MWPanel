// Package mockrouteros supplies a configurable RouterOS REST test double. It is
// used only by tests and the explicitly launched local development mock command.
package mockrouteros

import (
	"encoding/json"
	"fmt"
	"github.com/maahdima/mwp/api/utils/wireguard"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Router struct {
	mu                 sync.Mutex
	Username, Password string
	Name               string
	objects            map[string]map[string]map[string]any
	sequence           map[string]int
	failures           map[string]int
	afterFailures      map[string]int
}

func New(name, username, password string) *Router {
	r := &Router{Name: name, Username: username, Password: password, objects: make(map[string]map[string]map[string]any), sequence: make(map[string]int), failures: make(map[string]int), afterFailures: make(map[string]int)}
	for _, path := range []string{"/interface/wireguard/peers", "/interface/wireguard", "/queue/simple", "/system/scheduler", "/ip/address", "/interface"} {
		r.objects[path] = make(map[string]map[string]any)
	}
	return r
}

// FailNext injects HTTP 503 responses for a precise REST method/path.
func (r *Router) FailNext(method, path string, count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failures[method+" "+path] = count
}

// FailAfterNext simulates a lost/failed response after a mutation took effect.
func (r *Router) FailAfterNext(method, path string, count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.afterFailures[method+" "+path] = count
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	user, password, ok := req.BasicAuth()
	if !ok || user != r.Username || password != r.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/rest")
	key := req.Method + " " + path
	if r.failures[key] > 0 {
		r.failures[key]--
		http.Error(w, "injected router failure", 503)
		return
	}
	if req.Method == http.MethodGet {
		switch path {
		case "/system/identity":
			json.NewEncoder(w).Encode(map[string]string{"name": r.Name})
			return
		case "/system/resource":
			json.NewEncoder(w).Encode(map[string]string{"uptime": "1d", "version": "7.20 (mock)", "cpu-load": "0", "free-memory": "268435456", "total-memory": "536870912"})
			return
		case "/ip/dns":
			json.NewEncoder(w).Encode(map[string]string{"servers": "1.1.1.1"})
			return
		}
	}
	collection := path
	id := ""
	if _, ok := r.objects[collection]; !ok {
		at := strings.LastIndex(path, "/")
		if at >= 0 {
			collection, id = path[:at], path[at+1:]
		}
	}
	table, ok := r.objects[collection]
	if !ok {
		http.NotFound(w, req)
		return
	}
	switch req.Method {
	case http.MethodGet:
		if id != "" {
			value, ok := table[id]
			if !ok {
				http.NotFound(w, req)
				return
			}
			json.NewEncoder(w).Encode(value)
			return
		}
		list := make([]map[string]any, 0, len(table))
		for _, value := range table {
			list = append(list, value)
		}
		json.NewEncoder(w).Encode(list)
	case http.MethodPut, http.MethodPatch:
		var input map[string]any
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20)).Decode(&input); err != nil || input == nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		if req.Method == http.MethodPut {
			if id != "" {
				http.Error(w, "create requires a collection", 400)
				return
			}
			r.sequence[collection]++
			id = "*" + strconv.Itoa(r.sequence[collection])
			table[id] = map[string]any{".id": id, "disabled": "false"}
			if collection == "/interface/wireguard" {
				key, private, err := wireguard.GeneratePrivateKey()
				if err != nil {
					http.Error(w, "key generation failed", 500)
					return
				}
				public, err := wireguard.GeneratePublicKey(key)
				if err != nil {
					http.Error(w, "key generation failed", 500)
					return
				}
				table[id]["private-key"] = private
				table[id]["public-key"] = public
				table[id]["running"] = "true"
				table[id]["mtu"] = "1420"
				table[id]["rx-byte"] = "0"
				table[id]["tx-byte"] = "0"
				r.objects["/interface"][id] = table[id]
			}
			if collection == "/interface/wireguard/peers" {
				table[id]["rx"] = "0"
				table[id]["tx"] = "0"
				table[id]["last-handshake"] = "0s"
			}
		} else if _, ok := table[id]; id == "" || !ok {
			http.NotFound(w, req)
			return
		}
		for key, value := range input {
			if key != ".id" {
				table[id][key] = value
			}
		}
		if r.afterFailures[key] > 0 {
			r.afterFailures[key]--
			http.Error(w, "injected ambiguous response", 503)
			return
		}
		json.NewEncoder(w).Encode(table[id])
	case http.MethodDelete:
		if _, ok := table[id]; id == "" || !ok {
			http.NotFound(w, req)
			return
		}
		delete(table, id)
		if collection == "/interface/wireguard" {
			delete(r.objects["/interface"], id)
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PUT, PATCH, DELETE")
		http.Error(w, fmt.Sprintf("unsupported method %s", req.Method), 405)
	}
}
