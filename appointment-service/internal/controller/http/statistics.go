package handlers

import (
	"errors"
	"net/http"
	"time"

	"appointment-service/internal/entity"
	"appointment-service/internal/usecase/statistics"
	"github.com/gin-gonic/gin"
)

var statisticsService *statistics.Service

func SetStatisticsService(service *statistics.Service) {
	statisticsService = service
}

func statisticsErrorResponse(c *gin.Context, err error) {
	switch {
	case errors.Is(err, statistics.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки запроса", "details": err.Error()})
	}
}

// GetStatisticsHandler возвращает статистику за указанный период
func GetStatisticsHandler(c *gin.Context) {
	type Request struct {
		StartDate string `json:"start_date" binding:"required"`
		EndDate   string `json:"end_date" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты для start_date"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты для end_date"})
		return
	}

	if statisticsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Statistics service is not configured"})
		return
	}

	stats, err := statisticsService.GetByPeriod(c.Request.Context(), startDate, endDate)
	if err != nil {
		statisticsErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, entity.Statistics{
		TotalVisits:        stats.TotalVisits,
		TotalEarnings:      stats.TotalEarnings,
		TotalServices:      stats.TotalServices,
		TotalSubscriptions: stats.TotalSubscriptions,
	})
}

// GetCurrentMonthStatisticsHandler отдаёт статистику за текущий месяц
func GetCurrentMonthStatisticsHandler(c *gin.Context) {
	if statisticsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Statistics service is not configured"})
		return
	}

	stats, err := statisticsService.GetCurrentMonth(c.Request.Context())
	if err != nil {
		statisticsErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, entity.Statistics{
		TotalVisits:        stats.TotalVisits,
		TotalEarnings:      stats.TotalEarnings,
		TotalServices:      stats.TotalServices,
		TotalSubscriptions: stats.TotalSubscriptions,
	})
}
