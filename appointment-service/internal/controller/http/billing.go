package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"appointment-service/internal/entity"
	"appointment-service/internal/usecase/billing"
	"github.com/gin-gonic/gin"
)

var billingService *billing.Service

func SetBillingService(service *billing.Service) {
	billingService = service
}

func billingErrorResponse(c *gin.Context, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, billing.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
	case errors.Is(err, billing.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage, "details": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки запроса", "details": err.Error()})
	}
}

// AddVisitTransaction выполняет транзакцию для добавления посещения по абонементу
func AddVisitTransaction(c *gin.Context) {
	type Request struct {
		SubscriptionID int       `json:"subscription_id" binding:"required"`
		AppointmentID  int       `json:"appointment_id" binding:"required"`
		VisitDate      time.Time `json:"visit_date" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	if billingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing service is not configured"})
		return
	}

	if err := billingService.AddSubscriptionVisit(c.Request.Context(), billing.AddSubscriptionVisitRequest{
		SubscriptionID: req.SubscriptionID,
		AppointmentID:  req.AppointmentID,
		VisitDate:      req.VisitDate,
	}); err != nil {
		billingErrorResponse(c, err, "Абонемент не найден")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Посещение добавлено успешно"})
}

// GetActiveSubscriptionHandler возвращает активный абонемент по client_id и service_id
func GetActiveSubscriptionHandler(c *gin.Context) {
	type Request struct {
		ClientID  int `json:"client_id" binding:"required"`
		ServiceID int `json:"service_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	if billingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing service is not configured"})
		return
	}

	result, err := billingService.GetActiveSubscription(c.Request.Context(), req.ClientID, req.ServiceID)
	if err != nil {
		billingErrorResponse(c, err, "Активный абонемент не найден")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription_id": result.SubscriptionID,
		"current_balance": result.CurrentBalance,
	})
}

// GetSubscriptionsHandler возвращает список абонементов клиента
func GetSubscriptionsHandler(c *gin.Context) {
	clientIDParam := c.Query("client_id")
	if clientIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр client_id обязателен"})
		return
	}

	clientID, err := strconv.Atoi(clientIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат client_id", "details": err.Error()})
		return
	}

	if billingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing service is not configured"})
		return
	}

	subscriptions, err := billingService.GetClientSubscriptions(c.Request.Context(), clientID)
	if err != nil {
		billingErrorResponse(c, err, "Абонементы клиента не найдены")
		return
	}

	result := make([]map[string]interface{}, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		result = append(result, map[string]interface{}{
			"subscription_id": subscription.SubscriptionID,
			"current_balance": subscription.CurrentBalance,
		})
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": result})
}

// GetSubscriptionTypes возвращает типы абонементов
func GetSubscriptionTypes(c *gin.Context) {
	if billingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing service is not configured"})
		return
	}

	types, err := billingService.GetSubscriptionTypes(c.Request.Context())
	if err != nil {
		billingErrorResponse(c, err, "Типы абонементов не найдены")
		return
	}

	subscriptionTypes := make(map[string]entity.SubscriptionType, len(types))
	for _, subscriptionType := range types {
		subscriptionTypes[subscriptionType.Name] = entity.SubscriptionType{
			ID:            subscriptionType.ID,
			Name:          subscriptionType.Name,
			Cost:          subscriptionType.Cost,
			SessionsCount: subscriptionType.SessionsCount,
			ServiceIDs:    subscriptionType.ServiceIDs,
		}
	}

	c.JSON(http.StatusOK, subscriptionTypes)
}

// SellSubscription создаёт и продаёт абонемент клиенту
func SellSubscription(c *gin.Context) {
	var request struct {
		ClientID           int     `json:"client_id"`
		SubscriptionTypeID int     `json:"subscription_types_id"`
		Cost               float64 `json:"cost"`
		CurrentBalance     float64 `json:"sessions_count"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
		return
	}

	if billingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing service is not configured"})
		return
	}

	subscriptionTypeName, err := billingService.SellSubscription(c.Request.Context(), billing.SellSubscriptionRequest{
		ClientID:           request.ClientID,
		SubscriptionTypeID: request.SubscriptionTypeID,
		Cost:               request.Cost,
		CurrentBalance:     int(request.CurrentBalance),
	})
	if err != nil {
		billingErrorResponse(c, err, "Абонемент с указанным ID не найден")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Абонемент успешно продан и добавлен", "subscription": subscriptionTypeName})
}

// UpdatePaymentStatusMain обновляет статус оплаты записи
func UpdatePaymentStatusMain(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var appointment entity.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	if billingService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Billing service is not configured"})
		return
	}

	clientID := IntStringToInt(appointment.ClientID)
	paymentStatus := ""
	if appointment.PaymentStatus != nil {
		paymentStatus = *appointment.PaymentStatus
	}

	err = billingService.UpdateAppointmentPaymentStatus(c.Request.Context(), billing.UpdateAppointmentPaymentRequest{
		AppointmentID: id,
		ClientID:      clientID,
		PaymentStatus: paymentStatus,
		Amount:        appointment.Amount,
	})
	if err != nil {
		billingErrorResponse(c, err, "Запись не найдена")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус и данные обновлены успешно"})
}
