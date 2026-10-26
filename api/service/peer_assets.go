package service

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils/wireguard"
)

func normalizedPresharedKey(value *string) string {
	if value == nil || *value == "" || *value == "none" {
		return ""
	}
	return *value
}
func validWireGuardKey(value string) bool {
	if len(value) != 44 {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	return err == nil && len(decoded) == 32
}
func peerConfigText(peer model.Peer, interfaceKey string) (string, error) {
	config := fmt.Sprintf(wireguard.Template, peer.PrivateKey, peer.AllowedAddress, common.DefaultDns, interfaceKey, peer.Endpoint, peer.EndpointPort, common.AllowedIpsIncludeLocal, peer.PersistentKeepalive)
	if key := normalizedPresharedKey(peer.PresharedKey); key != "" {
		if !validWireGuardKey(key) {
			return "", ErrRouterPeerIdentity
		}
		config = strings.Replace(config, "\nPublicKey = ", "\nPresharedKey = "+key+"\nPublicKey = ", 1)
	}
	return config, nil
}
