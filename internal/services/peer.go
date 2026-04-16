package services

import (
	"fmt"
	"net"
	"time"

	"github.com/imzami/routevpn-core/internal/models"
	"github.com/imzami/routevpn-core/internal/wg"
	"github.com/jmoiron/sqlx"
)

type CreatePeerInput struct {
	WhatsappNumber string
	PackageID      int64
	ResellerID     int64
}

func CreatePeer(svc *Container, input CreatePeerInput) (*models.VPNPeer, error) {
	tx, err := svc.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var pkg models.Package
	if err := tx.Get(&pkg, "SELECT * FROM packages WHERE id=$1", input.PackageID); err != nil {
		return nil, fmt.Errorf("package not found")
	}

	var reseller models.User
	if err := tx.Get(&reseller, "SELECT * FROM users WHERE id=$1 FOR UPDATE", input.ResellerID); err != nil {
		return nil, fmt.Errorf("reseller not found")
	}
	if reseller.Balance.LessThan(pkg.PriceBDT) {
		return nil, fmt.Errorf("insufficient balance: have ৳%s, need ৳%s", reseller.Balance.StringFixed(2), pkg.PriceBDT.StringFixed(2))
	}

	newBalance := reseller.Balance.Sub(pkg.PriceBDT)
	if _, err := tx.Exec("UPDATE users SET balance=$1 WHERE id=$2", newBalance, input.ResellerID); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(
		"INSERT INTO transactions (reseller_id, amount, type, note) VALUES ($1, $2, 'debit', $3)",
		input.ResellerID, pkg.PriceBDT, fmt.Sprintf("Peer creation: %s for %s", pkg.Name, input.WhatsappNumber),
	); err != nil {
		return nil, err
	}

	privKey, pubKey, err := wg.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("keygen failed: %w", err)
	}
	psk, err := wg.GeneratePresharedKey()
	if err != nil {
		return nil, fmt.Errorf("psk failed: %w", err)
	}

	assignedIP, err := allocateIP(tx, svc.Cfg.WGSubnet)
	if err != nil {
		return nil, fmt.Errorf("IP allocation failed: %w", err)
	}

	expiry := time.Now().AddDate(0, 0, pkg.DurationDays)

	peer := &models.VPNPeer{
		UserID:         input.ResellerID,
		WhatsappNumber: input.WhatsappNumber,
		PublicKey:      pubKey,
		PrivateKey:     privKey,
		PresharedKey:   psk,
		AssignedIP:     assignedIP,
		ExpiryDate:     expiry,
		Status:         "active",
		PackageID:      input.PackageID,
	}

	err = tx.QueryRowx(
		`INSERT INTO vpn_peers (user_id, whatsapp_number, public_key, private_key, preshared_key, assigned_ip, expiry_date, status, package_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at`,
		peer.UserID, peer.WhatsappNumber, peer.PublicKey, peer.PrivateKey, peer.PresharedKey,
		peer.AssignedIP, peer.ExpiryDate, peer.Status, peer.PackageID,
	).Scan(&peer.ID, &peer.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := wg.AddPeer(svc.Cfg.WGInterface, pubKey, psk, assignedIP); err != nil {
		return nil, fmt.Errorf("wg add peer failed: %w", err)
	}

	return peer, tx.Commit()
}

func allocateIP(tx *sqlx.Tx, subnet string) (string, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", err
	}

	var usedIPs []string
	_ = tx.Select(&usedIPs, "SELECT assigned_ip FROM vpn_peers")

	used := make(map[string]bool, len(usedIPs))
	for _, u := range usedIPs {
		ip, _, e := net.ParseCIDR(u)
		if e == nil {
			used[ip.String()] = true
		} else {
			used[u] = true
		}
	}

	current := cloneIP(ipNet.IP)
	incIP(current) // skip .0
	incIP(current) // skip .1 (server)

	for ipNet.Contains(current) {
		if !used[current.String()] {
			return current.String() + "/32", nil
		}
		incIP(current)
	}
	return "", fmt.Errorf("no available IPs in subnet")
}

func cloneIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
