package handlers

import (
	"errors"
	"fmt"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/entity"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var servicesService *services.Service

func SetServicesService(service *services.Service) {
	servicesService = service
}

func AddService(c *gin.Context) {
	var req entity.Service
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if servicesService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Services service is not configured"})
		return
	}

	created, err := servicesService.Create(c.Request.Context(), services.CreateRequest{
		Name:     req.Name,
		Duration: req.Duration,
		Price:    req.Price,
	})
	if err != nil {
		servicesErrorResponse(c, err, "Услуга не найдена")
		return
	}

	c.JSON(http.StatusOK, entity.Service{
		ID:       created.ID,
		Name:     created.Name,
		Duration: created.Duration,
		Price:    created.Price,
	})
}

func GetServices(c *gin.Context) {
	if servicesService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Services service is not configured"})
		return
	}

	list, err := servicesService.List(c.Request.Context())
	if err != nil {
		servicesErrorResponse(c, err, "Услуги не найдены")
		return
	}

	result := make([]entity.Service, 0, len(list))
	legacyResult := make(map[string]entity.Service, len(list))
	for _, service := range list {
		item := entity.Service{
			ID:       service.ID,
			Name:     service.Name,
			Duration: service.Duration,
			Price:    service.Price,
		}
		result = append(result, item)
		key := service.Name
		if service.Duration > 0 && service.Price > 0 {
			key = fmt.Sprintf("%s - %d мин. %.0f руб.", service.Name, service.Duration, service.Price)
		}
		legacyResult[key] = item
	}

	if c.Query("format") != "list" {
		c.JSON(http.StatusOK, legacyResult)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": result,
		"total": len(result),
	})
}

func DeleteService(c *gin.Context) {
	if servicesService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Services service is not configured"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := servicesService.Delete(c.Request.Context(), id); err != nil {
		servicesErrorResponse(c, err, "Услуга не найдена")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Услуга успешно удалена"})
}

func servicesErrorResponse(c *gin.Context, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, services.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
	case errors.Is(err, services.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage, "details": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки запроса", "details": err.Error()})
	}
}
