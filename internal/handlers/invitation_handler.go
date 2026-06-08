package handlers

import (
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type InvitationHandler struct {
	service *services.InvitationService
}

func NewInvitationHandler(service *services.InvitationService) *InvitationHandler {
	return &InvitationHandler{service: service}
}

func (h *InvitationHandler) Get(c *fiber.Ctx) error {
	data, err := h.service.GetByCode(c.UserContext(), c.Params("code"))
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil undangan", err.Error())
	}
	if data == nil {
		return utils.Error(c, fiber.StatusNotFound, "Undangan tidak ditemukan", nil)
	}
	return utils.Success(c, "Undangan ditemukan", data)
}
