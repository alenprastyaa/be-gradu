package middlewares

import (
	"strings"

	"graduation-invitation/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func AuthMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return utils.Error(c, fiber.StatusUnauthorized, "Token tidak tersedia", nil)
		}
		tokenText := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenText, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return utils.Error(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return utils.Error(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
		}
		sub, _ := claims["sub"].(string)
		adminID, err := uuid.Parse(sub)
		if err != nil {
			return utils.Error(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
		}
		c.Locals("admin_id", adminID)
		return c.Next()
	}
}
