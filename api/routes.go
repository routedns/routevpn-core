package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/imzami/routevpn-core/internal/services"
)

// RegisterRoutes mounts all /api/v1 JSON endpoints onto the Fiber app.
func RegisterRoutes(app *fiber.App, svc *services.Container) {
	v1 := app.Group("/api/v1")

	// ── Public ──────────────────────────────
	v1.Post("/auth/login", Login(svc))

	// ── Health ──────────────────────────────
	v1.Get("/health", func(c fiber.Ctx) error {
		if err := svc.DB.Ping(); err != nil {
			return Fail(c, 503, "database unreachable")
		}
		return OK(c, fiber.Map{"status": "ok"})
	})

	// ── Authenticated ──────────────────────
	authed := v1.Group("", BearerAuth(svc))
	authed.Get("/auth/me", Me(svc))

	// ── Admin ──────────────────────────────
	admin := authed.Group("/admin", RequireRole("admin"))

	admin.Get("/stats", AdminStats(svc))

	admin.Get("/resellers", AdminListResellers(svc))
	admin.Post("/resellers", AdminCreateReseller(svc))
	admin.Get("/resellers/:id", AdminGetReseller(svc))
	admin.Post("/resellers/:id/balance", AdminAddBalance(svc))

	admin.Get("/packages", AdminListPackages(svc))
	admin.Post("/packages", AdminCreatePackage(svc))
	admin.Put("/packages/:id", AdminUpdatePackage(svc))
	admin.Delete("/packages/:id", AdminDeletePackage(svc))

	admin.Get("/peers", AdminListPeers(svc))
	admin.Get("/peers/:id", AdminGetPeer(svc))
	admin.Delete("/peers/:id", AdminRemovePeer(svc))

	admin.Get("/transactions", AdminListTransactions(svc))

	// ── Reseller ───────────────────────────
	reseller := authed.Group("/reseller", RequireRole("reseller"))

	reseller.Get("/stats", ResellerStats(svc))
	reseller.Get("/packages", ResellerListPackages(svc))

	reseller.Get("/peers", ResellerListPeers(svc))
	reseller.Post("/peers", ResellerCreatePeer(svc))
	reseller.Get("/peers/:id/config", ResellerGetPeerConfig(svc))
	reseller.Get("/peers/:id/qr", ResellerGetPeerQR(svc))

	reseller.Get("/transactions", ResellerListTransactions(svc))
}
