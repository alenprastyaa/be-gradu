package handlers

import (
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AttendanceHandler struct {
	service *services.AttendanceService
}

func NewAttendanceHandler(service *services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{service: service}
}

func (h *AttendanceHandler) Scan(c *fiber.Ctx) error {
	var body struct {
		QRPayload string `json:"qr_payload"`
		Payload   string `json:"payload"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	payload := body.QRPayload
	if payload == "" {
		payload = body.Payload
	}
	result, err := h.service.Scan(c.UserContext(), payload)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	message := "Absensi berhasil"
	if result["status"] == "already_attended" {
		message = "Siswa sudah absen sebelumnya"
	}
	return utils.Success(c, message, result)
}

func (h *AttendanceHandler) Summary(c *fiber.Ctx) error {
	data, err := h.service.Summary(c.UserContext())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil rekap", err.Error())
	}
	return utils.Success(c, "Rekap absensi", data)
}
