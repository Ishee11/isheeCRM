package handlers

import (
	"errors"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/entity"
	appointmentsuc "github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/appointments"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AppointmentsHandler struct {
	service *appointmentsuc.Service
}

func NewAppointmentsHandler(service *appointmentsuc.Service) *AppointmentsHandler {
	return &AppointmentsHandler{service: service}
}

func appointmentsErrorResponse(c *gin.Context, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, appointmentsuc.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
	case errors.Is(err, appointmentsuc.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage, "details": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки запроса", "details": err.Error()})
	}
}

// IntStringToInt конвертирует кастомный тип в int
func IntStringToInt(value entity.IntString) int {
	return int(value)
}

func (h *AppointmentsHandler) CreateVisit(c *gin.Context) {
	var appointment entity.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}
	if appointment.StartTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time обязателен"})
		return
	}

	result, err := h.service.Create(c.Request.Context(), appointmentsuc.CreateRequest{
		ServiceID: int(appointment.ServiceID),
		ClientID:  int(appointment.ClientID),
		StartTime: *appointment.StartTime,
	})
	if err != nil {
		appointmentsErrorResponse(c, err, "Запись не добавлена")
		return
	}
	appointment.ID = uint(result.ID)
	c.JSON(http.StatusCreated, appointment)
}
func (h *AppointmentsHandler) GetVisits(c *gin.Context) {
	startParam := c.DefaultQuery("start", "")
	fromParam := c.DefaultQuery("from", "")
	toParam := c.DefaultQuery("to", "")
	clientIDParam := c.DefaultQuery("client_id", "")
	status := c.DefaultQuery("status", "")
	onlyUnpaidStr := c.DefaultQuery("unpaid", "")
	onlyUnpaid := onlyUnpaidStr == "true"

	var fromTime *time.Time
	if fromParam == "" {
		fromParam = startParam
	}
	if fromParam != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра from"})
			return
		}
		fromTime = &parsedFrom
	}

	var toTime *time.Time
	if toParam != "" {
		parsedTo, err := time.Parse("2006-01-02", toParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра to"})
			return
		}
		toTime = &parsedTo
	}

	var clientID int
	if clientIDParam != "" {
		parsedClientID, err := strconv.Atoi(clientIDParam)
		if err != nil || parsedClientID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра client_id"})
			return
		}
		clientID = parsedClientID
	}

	list, err := h.service.List(c.Request.Context(), appointmentsuc.ListFilter{
		OnlyUnpaid:        onlyUnpaid,
		ClientID:          clientID,
		AppointmentStatus: status,
		From:              fromTime,
		To:                toTime,
	})
	if err != nil {
		appointmentsErrorResponse(c, err, "Записи не найдены")
		return
	}
	result := make([]entity.Appointment, 0, len(list))
	for _, a := range list {
		result = append(result, entity.Appointment{
			ID:                uint(a.ID),
			ServiceID:         entity.IntString(a.ServiceID),
			ClientID:          entity.IntString(a.ClientID),
			StartTime:         &a.StartTime,
			PaymentStatus:     a.PaymentStatus,
			AppointmentStatus: a.AppointmentStatus,
			ServiceName:       a.ServiceName,
			ClientName:        a.ClientName,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"items": result,
		"total": len(result),
	})
}

func (h *AppointmentsHandler) UpdateVisitStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	var payload struct {
		AppointmentStatus *string `json:"appointment_status"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil || payload.AppointmentStatus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Требуется appointment_status"})
		return
	}

	_, err = h.service.UpdateStatus(c.Request.Context(), id, *payload.AppointmentStatus)
	if err != nil {
		appointmentsErrorResponse(c, err, "Запись не найдена")
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *AppointmentsHandler) MoveVisit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	var payload struct {
		StartTime *time.Time `json:"start_time"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil || payload.StartTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нужно start_time"})
		return
	}

	if _, err := h.service.Move(c.Request.Context(), id, *payload.StartTime); err != nil {
		appointmentsErrorResponse(c, err, "Запись не найдена")
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *AppointmentsHandler) DeleteVisit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		appointmentsErrorResponse(c, err, "Запись не найдена")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Удалено"})
}
