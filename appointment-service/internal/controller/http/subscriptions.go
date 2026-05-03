package handlers

import (
	"errors"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/subscriptions"
	"net/http"

	"github.com/gin-gonic/gin"
)

var subscriptionService *subscriptions.Service

func SetSubscriptionService(service *subscriptions.Service) {
	subscriptionService = service
}

func AddSubscriptionType(c *gin.Context) {
	var req struct {
		Name          string  `json:"name" binding:"required"`
		Cost          float64 `json:"cost" binding:"required"`
		SessionsCount int     `json:"sessions_count" binding:"required"`
		ServiceIDs    []int   `json:"service_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	if subscriptionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Subscription service is not configured"})
		return
	}

	created, err := subscriptionService.Create(c.Request.Context(), subscriptions.CreateRequest{
		Name:          req.Name,
		Cost:          req.Cost,
		SessionsCount: req.SessionsCount,
		ServiceIDs:    req.ServiceIDs,
	})
	if err != nil {
		switch {
		case errors.Is(err, subscriptions.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении абонемента", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id":             created.ID,
		"name":           created.Name,
		"cost":           created.Cost,
		"sessions_count": created.SessionsCount,
		"service_ids":    created.ServiceIDs,
	})
}
