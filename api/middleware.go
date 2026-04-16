package api

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/imzami/routevpn-core/internal/services"
)

// BearerAuth extracts JWT from Authorization: Bearer <token> header.
func BearerAuth(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return Unauthorized(c, "missing authorization header")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return Unauthorized(c, "invalid authorization format, use: Bearer <token>")
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(svc.Cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			return Unauthorized(c, "invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

// RequireRole restricts access to a specific role via JSON error.
func RequireRole(role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Locals("role") != role {
			return Forbidden(c, "insufficient permissions")
		}
		return c.Next()
	}
}

// RequireAnyRole allows access if the user has any of the specified roles.
func RequireAnyRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, _ := c.Locals("role").(string)
		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}
		return Forbidden(c, "insufficient permissions")
	}
}

// Claims is the JWT claims structure used by the API.
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
