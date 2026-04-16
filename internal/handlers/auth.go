package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/imzami/routevpn-core/internal/middleware"
	"github.com/imzami/routevpn-core/internal/services"
)

type AuthHandler struct {
	svc *services.Container
}

func NewAuthHandler(svc *services.Container) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) LoginPage(c fiber.Ctx) error {
	return c.Render("auth/login", fiber.Map{"error": ""})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	user, err := services.AuthenticateUser(h.svc.DB, email, password)
	if err != nil {
		return c.Render("auth/login", fiber.Map{"error": "Invalid email or password"})
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role, h.svc.Cfg.JWTSecret)
	if err != nil {
		return c.Render("auth/login", fiber.Map{"error": "Internal error"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	if user.Role == "admin" {
		return c.Redirect().To("/admin/")
	}
	return c.Redirect().To("/reseller/")
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})
	return c.Redirect().To("/auth/login")
}
