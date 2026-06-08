package handlers

import (
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	token, admin, err := h.service.Login(c.UserContext(), body.Email, body.Password)
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, err.Error(), nil)
	}
	return utils.Success(c, "Login berhasil", fiber.Map{"token": token, "admin": admin})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	adminID, ok := c.Locals("admin_id").(uuid.UUID)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Token tidak valid", nil)
	}
	admin, err := h.service.Me(c.UserContext(), adminID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data admin", err.Error())
	}
	if admin == nil {
		return utils.Error(c, fiber.StatusNotFound, "Admin tidak ditemukan", nil)
	}
	return utils.Success(c, "Data admin", admin)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return utils.Success(c, "Logout berhasil", nil)
}
