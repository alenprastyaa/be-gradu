package handlers

import (
	"graduation-invitation/internal/models"
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SuperAdminHandler struct {
	service *services.SuperAdminService
}

func NewSuperAdminHandler(service *services.SuperAdminService) *SuperAdminHandler {
	return &SuperAdminHandler{service: service}
}

func (h *SuperAdminHandler) ListSchools(c *fiber.Ctx) error {
	data, err := h.service.ListSchools(c.UserContext())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil sekolah", err.Error())
	}
	return utils.Success(c, "Daftar sekolah", data)
}

func (h *SuperAdminHandler) CreateSchool(c *fiber.Ctx) error {
	var input models.SchoolInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.CreateSchool(c.UserContext(), input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return utils.Created(c, "Sekolah berhasil dibuat", data)
}

func (h *SuperAdminHandler) UpdateSchool(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	var input models.SchoolInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.UpdateSchool(c.UserContext(), id, input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if data == nil {
		return utils.Error(c, fiber.StatusNotFound, "Sekolah tidak ditemukan", nil)
	}
	return utils.Success(c, "Sekolah berhasil diperbarui", data)
}

func (h *SuperAdminHandler) DeleteSchool(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	if err := h.service.DeleteSchool(c.UserContext(), id); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal menghapus sekolah", err.Error())
	}
	return utils.Success(c, "Sekolah berhasil dihapus", nil)
}

func (h *SuperAdminHandler) ListAdmins(c *fiber.Ctx) error {
	data, err := h.service.ListAdmins(c.UserContext())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil admin", err.Error())
	}
	return utils.Success(c, "Daftar admin", data)
}

func (h *SuperAdminHandler) CreateAdmin(c *fiber.Ctx) error {
	var input models.AdminInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.CreateAdmin(c.UserContext(), input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return utils.Created(c, "Admin berhasil dibuat", data)
}

func (h *SuperAdminHandler) UpdateAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	var input models.AdminInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.UpdateAdmin(c.UserContext(), id, input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if data == nil {
		return utils.Error(c, fiber.StatusNotFound, "Admin tidak ditemukan", nil)
	}
	return utils.Success(c, "Admin berhasil diperbarui", data)
}

func (h *SuperAdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	if err := h.service.DeleteAdmin(c.UserContext(), id); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal menghapus admin", err.Error())
	}
	return utils.Success(c, "Admin berhasil dihapus", nil)
}
