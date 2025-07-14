package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

var Template = `[Interface]
PrivateKey = %s
Address = %s
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s:%s
AllowedIPs = %s
PersistentKeepalive = %s`

func GeneratePrivateKey() ([]byte, string, error) {
	var privateKey [32]byte

	_, err := rand.Read(privateKey[:])
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey[:])
	return privateKey[:], privateKeyBase64, nil
}

func GeneratePublicKey(privateKey []byte) (string, error) {
	var privKey [32]byte
	copy(privKey[:], privateKey)

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &privKey)

	pubKeyBase64 := base64.StdEncoding.EncodeToString(pubKey[:])
	return pubKeyBase64, nil
}
