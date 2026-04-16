package api

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/imzami/routevpn-core/internal/services"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	Role      string `json:"role"`
	Email     string `json:"email"`
}

// Login authenticates a user and returns a Bearer JWT.
func Login(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		var req loginRequest
		if err := c.Bind().JSON(&req); err != nil {
			return BadRequest(c, "invalid JSON body")
		}
		if req.Email == "" || req.Password == "" {
			return BadRequest(c, "email and password are required")
		}

		user, err := services.AuthenticateUser(svc.DB, req.Email, req.Password)
		if err != nil {
			return Unauthorized(c, "invalid credentials")
		}

		exp := time.Now().Add(24 * time.Hour)
		claims := Claims{
			UserID: user.ID,
			Email:  user.Email,
			Role:   user.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(exp),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(svc.Cfg.JWTSecret))
		if err != nil {
			return Internal(c, "failed to generate token")
		}

		return OK(c, tokenResponse{
			Token:     signed,
			ExpiresAt: exp.Format(time.RFC3339),
			Role:      user.Role,
			Email:     user.Email,
		})
	}
}

// Me returns the current authenticated user's profile.
func Me(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(int64)

		var user struct {
			ID        int64  `json:"id" db:"id"`
			Email     string `json:"email" db:"email"`
			Role      string `json:"role" db:"role"`
			Balance   string `json:"balance"`
			CreatedAt string `json:"created_at" db:"created_at"`
		}

		row := svc.DB.QueryRowx("SELECT id, email, role, balance::text, created_at::text FROM users WHERE id=$1", userID)
		if err := row.StructScan(&user); err != nil {
			return NotFound(c, "user not found")
		}
		return OK(c, user)
	}
}
