package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/imzami/routevpn-core/internal/models"
	"github.com/imzami/routevpn-core/internal/services"
	"github.com/imzami/routevpn-core/internal/wg"
	"github.com/shopspring/decimal"
)

type ResellerHandler struct {
	svc *services.Container
}

func NewResellerHandler(svc *services.Container) *ResellerHandler {
	return &ResellerHandler{svc: svc}
}

func (h *ResellerHandler) Dashboard(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	var user models.User
	_ = h.svc.DB.Get(&user, "SELECT * FROM users WHERE id=$1", userID)

	var totalPeers int
	_ = h.svc.DB.Get(&totalPeers, "SELECT COUNT(*) FROM vpn_peers WHERE user_id=$1", userID)

	var activePeers int
	_ = h.svc.DB.Get(&activePeers, "SELECT COUNT(*) FROM vpn_peers WHERE user_id=$1 AND status='active'", userID)

	var totalSpent decimal.Decimal
	_ = h.svc.DB.Get(&totalSpent, "SELECT COALESCE(SUM(amount),0) FROM transactions WHERE reseller_id=$1 AND type='debit'", userID)

	return c.Render("reseller/dashboard", fiber.Map{
		"email":       c.Locals("email"),
		"balance":     user.Balance.StringFixed(2),
		"totalPeers":  totalPeers,
		"activePeers": activePeers,
		"totalSpent":  totalSpent.StringFixed(2),
	}, "layouts/reseller")
}

func (h *ResellerHandler) Peers(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	var user models.User
	_ = h.svc.DB.Get(&user, "SELECT * FROM users WHERE id=$1", userID)

	var peers []models.VPNPeer
	_ = h.svc.DB.Select(&peers, `
		SELECT vp.*, p.name as package_name
		FROM vpn_peers vp
		JOIN packages p ON vp.package_id = p.id
		WHERE vp.user_id=$1 ORDER BY vp.created_at DESC`, userID)

	var packages []models.Package
	_ = h.svc.DB.Select(&packages, "SELECT * FROM packages ORDER BY duration_days ASC")

	return c.Render("reseller/peers", fiber.Map{
		"email":    c.Locals("email"),
		"balance":  user.Balance.StringFixed(2),
		"peers":    peers,
		"packages": packages,
	}, "layouts/reseller")
}

func (h *ResellerHandler) CreatePeer(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	whatsapp := c.FormValue("whatsapp_number")
	pkgID, _ := strconv.ParseInt(c.FormValue("package_id"), 10, 64)

	if whatsapp == "" || pkgID == 0 {
		return c.Status(400).SendString("WhatsApp number and package are required")
	}

	_, err := services.CreatePeer(h.svc, services.CreatePeerInput{
		WhatsappNumber: whatsapp,
		PackageID:      pkgID,
		ResellerID:     userID,
	})
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}

	return c.Redirect().To("/reseller/peers")
}

func (h *ResellerHandler) DownloadConfig(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	peerID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	var peer models.VPNPeer
	err := h.svc.DB.Get(&peer, "SELECT * FROM vpn_peers WHERE id=$1 AND user_id=$2", peerID, userID)
	if err != nil {
		return c.Status(404).SendString("Peer not found")
	}

	conf := wg.GenerateClientConfig(&peer, h.svc.Cfg)

	c.Set("Content-Disposition", "attachment; filename=awg-"+peer.WhatsappNumber+".conf")
	c.Set("Content-Type", "text/plain")
	return c.SendString(conf)
}

func (h *ResellerHandler) QRCode(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	peerID, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	var peer models.VPNPeer
	err := h.svc.DB.Get(&peer, "SELECT * FROM vpn_peers WHERE id=$1 AND user_id=$2", peerID, userID)
	if err != nil {
		return c.Status(404).SendString("Peer not found")
	}

	conf := wg.GenerateClientConfig(&peer, h.svc.Cfg)
	png, err := wg.GenerateQRCode(conf, 512)
	if err != nil {
		return c.Status(500).SendString("QR generation failed")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(png)
}

func (h *ResellerHandler) SearchPeers(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	q := c.Query("q")
	status := c.Query("status")

	query := `
		SELECT vp.*, p.name as package_name
		FROM vpn_peers vp
		JOIN packages p ON vp.package_id = p.id
		WHERE vp.user_id=$1`
	args := []interface{}{userID}
	argIdx := 2

	if q != "" {
		query += ` AND (vp.whatsapp_number ILIKE $` + strconv.Itoa(argIdx) + ` OR vp.public_key ILIKE $` + strconv.Itoa(argIdx) + `)`
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if status == "active" || status == "expired" {
		query += ` AND vp.status = $` + strconv.Itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY vp.created_at DESC LIMIT 100"

	var peers []models.VPNPeer
	_ = h.svc.DB.Select(&peers, query, args...)

	var user models.User
	_ = h.svc.DB.Get(&user, "SELECT * FROM users WHERE id=$1", userID)

	var packages []models.Package
	_ = h.svc.DB.Select(&packages, "SELECT * FROM packages ORDER BY duration_days ASC")

	return c.Render("reseller/peers", fiber.Map{
		"email":    c.Locals("email"),
		"balance":  user.Balance.StringFixed(2),
		"peers":    peers,
		"packages": packages,
		"search":   q,
		"status":   status,
	}, "layouts/reseller")
}

func (h *ResellerHandler) Transactions(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	var user models.User
	_ = h.svc.DB.Get(&user, "SELECT * FROM users WHERE id=$1", userID)

	var txns []models.Transaction
	_ = h.svc.DB.Select(&txns, "SELECT * FROM transactions WHERE reseller_id=$1 ORDER BY created_at DESC LIMIT 100", userID)

	return c.Render("reseller/transactions", fiber.Map{
		"email":        c.Locals("email"),
		"balance":      user.Balance.StringFixed(2),
		"transactions": txns,
	}, "layouts/reseller")
}
