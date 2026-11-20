package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type PeerDefaultsConfig struct {
	DefaultExpireDays *int
}

const envPeerDefaultExpireDays = "PEER_DEFAULT_EXPIRE_DAYS"

var noPeerDefaultExpire = map[string]struct{}{
	"0":          {},
	"-1":         {},
	"none":       {},
	"never":      {},
	"indefinite": {},
	"off":        {},
	"false":      {},
	"no":         {},
	"unlimited":  {},
}

func GetPeerDefaultsConfig() (PeerDefaultsConfig, error) {
	raw, ok := os.LookupEnv(envPeerDefaultExpireDays)
	return parsePeerDefaultExpireDays(raw, ok)
}

func parsePeerDefaultExpireDays(raw string, defined bool) (PeerDefaultsConfig, error) {
	if !defined {
		days := 30
		return PeerDefaultsConfig{DefaultExpireDays: &days}, nil
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return PeerDefaultsConfig{DefaultExpireDays: nil}, nil
	}

	lower := strings.ToLower(raw)
	if _, indefinite := noPeerDefaultExpire[lower]; indefinite {
		return PeerDefaultsConfig{DefaultExpireDays: nil}, nil
	}

	days, err := strconv.Atoi(raw)
	if err != nil {
		return PeerDefaultsConfig{}, fmt.Errorf("invalid %s %q: use a positive number of days or none/0 for no expiry", envPeerDefaultExpireDays, raw)
	}
	if days <= 0 {
		return PeerDefaultsConfig{DefaultExpireDays: nil}, nil
	}

	return PeerDefaultsConfig{DefaultExpireDays: &days}, nil
}

func (c PeerDefaultsConfig) DefaultExpireDate(now time.Time) *string {
	if c.DefaultExpireDays == nil {
		return nil
	}
	formatted := now.AddDate(0, 0, *c.DefaultExpireDays).Format("2006-01-02")
	return &formatted
}
