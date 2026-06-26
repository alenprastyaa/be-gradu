package handlers

import (
	"graduation-invitation/internal/models"
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeacherInviteHandler struct {
	service *services.TeacherInviteService
}

func NewTeacherInviteHandler(service *services.TeacherInviteService) *TeacherInviteHandler {
	return &TeacherInviteHandler{service: service}
}

func (h *TeacherInviteHandler) List(c *fiber.Ctx) error {
	data, err := h.service.List(c.UserContext(), models.TeacherInviteFilter{
		Search:           c.Query("search"),
		AttendanceStatus: c.Query("attendance_status"),
		Page:             queryTeacherInt(c, "page", 1),
		Limit:            queryTeacherInt(c, "limit", 10),
	})
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil undangan guru", err.Error())
	}
	return utils.Success(c, "Data undangan guru", data)
}

func (h *TeacherInviteHandler) Create(c *fiber.Ctx) error {
	var input models.TeacherInviteInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.Create(c.UserContext(), input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return utils.Created(c, "Undangan guru berhasil dibuat", data)
}

func (h *TeacherInviteHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	var input models.TeacherInviteInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.Update(c.UserContext(), id, input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if data == nil {
		return utils.Error(c, fiber.StatusNotFound, "Undangan guru tidak ditemukan", nil)
	}
	return utils.Success(c, "Undangan guru berhasil diperbarui", data)
}

func (h *TeacherInviteHandler) UpdateAttendanceStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	var body struct {
		AttendanceStatus string `json:"attendance_status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.UpdateAttendanceStatus(c.UserContext(), id, body.AttendanceStatus)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if data == nil {
		return utils.Error(c, fiber.StatusNotFound, "Undangan guru tidak ditemukan", nil)
	}
	return utils.Success(c, "Status absensi guru berhasil diperbarui", data)
}

func (h *TeacherInviteHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	if err := h.service.Delete(c.UserContext(), id); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menghapus undangan guru", err.Error())
	}
	return utils.Success(c, "Undangan guru berhasil dihapus", nil)
}

func (h *TeacherInviteHandler) Import(c *fiber.Ctx) error {
	header, err := c.FormFile("file")
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "File import wajib diunggah", err.Error())
	}
	file, err := header.Open()
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal membuka file", err.Error())
	}
	defer file.Close()
	rows, err := utils.ParseTeacherImportFile(file, header.Filename)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	result, err := h.service.Import(c.UserContext(), rows)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), result)
	}
	return utils.Success(c, "Import undangan guru berhasil", result)
}

func (h *TeacherInviteHandler) ImportTemplate(c *fiber.Ctx) error {
	data, err := h.service.ImportTemplateXLSX()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal membuat template import guru", err.Error())
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=template-import-guru.xlsx")
	c.Set("Content-Length", strconv.Itoa(len(data)))
	return c.Send(data)
}

func queryTeacherInt(c *fiber.Ctx, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
}
