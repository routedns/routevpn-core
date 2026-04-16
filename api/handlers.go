package api

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/imzami/routevpn-core/internal/models"
	"github.com/imzami/routevpn-core/internal/services"
	"github.com/imzami/routevpn-core/internal/wg"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

// ──────────────────────────────────────────────
// Admin: Dashboard Stats
// ──────────────────────────────────────────────

func AdminStats(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		type stats struct {
			TotalResellers int    `json:"total_resellers"`
			TotalPeers     int    `json:"total_peers"`
			ActivePeers    int    `json:"active_peers"`
			ExpiredPeers   int    `json:"expired_peers"`
			TotalRevenue   string `json:"total_revenue_bdt"`
			TotalCredits   string `json:"total_credits_bdt"`
		}
		var s stats
		_ = svc.DB.Get(&s.TotalResellers, "SELECT COUNT(*) FROM users WHERE role='reseller'")
		_ = svc.DB.Get(&s.TotalPeers, "SELECT COUNT(*) FROM vpn_peers")
		_ = svc.DB.Get(&s.ActivePeers, "SELECT COUNT(*) FROM vpn_peers WHERE status='active'")
		_ = svc.DB.Get(&s.ExpiredPeers, "SELECT COUNT(*) FROM vpn_peers WHERE status='expired'")

		var rev, cred decimal.Decimal
		_ = svc.DB.Get(&rev, "SELECT COALESCE(SUM(amount),0) FROM transactions WHERE type='debit'")
		_ = svc.DB.Get(&cred, "SELECT COALESCE(SUM(amount),0) FROM transactions WHERE type='credit'")
		s.TotalRevenue = rev.StringFixed(2)
		s.TotalCredits = cred.StringFixed(2)

		return OK(c, s)
	}
}

// ──────────────────────────────────────────────
// Admin: Reseller CRUD
// ──────────────────────────────────────────────

func AdminListResellers(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		page, perPage := parsePagination(c)
		offset := (page - 1) * perPage

		var total int64
		_ = svc.DB.Get(&total, "SELECT COUNT(*) FROM users WHERE role='reseller'")

		var resellers []models.User
		err := svc.DB.Select(&resellers,
			"SELECT * FROM users WHERE role='reseller' ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			perPage, offset)
		if err != nil {
			return Internal(c, "failed to fetch resellers")
		}

		return OKMeta(c, resellers, &Meta{Page: page, PerPage: perPage, Total: total})
	}
}

type createResellerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func AdminCreateReseller(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req createResellerReq
		if err := c.Bind().JSON(&req); err != nil {
			return BadRequest(c, "invalid JSON body")
		}
		if req.Email == "" || len(req.Password) < 6 {
			return BadRequest(c, "email required, password min 6 characters")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return Internal(c, "password hashing failed")
		}

		adminID := c.Locals("user_id").(int64)
		var reseller models.User
		err = svc.DB.QueryRowx(
			`INSERT INTO users (email, password, role, created_by) VALUES ($1, $2, 'reseller', $3)
			 RETURNING id, email, role, balance, created_by, created_at`,
			req.Email, string(hash), adminID,
		).StructScan(&reseller)
		if err != nil {
			return BadRequest(c, "failed to create reseller: "+err.Error())
		}

		return Created(c, reseller)
	}
}

func AdminGetReseller(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid reseller id")
		}

		var reseller models.User
		if err := svc.DB.Get(&reseller, "SELECT * FROM users WHERE id=$1 AND role='reseller'", id); err != nil {
			return NotFound(c, "reseller not found")
		}

		// Include peer count and transaction summary
		type detail struct {
			models.User
			PeerCount  int    `json:"peer_count"`
			TotalSpent string `json:"total_spent_bdt"`
		}
		d := detail{User: reseller}
		_ = svc.DB.Get(&d.PeerCount, "SELECT COUNT(*) FROM vpn_peers WHERE user_id=$1", id)
		var spent decimal.Decimal
		_ = svc.DB.Get(&spent, "SELECT COALESCE(SUM(amount),0) FROM transactions WHERE reseller_id=$1 AND type='debit'", id)
		d.TotalSpent = spent.StringFixed(2)

		return OK(c, d)
	}
}

type addBalanceReq struct {
	Amount string `json:"amount"`
	Note   string `json:"note"`
}

func AdminAddBalance(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid reseller id")
		}

		var req addBalanceReq
		if err := c.Bind().JSON(&req); err != nil {
			return BadRequest(c, "invalid JSON body")
		}

		amount, err := decimal.NewFromString(req.Amount)
		if err != nil || amount.LessThanOrEqual(decimal.Zero) {
			return BadRequest(c, "amount must be a positive number")
		}

		note := req.Note
		if note == "" {
			note = "Admin balance top-up"
		}

		tx, err := svc.DB.Beginx()
		if err != nil {
			return Internal(c, "transaction start failed")
		}
		defer tx.Rollback()

		res, err := tx.Exec("UPDATE users SET balance = balance + $1 WHERE id=$2 AND role='reseller'", amount, id)
		if err != nil {
			return Internal(c, "balance update failed")
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return NotFound(c, "reseller not found")
		}

		if _, err := tx.Exec(
			"INSERT INTO transactions (reseller_id, amount, type, note) VALUES ($1, $2, 'credit', $3)",
			id, amount, note,
		); err != nil {
			return Internal(c, "transaction record failed")
		}

		if err := tx.Commit(); err != nil {
			return Internal(c, "commit failed")
		}

		// Return updated reseller
		var reseller models.User
		_ = svc.DB.Get(&reseller, "SELECT * FROM users WHERE id=$1", id)
		return OK(c, reseller)
	}
}

// ──────────────────────────────────────────────
// Admin: Package CRUD
// ──────────────────────────────────────────────

func AdminListPackages(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		var packages []models.Package
		if err := svc.DB.Select(&packages, "SELECT * FROM packages ORDER BY duration_days ASC"); err != nil {
			return Internal(c, "failed to fetch packages")
		}
		return OK(c, packages)
	}
}

type packageReq struct {
	Name         string `json:"name"`
	PriceBDT     string `json:"price_bdt"`
	DurationDays int    `json:"duration_days"`
}

func AdminCreatePackage(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req packageReq
		if err := c.Bind().JSON(&req); err != nil {
			return BadRequest(c, "invalid JSON body")
		}
		if req.Name == "" || req.DurationDays < 1 {
			return BadRequest(c, "name required, duration_days must be >= 1")
		}
		price, err := decimal.NewFromString(req.PriceBDT)
		if err != nil || price.LessThanOrEqual(decimal.Zero) {
			return BadRequest(c, "price_bdt must be a positive number")
		}

		var pkg models.Package
		err = svc.DB.QueryRowx(
			"INSERT INTO packages (name, price_bdt, duration_days) VALUES ($1, $2, $3) RETURNING *",
			req.Name, price, req.DurationDays,
		).StructScan(&pkg)
		if err != nil {
			return Internal(c, "failed to create package: "+err.Error())
		}
		return Created(c, pkg)
	}
}

func AdminUpdatePackage(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid package id")
		}

		var req packageReq
		if err := c.Bind().JSON(&req); err != nil {
			return BadRequest(c, "invalid JSON body")
		}
		if req.Name == "" || req.DurationDays < 1 {
			return BadRequest(c, "name required, duration_days must be >= 1")
		}
		price, err := decimal.NewFromString(req.PriceBDT)
		if err != nil || price.LessThanOrEqual(decimal.Zero) {
			return BadRequest(c, "price_bdt must be a positive number")
		}

		var pkg models.Package
		err = svc.DB.QueryRowx(
			"UPDATE packages SET name=$1, price_bdt=$2, duration_days=$3 WHERE id=$4 RETURNING *",
			req.Name, price, req.DurationDays, id,
		).StructScan(&pkg)
		if err != nil {
			return NotFound(c, "package not found")
		}
		return OK(c, pkg)
	}
}

func AdminDeletePackage(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid package id")
		}

		// Prevent deleting packages with active peers
		var activeCount int
		_ = svc.DB.Get(&activeCount, "SELECT COUNT(*) FROM vpn_peers WHERE package_id=$1 AND status='active'", id)
		if activeCount > 0 {
			return BadRequest(c, fmt.Sprintf("cannot delete: %d active peers use this package", activeCount))
		}

		res, err := svc.DB.Exec("DELETE FROM packages WHERE id=$1", id)
		if err != nil {
			return Internal(c, "delete failed")
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return NotFound(c, "package not found")
		}
		return OK(c, fiber.Map{"deleted": true})
	}
}

// ──────────────────────────────────────────────
// Admin: Peer Management
// ──────────────────────────────────────────────

func AdminListPeers(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		page, perPage := parsePagination(c)
		offset := (page - 1) * perPage

		q := c.Query("q")
		status := c.Query("status")
		resellerID := c.Query("reseller_id")

		query := `SELECT vp.*, u.email as reseller_email, p.name as package_name
			FROM vpn_peers vp
			JOIN users u ON vp.user_id = u.id
			JOIN packages p ON vp.package_id = p.id WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM vpn_peers vp WHERE 1=1`

		args := []interface{}{}
		countArgs := []interface{}{}
		idx := 1

		if q != "" {
			clause := fmt.Sprintf(` AND (vp.whatsapp_number ILIKE $%d OR vp.public_key ILIKE $%d)`, idx, idx)
			query += clause
			countQuery += clause
			args = append(args, "%"+q+"%")
			countArgs = append(countArgs, "%"+q+"%")
			idx++
		}
		if status == "active" || status == "expired" {
			clause := fmt.Sprintf(` AND vp.status = $%d`, idx)
			query += clause
			countQuery += clause
			args = append(args, status)
			countArgs = append(countArgs, status)
			idx++
		}
		if resellerID != "" {
			rid, err := strconv.ParseInt(resellerID, 10, 64)
			if err == nil {
				clause := fmt.Sprintf(` AND vp.user_id = $%d`, idx)
				query += clause
				countQuery += clause
				args = append(args, rid)
				countArgs = append(countArgs, rid)
				idx++
			}
		}

		var total int64
		_ = svc.DB.Get(&total, countQuery, countArgs...)

		query += fmt.Sprintf(` ORDER BY vp.created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
		args = append(args, perPage, offset)

		var peers []models.VPNPeer
		if err := svc.DB.Select(&peers, query, args...); err != nil {
			return Internal(c, "query failed")
		}

		return OKMeta(c, peers, &Meta{Page: page, PerPage: perPage, Total: total})
	}
}

func AdminGetPeer(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid peer id")
		}

		var peer models.VPNPeer
		err = svc.DB.Get(&peer, `
			SELECT vp.*, u.email as reseller_email, p.name as package_name
			FROM vpn_peers vp
			JOIN users u ON vp.user_id = u.id
			JOIN packages p ON vp.package_id = p.id
			WHERE vp.id=$1`, id)
		if err != nil {
			return NotFound(c, "peer not found")
		}
		return OK(c, peer)
	}
}

func AdminRemovePeer(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid peer id")
		}

		var peer models.VPNPeer
		if err := svc.DB.Get(&peer, "SELECT * FROM vpn_peers WHERE id=$1", id); err != nil {
			return NotFound(c, "peer not found")
		}

		// Remove from WG interface
		_ = wg.RemovePeer(svc.Cfg.WGInterface, peer.PublicKey)

		// Update status
		if _, err := svc.DB.Exec("UPDATE vpn_peers SET status='expired' WHERE id=$1", id); err != nil {
			return Internal(c, "failed to update peer status")
		}

		return OK(c, fiber.Map{"removed": true, "peer_id": id})
	}
}

// ──────────────────────────────────────────────
// Admin: Transactions
// ──────────────────────────────────────────────

func AdminListTransactions(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		page, perPage := parsePagination(c)
		offset := (page - 1) * perPage

		resellerID := c.Query("reseller_id")
		txType := c.Query("type")

		query := `SELECT t.*, u.email as reseller_email
			FROM transactions t
			JOIN users u ON t.reseller_id = u.id WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM transactions t WHERE 1=1`
		args := []interface{}{}
		countArgs := []interface{}{}
		idx := 1

		if resellerID != "" {
			if rid, err := strconv.ParseInt(resellerID, 10, 64); err == nil {
				clause := fmt.Sprintf(` AND t.reseller_id = $%d`, idx)
				query += clause
				countQuery += clause
				args = append(args, rid)
				countArgs = append(countArgs, rid)
				idx++
			}
		}
		if txType == "credit" || txType == "debit" {
			clause := fmt.Sprintf(` AND t.type = $%d`, idx)
			query += clause
			countQuery += clause
			args = append(args, txType)
			countArgs = append(countArgs, txType)
			idx++
		}

		var total int64
		_ = svc.DB.Get(&total, countQuery, countArgs...)

		query += fmt.Sprintf(` ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
		args = append(args, perPage, offset)

		var txns []models.Transaction
		if err := svc.DB.Select(&txns, query, args...); err != nil {
			return Internal(c, "query failed")
		}

		return OKMeta(c, txns, &Meta{Page: page, PerPage: perPage, Total: total})
	}
}

// ──────────────────────────────────────────────
// Reseller: Endpoints
// ──────────────────────────────────────────────

func ResellerStats(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(int64)

		type stats struct {
			Balance     string `json:"balance_bdt"`
			TotalPeers  int    `json:"total_peers"`
			ActivePeers int    `json:"active_peers"`
			TotalSpent  string `json:"total_spent_bdt"`
		}
		var s stats
		var bal, spent decimal.Decimal
		_ = svc.DB.Get(&bal, "SELECT balance FROM users WHERE id=$1", userID)
		_ = svc.DB.Get(&s.TotalPeers, "SELECT COUNT(*) FROM vpn_peers WHERE user_id=$1", userID)
		_ = svc.DB.Get(&s.ActivePeers, "SELECT COUNT(*) FROM vpn_peers WHERE user_id=$1 AND status='active'", userID)
		_ = svc.DB.Get(&spent, "SELECT COALESCE(SUM(amount),0) FROM transactions WHERE reseller_id=$1 AND type='debit'", userID)
		s.Balance = bal.StringFixed(2)
		s.TotalSpent = spent.StringFixed(2)

		return OK(c, s)
	}
}

type createPeerReq struct {
	WhatsappNumber string `json:"whatsapp_number"`
	PackageID      int64  `json:"package_id"`
}

func ResellerCreatePeer(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(int64)

		var req createPeerReq
		if err := c.Bind().JSON(&req); err != nil {
			return BadRequest(c, "invalid JSON body")
		}
		if req.WhatsappNumber == "" || req.PackageID == 0 {
			return BadRequest(c, "whatsapp_number and package_id are required")
		}

		peer, err := services.CreatePeer(svc, services.CreatePeerInput{
			WhatsappNumber: req.WhatsappNumber,
			PackageID:      req.PackageID,
			ResellerID:     userID,
		})
		if err != nil {
			return BadRequest(c, err.Error())
		}

		return Created(c, peer)
	}
}

func ResellerListPeers(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(int64)
		page, perPage := parsePagination(c)
		offset := (page - 1) * perPage

		q := c.Query("q")
		status := c.Query("status")

		query := `SELECT vp.*, p.name as package_name
			FROM vpn_peers vp
			JOIN packages p ON vp.package_id = p.id
			WHERE vp.user_id=$1`
		countQuery := `SELECT COUNT(*) FROM vpn_peers WHERE user_id=$1`
		args := []interface{}{userID}
		countArgs := []interface{}{userID}
		idx := 2

		if q != "" {
			clause := fmt.Sprintf(` AND (vp.whatsapp_number ILIKE $%d OR vp.public_key ILIKE $%d)`, idx, idx)
			query += clause
			countQuery += fmt.Sprintf(` AND (whatsapp_number ILIKE $%d OR public_key ILIKE $%d)`, idx, idx)
			args = append(args, "%"+q+"%")
			countArgs = append(countArgs, "%"+q+"%")
			idx++
		}
		if status == "active" || status == "expired" {
			clause := fmt.Sprintf(` AND vp.status = $%d`, idx)
			query += clause
			countQuery += fmt.Sprintf(` AND status = $%d`, idx)
			args = append(args, status)
			countArgs = append(countArgs, status)
			idx++
		}

		var total int64
		_ = svc.DB.Get(&total, countQuery, countArgs...)

		query += fmt.Sprintf(` ORDER BY vp.created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
		args = append(args, perPage, offset)

		var peers []models.VPNPeer
		if err := svc.DB.Select(&peers, query, args...); err != nil {
			return Internal(c, "query failed")
		}

		return OKMeta(c, peers, &Meta{Page: page, PerPage: perPage, Total: total})
	}
}

func ResellerGetPeerConfig(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(int64)
		peerID, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid peer id")
		}

		var peer models.VPNPeer
		if err := svc.DB.Get(&peer, "SELECT * FROM vpn_peers WHERE id=$1 AND user_id=$2", peerID, userID); err != nil {
			return NotFound(c, "peer not found")
		}

		conf := wg.GenerateClientConfig(&peer, svc.Cfg)

		format := c.Query("format", "text")
		if format == "json" {
			return OK(c, fiber.Map{
				"peer_id": peer.ID,
				"config":  conf,
			})
		}

		c.Set("Content-Type", "text/plain; charset=utf-8")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=awg-%s.conf", peer.WhatsappNumber))
		return c.SendString(conf)
	}
}

func ResellerGetPeerQR(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(int64)
		peerID, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return BadRequest(c, "invalid peer id")
		}

		var peer models.VPNPeer
		if err := svc.DB.Get(&peer, "SELECT * FROM vpn_peers WHERE id=$1 AND user_id=$2", peerID, userID); err != nil {
			return NotFound(c, "peer not found")
		}

		conf := wg.GenerateClientConfig(&peer, svc.Cfg)
		png, err := wg.GenerateQRCode(conf, 512)
		if err != nil {
			return Internal(c, "QR generation failed")
		}

		format := c.Query("format", "png")
		if format == "base64" {
			return OK(c, fiber.Map{
				"peer_id": peer.ID,
				"qr_png":  base64.StdEncoding.EncodeToString(png),
			})
		}

		c.Set("Content-Type", "image/png")
		return c.Send(png)
	}
}

func ResellerListTransactions(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(int64)
		page, perPage := parsePagination(c)
		offset := (page - 1) * perPage

		txType := c.Query("type")

		query := `SELECT * FROM transactions WHERE reseller_id=$1`
		countQuery := `SELECT COUNT(*) FROM transactions WHERE reseller_id=$1`
		args := []interface{}{userID}
		countArgs := []interface{}{userID}
		idx := 2

		if txType == "credit" || txType == "debit" {
			clause := fmt.Sprintf(` AND type = $%d`, idx)
			query += clause
			countQuery += clause
			args = append(args, txType)
			countArgs = append(countArgs, txType)
			idx++
		}

		var total int64
		_ = svc.DB.Get(&total, countQuery, countArgs...)

		query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
		args = append(args, perPage, offset)

		var txns []models.Transaction
		if err := svc.DB.Select(&txns, query, args...); err != nil {
			return Internal(c, "query failed")
		}

		return OKMeta(c, txns, &Meta{Page: page, PerPage: perPage, Total: total})
	}
}

func ResellerListPackages(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		var packages []models.Package
		if err := svc.DB.Select(&packages, "SELECT * FROM packages ORDER BY duration_days ASC"); err != nil {
			return Internal(c, "failed to fetch packages")
		}
		return OK(c, packages)
	}
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func parsePagination(c fiber.Ctx) (page, perPage int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	perPage, _ = strconv.Atoi(c.Query("per_page", "25"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return
}
