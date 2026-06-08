package handlers

import (
	"strconv"

	"graduation-invitation/internal/models"
	"graduation-invitation/internal/services"
	"graduation-invitation/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentHandler struct {
	students *services.StudentService
	seats    *services.SeatService
}

func NewStudentHandler(students *services.StudentService, seats *services.SeatService) *StudentHandler {
	return &StudentHandler{students: students, seats: seats}
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	data, err := h.students.List(c.UserContext(), models.StudentFilter{
		ClassName:        c.Query("class_name"),
		Major:            c.Query("major"),
		AttendanceStatus: c.Query("attendance_status"),
		Search:           c.Query("search"),
		Page:             queryInt(c, "page", 1),
		Limit:            queryInt(c, "limit", 10),
	})
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data siswa", err.Error())
	}
	return utils.Success(c, "Data siswa", data)
}

func (h *StudentHandler) SeatMap(c *fiber.Ctx) error {
	data, err := h.students.SeatMap(c.UserContext(), models.StudentFilter{
		ClassName:        c.Query("class_name"),
		Major:            c.Query("major"),
		AttendanceStatus: c.Query("attendance_status"),
		Search:           c.Query("search"),
	})
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil denah bangku", err.Error())
	}
	return utils.Success(c, "Denah bangku", data)
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var input models.StudentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	student, err := h.students.Create(c.UserContext(), input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return utils.Created(c, "Siswa berhasil dibuat", student)
}

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	student, err := h.students.Get(c.UserContext(), id)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil siswa", err.Error())
	}
	if student == nil {
		return utils.Error(c, fiber.StatusNotFound, "Siswa tidak ditemukan", nil)
	}
	return utils.Success(c, "Detail siswa", student)
}

func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	var input models.StudentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request tidak valid", err.Error())
	}
	student, err := h.students.Update(c.UserContext(), id, input)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if student == nil {
		return utils.Error(c, fiber.StatusNotFound, "Siswa tidak ditemukan", nil)
	}
	return utils.Success(c, "Siswa berhasil diperbarui", student)
}

func (h *StudentHandler) UpdateAttendanceStatus(c *fiber.Ctx) error {
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
	student, err := h.students.UpdateAttendanceStatus(c.UserContext(), id, body.AttendanceStatus)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if student == nil {
		return utils.Error(c, fiber.StatusNotFound, "Siswa tidak ditemukan", nil)
	}
	return utils.Success(c, "Status absensi berhasil diperbarui", student)
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "ID tidak valid", nil)
	}
	if err := h.students.Delete(c.UserContext(), id); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menghapus siswa", err.Error())
	}
	return utils.Success(c, "Siswa berhasil dihapus", nil)
}

func (h *StudentHandler) ResetAll(c *fiber.Ctx) error {
	deleted, err := h.students.ResetAll(c.UserContext())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal reset data siswa", err.Error())
	}
	return utils.Success(c, "Semua data siswa berhasil direset", fiber.Map{
		"deleted": deleted,
	})
}

func (h *StudentHandler) Import(c *fiber.Ctx) error {
	header, err := c.FormFile("file")
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "File import wajib diunggah", err.Error())
	}
	file, err := header.Open()
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Gagal membuka file", err.Error())
	}
	defer file.Close()
	rows, err := utils.ParseImportFile(file, header.Filename)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	result, err := h.students.Import(c.UserContext(), rows)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), result)
	}
	return utils.Success(c, "Import siswa berhasil", result)
}

func (h *StudentHandler) Export(c *fiber.Ctx) error {
	data, err := h.students.ExportAttendanceXLSX(c.UserContext())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal export data", err.Error())
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=data-siswa-per-kelas.xlsx")
	c.Set("Content-Length", strconv.Itoa(len(data)))
	return c.Send(data)
}

func (h *StudentHandler) ImportTemplate(c *fiber.Ctx) error {
	data, err := h.students.ImportTemplateXLSX()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal membuat template import", err.Error())
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=template-import-siswa.xlsx")
	c.Set("Content-Length", strconv.Itoa(len(data)))
	return c.Send(data)
}

func (h *StudentHandler) RegenerateSeats(c *fiber.Ctx) error {
	if err := h.seats.RegenerateAllSeatNumbers(c.UserContext()); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal regenerate nomor bangku", err.Error())
	}
	return utils.Success(c, "Nomor bangku berhasil disusun ulang", nil)
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
}
