package wg

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/curve25519"
)

// isDryRun checks at runtime whether WG commands should be skipped.
func isDryRun() bool {
	return os.Getenv("WG_DRY_RUN") == "true"
}

func GenerateKeyPair() (privateKey, publicKey string, err error) {
	var priv, pub [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return "", "", err
	}
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	pubBytes, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return "", "", err
	}
	copy(pub[:], pubBytes)

	return base64.StdEncoding.EncodeToString(priv[:]),
		base64.StdEncoding.EncodeToString(pub[:]), nil
}

func GeneratePresharedKey() (string, error) {
	var key [32]byte
	if _, err := rand.Read(key[:]); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key[:]), nil
}

func AddPeer(iface, publicKey, presharedKey, allowedIP string) error {
	if isDryRun() {
		log.Printf("[wg-dry] add peer %s allowed-ips %s on %s", publicKey[:8]+"...", allowedIP, iface)
		return nil
	}
	cmd := exec.Command("wg", "set", iface,
		"peer", publicKey,
		"preshared-key", "/dev/stdin",
		"allowed-ips", allowedIP,
	)
	cmd.Stdin = strings.NewReader(presharedKey)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	return nil
}

func RemovePeer(iface, publicKey string) error {
	if isDryRun() {
		log.Printf("[wg-dry] remove peer %s", publicKey[:8]+"...")
		return nil
	}
	cmd := exec.Command("wg", "set", iface, "peer", publicKey, "remove")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	return nil
}
