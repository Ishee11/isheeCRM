package handlers

import (
	"appointment-service/internal/entity"
	appointmentsuc "appointment-service/internal/usecase/appointments"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateVisit(c *gin.Context) {
	var appointment entity.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}
	if appointment.StartTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time обязателен"})
		return
	}
	if appointmentsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Appointments service is not configured"})
		return
	}

	result, err := appointmentsService.Create(c.Request.Context(), appointmentsuc.CreateRequest{
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
func GetVisits(c *gin.Context) {
	startParam := c.DefaultQuery("start", "")
	onlyUnpaidStr := c.DefaultQuery("unpaid", "")
	onlyUnpaid := onlyUnpaidStr == "true"

	var startTime *time.Time
	if startParam != "" {
		parsedStart, err := time.Parse("2006-01-02", startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра start"})
			return
		}
		startTime = &parsedStart
	}

	if appointmentsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Appointments service is not configured"})
		return
	}

	list, err := appointmentsService.List(c.Request.Context(), appointmentsuc.ListFilter{
		OnlyUnpaid: onlyUnpaid,
		Start:      startTime,
	})
	if err != nil {
		appointmentsErrorResponse(c, err, "Записи не найдены")
		return
	}
	result := make(map[string]entity.Appointment, len(list))
	for _, a := range list {
		key := fmt.Sprintf("%s в %s %s %s",
			a.StartTime.Format("2006.01.02"),
			a.StartTime.Format("15:04"),
			a.ServiceName,
			a.ClientName,
		)
		result[key] = entity.Appointment{
			ID:                uint(a.ID),
			ServiceID:         entity.IntString(a.ServiceID),
			ClientID:          entity.IntString(a.ClientID),
			StartTime:         &a.StartTime,
			PaymentStatus:     a.PaymentStatus,
			AppointmentStatus: a.AppointmentStatus,
			ServiceName:       a.ServiceName,
			ClientName:        a.ClientName,
		}
	}
	c.JSON(http.StatusOK, result)
}

func UpdateVisitStatus(c *gin.Context) {
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
	if appointmentsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Appointments service is not configured"})
		return
	}
	_, err = appointmentsService.UpdateStatus(c.Request.Context(), id, *payload.AppointmentStatus)
	if err != nil {
		appointmentsErrorResponse(c, err, "Запись не найдена")
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func MoveVisit(c *gin.Context) {
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
	if appointmentsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Appointments service is not configured"})
		return
	}
	if _, err := appointmentsService.Move(c.Request.Context(), id, *payload.StartTime); err != nil {
		appointmentsErrorResponse(c, err, "Запись не найдена")
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func DeleteVisit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	if appointmentsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Appointments service is not configured"})
		return
	}
	if err := appointmentsService.Delete(c.Request.Context(), id); err != nil {
		appointmentsErrorResponse(c, err, "Запись не найдена")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Удалено"})
}
