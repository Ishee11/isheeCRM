package handlers

import (
	"errors"
	"net/http"

	"appointment-service/internal/entity"
	appointmentsuc "appointment-service/internal/usecase/appointments"
	"github.com/gin-gonic/gin"
)

var appointmentsService *appointmentsuc.Service

func SetAppointmentsService(service *appointmentsuc.Service) {
	appointmentsService = service
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
