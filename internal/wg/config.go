package wg

import (
	"bytes"
	"fmt"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/imzami/routevpn-core/internal/config"
	"github.com/imzami/routevpn-core/internal/models"
)

func GenerateClientConfig(peer *models.VPNPeer, cfg *config.Config) string {
	return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s
Jc = %s
Jmin = %s
Jmax = %s
S1 = %s
S2 = %s
H1 = %s
H2 = %s
H3 = %s
H4 = %s

[Peer]
PublicKey = %s
PresharedKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`,
		peer.PrivateKey,
		peer.AssignedIP,
		cfg.WGDNS,
		cfg.AWGJc, cfg.AWGJmin, cfg.AWGJmax,
		cfg.AWGS1, cfg.AWGS2,
		cfg.AWGH1, cfg.AWGH2, cfg.AWGH3, cfg.AWGH4,
		cfg.ServerPublicKey,
		peer.PresharedKey,
		cfg.WGEndpoint,
	)
}

func GenerateQRCode(content string, size int) ([]byte, error) {
	qrCode, err := qr.Encode(content, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}
	qrCode, err = barcode.Scale(qrCode, size, size)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, qrCode); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
