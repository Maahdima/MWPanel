// This command starts an explicitly selected, disposable development router.
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/maahdima/mwp/api/internal/mockrouteros"
)

func main() {
	address := flag.String("listen", "127.0.0.1:18080", "loopback listen address")
	name := flag.String("name", "MWPanel local mock", "router identity")
	username := flag.String("username", "mockadmin", "mock REST username")
	flag.Parse()
	password := os.Getenv("MOCK_ROUTEROS_PASSWORD")
	if password == "" {
		log.Fatal("set MOCK_ROUTEROS_PASSWORD")
	}
	host, _, err := net.SplitHostPort(*address)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		log.Fatal("mock listen address must use a loopback IP")
	}
	server := &http.Server{Addr: *address, Handler: mockrouteros.New(*name, *username, password), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("disposable Mock RouterOS on %s", *address)
	log.Fatal(server.ListenAndServe())
}
