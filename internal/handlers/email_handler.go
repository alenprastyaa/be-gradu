package handlers

import (
	"bytes"
	"encoding/json"
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type bulkStudentIDsRequest struct {
	StudentIDs []uuid.UUID `json:"student_ids"`
}

type EmailHandler struct {
	service *services.EmailService
}

func NewEmailHandler(service *services.EmailService) *EmailHandler {
	return &EmailHandler{service: service}
}

func (h *EmailHandler) One(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	data, err := h.service.SendForStudent(c.UserContext(), id)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal mengirim email", err.Error())
	}
	return utils.Success(c, "Email undangan diproses", data)
}

func (h *EmailHandler) All(c *fiber.Ctx) error {
	var data interface{}
	var err error
	body := bytes.TrimSpace(c.Body())
	if len(body) == 0 {
		data, err = h.service.SendAll(c.UserContext())
	} else {
		var req bulkStudentIDsRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
		}
		if len(req.StudentIDs) == 0 {
			return utils.Error(c, fiber.StatusBadRequest, "Pilih minimal satu siswa", nil)
		}
		data, err = h.service.SendForStudents(c.UserContext(), req.StudentIDs)
	}
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengirim email", err.Error())
	}
	return utils.Success(c, "Email undangan diproses", data)
}

func (h *EmailHandler) ResetSent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	data, err := h.service.ResetSent(c.UserContext(), id)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal reset status email", err.Error())
	}
	return utils.Success(c, "Status email berhasil direset", data)
}

func (h *EmailHandler) BrevoWebhook(c *fiber.Ctx) error {
	secret := strings.TrimSpace(c.Get("X-Brevo-Webhook-Secret"))
	if secret == "" {
		secret = strings.TrimSpace(c.Query("secret"))
	}
	if !h.service.ValidateWebhookSecret(secret) {
		return utils.Error(c, fiber.StatusUnauthorized, "Webhook Brevo tidak valid", nil)
	}

	raw := bytes.TrimSpace(c.Body())
	if len(raw) == 0 {
		return utils.Error(c, fiber.StatusBadRequest, "Payload webhook kosong", nil)
	}
	var batch []json.RawMessage
	if err := json.Unmarshal(raw, &batch); err == nil {
		results := make([]map[string]interface{}, 0, len(batch))
		for _, item := range batch {
			result, err := h.service.HandleBrevoEvent(c.UserContext(), item)
			if err != nil {
				return utils.Error(c, fiber.StatusBadRequest, "Gagal memproses webhook Brevo", err.Error())
			}
			results = append(results, result)
		}
		return utils.Success(c, "Webhook Brevo diproses", results)
	}
	result, err := h.service.HandleBrevoEvent(c.UserContext(), json.RawMessage(raw))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal memproses webhook Brevo", err.Error())
	}
	return utils.Success(c, "Webhook Brevo diproses", result)
}

func (h *EmailHandler) SyncBrevoHistory(c *fiber.Ctx) error {
	days, err := strconv.Atoi(c.Query("days", "90"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Parameter days tidak valid", nil)
	}
	data, err := h.service.SyncBrevoHistory(c.UserContext(), days)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal sinkronisasi history Brevo", err.Error())
	}
	return utils.Success(c, "History Brevo berhasil disinkronkan", data)
}
