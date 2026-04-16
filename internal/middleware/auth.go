package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/imzami/routevpn-core/internal/services"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, email, role, secret string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Auth(svc *services.Container) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenStr := c.Cookies("token")
		if tokenStr == "" {
			return c.Redirect().To("/auth/login")
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(svc.Cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Redirect().To("/auth/login")
		}
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

func RequireRole(role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Locals("role") != role {
			return c.Status(403).SendString("Forbidden")
		}
		return c.Next()
	}
}
