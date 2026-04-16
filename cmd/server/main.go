package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"

	"github.com/imzami/routevpn-core/api"
	"github.com/imzami/routevpn-core/internal/config"
	"github.com/imzami/routevpn-core/internal/database"
	"github.com/imzami/routevpn-core/internal/handlers"
	"github.com/imzami/routevpn-core/internal/middleware"
	"github.com/imzami/routevpn-core/internal/services"
	"github.com/imzami/routevpn-core/internal/worker"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	cache, err := database.NewValkey(cfg.ValkeyURL)
	if err != nil {
		log.Fatalf("valkey: %v", err)
	}
	defer cache.Close()

	if err := services.EnsureDefaultAdmin(db, cfg); err != nil {
		log.Fatalf("admin seed: %v", err)
	}

	svc := &services.Container{
		DB:    db,
		Cache: cache,
		Cfg:   cfg,
	}

	w := worker.NewExpiryWorker(svc)
	go w.Start()
	defer w.Stop()

	engine := html.New("./templates", ".html")
	engine.Reload(true)

	app := fiber.New(fiber.Config{
		AppName:   "RouteVPN",
		BodyLimit: 4 * 1024 * 1024,
		Views:     engine,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use("/static", static.New("./static"))

	// Mount JSON REST API at /api/v1
	api.RegisterRoutes(app, svc)

	// Mount web UI routes
	setupRoutes(app, svc)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	_ = app.Shutdown()
}

func setupRoutes(app *fiber.App, svc *services.Container) {
	authH := handlers.NewAuthHandler(svc)
	adminH := handlers.NewAdminHandler(svc)
	resellerH := handlers.NewResellerHandler(svc)

	auth := app.Group("/auth")
	auth.Get("/login", authH.LoginPage)
	auth.Post("/login", authH.Login)
	auth.Get("/logout", authH.Logout)

	admin := app.Group("/admin", middleware.Auth(svc), middleware.RequireRole("admin"))
	admin.Get("/", adminH.Dashboard)
	admin.Get("/resellers", adminH.Resellers)
	admin.Post("/resellers", adminH.CreateReseller)
	admin.Post("/resellers/:id/balance", adminH.AddBalance)
	admin.Get("/packages", adminH.Packages)
	admin.Post("/packages", adminH.CreatePackage)
	admin.Put("/packages/:id", adminH.UpdatePackage)
	admin.Delete("/packages/:id", adminH.DeletePackage)
	admin.Get("/peers", adminH.Peers)
	admin.Get("/peers/search", adminH.SearchPeers)

	reseller := app.Group("/reseller", middleware.Auth(svc), middleware.RequireRole("reseller"))
	reseller.Get("/", resellerH.Dashboard)
	reseller.Get("/peers", resellerH.Peers)
	reseller.Post("/peers", resellerH.CreatePeer)
	reseller.Get("/peers/:id/config", resellerH.DownloadConfig)
	reseller.Get("/peers/:id/qr", resellerH.QRCode)
	reseller.Get("/peers/search", resellerH.SearchPeers)
	reseller.Get("/transactions", resellerH.Transactions)

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/auth/login")
	})
}
