package handlers

import (
	"fmt"
	"graduation-invitation/internal/models"
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type EventSettingsHandler struct {
	service *services.EventSettingsService
	storage *services.R2StorageService
}

type uploadedAsset struct {
	object *services.UploadedObject
	title  string
}

func NewEventSettingsHandler(service *services.EventSettingsService, storage *services.R2StorageService) *EventSettingsHandler {
	return &EventSettingsHandler{service: service, storage: storage}
}

func (h *EventSettingsHandler) Get(c *fiber.Ctx) error {
	data, err := h.service.Get(c.Context())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil template undangan", err.Error())
	}
	return utils.Success(c, "Template undangan", data)
}

func (h *EventSettingsHandler) List(c *fiber.Ctx) error {
	data, err := h.service.List(c.Context())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil template undangan", err.Error())
	}
	return utils.Success(c, "Daftar template undangan", data)
}

func (h *EventSettingsHandler) Create(c *fiber.Ctx) error {
	var input models.EventSettings
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.Create(c.Context(), input)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal membuat template undangan", err.Error())
	}
	return utils.Created(c, "Template undangan berhasil dibuat", data)
}

func (h *EventSettingsHandler) Update(c *fiber.Ctx) error {
	var input models.EventSettings
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	if id := c.Params("id"); id != "" {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
		}
		input.ID = parsed
	}
	data, err := h.service.Update(c.Context(), input)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menyimpan template undangan", err.Error())
	}
	return utils.Success(c, "Template undangan berhasil disimpan", data)
}

func (h *EventSettingsHandler) UpdateSeatMap(c *fiber.Ctx) error {
	var input struct {
		SeatMapColumns   int    `json:"seat_map_columns"`
		SeatMapColorMode string `json:"seat_map_color_mode"`
		SeatMapLayout    string `json:"seat_map_layout"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	data, err := h.service.UpdateSeatMap(c.Context(), input.SeatMapColumns, input.SeatMapColorMode, input.SeatMapLayout)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menyimpan denah bangku", err.Error())
	}
	return utils.Success(c, "Denah bangku berhasil disimpan", data)
}

func (h *EventSettingsHandler) UploadLogo(c *fiber.Ctx) error {
	uploaded, err := h.uploadObject(c, "file", "logo sekolah", 5*1024*1024, map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".webp": "image/webp",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
	}, "graduation/logo")
	if err != nil {
		return err
	}
	data, err := h.service.UpdateSchoolLogo(c.Context(), uploaded.object.URL, uploaded.object.Key)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Logo berhasil diupload, tetapi gagal disimpan ke template", err.Error())
	}
	return utils.Success(c, "Logo sekolah berhasil diupload", data)
}

func (h *EventSettingsHandler) Activate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	data, err := h.service.Activate(c.Context(), id)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengaktifkan template undangan", err.Error())
	}
	return utils.Success(c, "Template undangan aktif", data)
}

func (h *EventSettingsHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	if err := h.service.Delete(c.Context(), id); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menghapus template undangan", err.Error())
	}
	return utils.Success(c, "Template undangan berhasil dihapus", nil)
}

func (h *EventSettingsHandler) UploadAudio(c *fiber.Ctx) error {
	uploaded, err := h.uploadObject(c, "file", "MP3", 25*1024*1024, map[string]string{
		".mp3": "audio/mpeg",
	}, "graduation/audio")
	if err != nil {
		return err
	}
	title := uploaded.title
	return utils.Success(c, "MP3 berhasil diupload", fiber.Map{
		"audio_url":   uploaded.object.URL,
		"audio_key":   uploaded.object.Key,
		"audio_title": title,
	})
}

func (h *EventSettingsHandler) uploadObject(c *fiber.Ctx, formField string, label string, maxSize int64, allowedExt map[string]string, prefix string) (*uploadedAsset, error) {
	file, err := c.FormFile(formField)
	if err != nil {
		return nil, utils.Error(c, fiber.StatusBadRequest, "File "+label+" wajib diunggah", err.Error())
	}
	if file.Size <= 0 {
		return nil, utils.Error(c, fiber.StatusBadRequest, "File "+label+" kosong", nil)
	}
	if file.Size > maxSize {
		return nil, utils.Error(c, fiber.StatusBadRequest, "Ukuran file "+label+" terlalu besar", nil)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	contentType, ok := allowedExt[ext]
	if !ok {
		return nil, utils.Error(c, fiber.StatusBadRequest, "Format file "+label+" tidak didukung", nil)
	}

	src, err := file.Open()
	if err != nil {
		return nil, utils.Error(c, fiber.StatusBadRequest, "Gagal membaca file", err.Error())
	}
	defer src.Close()

	body, err := io.ReadAll(io.LimitReader(src, maxSize+1))
	if err != nil {
		return nil, utils.Error(c, fiber.StatusBadRequest, "Gagal membaca file", err.Error())
	}
	if int64(len(body)) > maxSize {
		return nil, utils.Error(c, fiber.StatusBadRequest, "Ukuran file "+label+" terlalu besar", nil)
	}

	if contentType == "" {
		contentType = mime.TypeByExtension(ext)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	key := fmt.Sprintf("%s/%s-%s%s", prefix, time.Now().UTC().Format("20060102150405"), uuid.NewString(), ext)
	uploaded, err := h.storage.PutObject(c.Context(), key, contentType, body)
	if err != nil {
		return nil, utils.Error(c, fiber.StatusInternalServerError, "Gagal upload file ke R2", err.Error())
	}
	return &uploadedAsset{
		object: uploaded,
		title:  strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename)),
	}, nil
}
