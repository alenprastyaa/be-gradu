package middlewares

import (
	"strings"

	"graduation-invitation/internal/authcontext"
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
		role, _ := claims["role"].(string)
		if role == "" {
			role = authcontext.RoleSchoolAdmin
		}
		var schoolID *uuid.UUID
		if rawSchoolID, _ := claims["school_id"].(string); rawSchoolID != "" {
			parsed, err := uuid.Parse(rawSchoolID)
			if err != nil {
				return utils.Error(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
			}
			schoolID = &parsed
		}
		c.Locals("admin_id", adminID)
		c.Locals("role", role)
		if schoolID != nil {
			c.Locals("school_id", *schoolID)
		}
		c.SetUserContext(authcontext.WithAuth(c.UserContext(), authcontext.AuthInfo{
			AdminID:  adminID,
			Email:    stringClaim(claims, "email"),
			Role:     role,
			SchoolID: schoolID,
		}))
		return c.Next()
	}
}

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		info, ok := authcontext.FromContext(c.UserContext())
		if !ok || info.Role != role {
			return utils.Error(c, fiber.StatusForbidden, "Akses tidak diizinkan", nil)
		}
		return c.Next()
	}
}

func stringClaim(claims jwt.MapClaims, key string) string {
	value, _ := claims[key].(string)
	return value
}
