package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/imzami/routevpn-core/internal/models"
	"github.com/imzami/routevpn-core/internal/services"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	svc *services.Container
}

func NewAdminHandler(svc *services.Container) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) Dashboard(c fiber.Ctx) error {
	var totalResellers int
	_ = h.svc.DB.Get(&totalResellers, "SELECT COUNT(*) FROM users WHERE role='reseller'")

	var totalPeers int
	_ = h.svc.DB.Get(&totalPeers, "SELECT COUNT(*) FROM vpn_peers")

	var activePeers int
	_ = h.svc.DB.Get(&activePeers, "SELECT COUNT(*) FROM vpn_peers WHERE status='active'")

	var totalRevenue decimal.Decimal
	_ = h.svc.DB.Get(&totalRevenue, "SELECT COALESCE(SUM(amount),0) FROM transactions WHERE type='debit'")

	return c.Render("admin/dashboard", fiber.Map{
		"email":          c.Locals("email"),
		"totalResellers": totalResellers,
		"totalPeers":     totalPeers,
		"activePeers":    activePeers,
		"totalRevenue":   totalRevenue.StringFixed(2),
	}, "layouts/admin")
}

func (h *AdminHandler) Resellers(c fiber.Ctx) error {
	var resellers []models.User
	_ = h.svc.DB.Select(&resellers, "SELECT * FROM users WHERE role='reseller' ORDER BY created_at DESC")

	return c.Render("admin/resellers", fiber.Map{
		"email":     c.Locals("email"),
		"resellers": resellers,
	}, "layouts/admin")
}

func (h *AdminHandler) CreateReseller(c fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).SendString("Failed to hash password")
	}

	adminID := c.Locals("user_id").(int64)
	_, err = h.svc.DB.Exec(
		"INSERT INTO users (email, password, role, created_by) VALUES ($1, $2, 'reseller', $3)",
		email, string(hash), adminID,
	)
	if err != nil {
		return c.Status(400).SendString("Failed to create reseller: " + err.Error())
	}

	return c.Redirect().To("/admin/resellers")
}

func (h *AdminHandler) AddBalance(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	amountStr := c.FormValue("amount")
	amount, err := decimal.NewFromString(amountStr)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return c.Status(400).SendString("Invalid amount")
	}

	tx, err := h.svc.DB.Beginx()
	if err != nil {
		return c.Status(500).SendString("Transaction failed")
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE users SET balance = balance + $1 WHERE id=$2 AND role='reseller'", amount, id); err != nil {
		return c.Status(400).SendString("Failed to update balance")
	}

	if _, err := tx.Exec(
		"INSERT INTO transactions (reseller_id, amount, type, note) VALUES ($1, $2, 'credit', 'Admin balance top-up')",
		id, amount,
	); err != nil {
		return c.Status(500).SendString("Failed to record transaction")
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).SendString("Commit failed")
	}

	return c.Redirect().To("/admin/resellers")
}

func (h *AdminHandler) Packages(c fiber.Ctx) error {
	var packages []models.Package
	_ = h.svc.DB.Select(&packages, "SELECT * FROM packages ORDER BY duration_days ASC")

	return c.Render("admin/packages", fiber.Map{
		"email":    c.Locals("email"),
		"packages": packages,
	}, "layouts/admin")
}

func (h *AdminHandler) CreatePackage(c fiber.Ctx) error {
	name := c.FormValue("name")
	priceStr := c.FormValue("price_bdt")
	daysStr := c.FormValue("duration_days")

	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		return c.Status(400).SendString("Invalid price")
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return c.Status(400).SendString("Invalid duration")
	}

	_, err = h.svc.DB.Exec(
		"INSERT INTO packages (name, price_bdt, duration_days) VALUES ($1, $2, $3)",
		name, price, days,
	)
	if err != nil {
		return c.Status(400).SendString("Failed: " + err.Error())
	}

	return c.Redirect().To("/admin/packages")
}

func (h *AdminHandler) UpdatePackage(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	name := c.FormValue("name")
	priceStr := c.FormValue("price_bdt")
	daysStr := c.FormValue("duration_days")

	price, _ := decimal.NewFromString(priceStr)
	days, _ := strconv.Atoi(daysStr)

	_, err := h.svc.DB.Exec(
		"UPDATE packages SET name=$1, price_bdt=$2, duration_days=$3 WHERE id=$4",
		name, price, days, id,
	)
	if err != nil {
		return c.Status(400).SendString("Failed: " + err.Error())
	}

	return c.Redirect().To("/admin/packages")
}

func (h *AdminHandler) DeletePackage(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	_, _ = h.svc.DB.Exec("DELETE FROM packages WHERE id=$1", id)
	return c.Redirect().To("/admin/packages")
}

func (h *AdminHandler) Peers(c fiber.Ctx) error {
	var peers []models.VPNPeer
	_ = h.svc.DB.Select(&peers, `
		SELECT vp.*, u.email as reseller_email, p.name as package_name
		FROM vpn_peers vp
		JOIN users u ON vp.user_id = u.id
		JOIN packages p ON vp.package_id = p.id
		ORDER BY vp.created_at DESC LIMIT 100`)

	return c.Render("admin/peers", fiber.Map{
		"email": c.Locals("email"),
		"peers": peers,
	}, "layouts/admin")
}

func (h *AdminHandler) SearchPeers(c fiber.Ctx) error {
	q := c.Query("q")
	status := c.Query("status")

	query := `
		SELECT vp.*, u.email as reseller_email, p.name as package_name
		FROM vpn_peers vp
		JOIN users u ON vp.user_id = u.id
		JOIN packages p ON vp.package_id = p.id
		WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if q != "" {
		query += ` AND (vp.whatsapp_number ILIKE $` + strconv.Itoa(argIdx) + ` OR vp.public_key ILIKE $` + strconv.Itoa(argIdx) + `)`
		args = append(args, "%"+q+"%")
		argIdx++
	}
	if status != "" && (status == "active" || status == "expired") {
		query += ` AND vp.status = $` + strconv.Itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY vp.created_at DESC LIMIT 100"

	var peers []models.VPNPeer
	_ = h.svc.DB.Select(&peers, query, args...)

	return c.Render("admin/peers", fiber.Map{
		"email":  c.Locals("email"),
		"peers":  peers,
		"search": q,
		"status": status,
	}, "layouts/admin")
}
