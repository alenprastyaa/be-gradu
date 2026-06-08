package handlers

import (
	"bytes"
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WhatsappHandler struct {
	service *services.WhatsappService
}

func NewWhatsappHandler(service *services.WhatsappService) *WhatsappHandler {
	return &WhatsappHandler{service: service}
}

func (h *WhatsappHandler) One(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	data, err := h.service.LinkForStudent(c.UserContext(), id)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal membuat link WhatsApp", err.Error())
	}
	return utils.Success(c, "Link WhatsApp", data)
}

func (h *WhatsappHandler) MarkSent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	data, err := h.service.MarkSent(c.UserContext(), id)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal menandai WhatsApp terkirim", err.Error())
	}
	return utils.Success(c, "WhatsApp berhasil ditandai terkirim", data)
}

func (h *WhatsappHandler) ResetSent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	data, err := h.service.ResetSent(c.UserContext(), id)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal mereset status WhatsApp", err.Error())
	}
	return utils.Success(c, "WhatsApp berhasil direset", data)
}

func (h *WhatsappHandler) All(c *fiber.Ctx) error {
	var data interface{}
	var err error
	body := bytes.TrimSpace(c.Body())
	if len(body) == 0 {
		data, err = h.service.Links(c.UserContext())
	} else {
		var req bulkStudentIDsRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
		}
		if len(req.StudentIDs) == 0 {
			return utils.Error(c, fiber.StatusBadRequest, "Pilih minimal satu siswa", nil)
		}
		data, err = h.service.LinksForStudents(c.UserContext(), req.StudentIDs)
	}
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal membuat link WhatsApp", err.Error())
	}
	return utils.Success(c, "Daftar link WhatsApp", data)
}
