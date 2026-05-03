package handlers

import (
	"errors"
	"net/http"
	"strconv"

	clientsuc "github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/clients"
	"github.com/gin-gonic/gin"
)

var clientsService *clientsuc.Service

func SetClientsService(service *clientsuc.Service) {
	clientsService = service
}

func clientsErrorResponse(c *gin.Context, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, clientsuc.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
	case errors.Is(err, clientsuc.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage, "details": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки запроса", "details": err.Error()})
	}
}

// CreateClient создание клиента
func CreateClient(c *gin.Context) {
	// Структура для входящих данных
	type Request struct {
		Phone string `json:"phone" binding:"required"`
		Name  string `json:"name" binding:"required"`
	}

	var req Request

	// Парсим входящие данные
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	if clientsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Clients service is not configured"})
		return
	}

	clientID, normalizedPhone, err := clientsService.Create(c.Request.Context(), req.Phone, req.Name)
	if err != nil {
		clientsErrorResponse(c, err, "Не удалось создать клиента")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Клиент успешно создан",
		"client_id": clientID,
		"phone":     normalizedPhone,
	})
}

// FindClientByPhoneHandler поиск клиента по номеру телефона (возвращает id клиента)
func FindClientByPhoneHandler(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Номер телефона обязателен",
			"details": "Параметр phone не передан или пустой",
		})
		return
	}

	if clientsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Clients service is not configured"})
		return
	}

	clientID, normalizedPhone, err := clientsService.FindByPhone(c.Request.Context(), phone)
	if err != nil {
		clientsErrorResponse(c, err, "Клиент не найден")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id": clientID,
		"phone":     normalizedPhone,
	})
}

// GetClientInfoHandler - хендлер для получения информации о клиенте
func GetClientInfoHandler(c *gin.Context) {
	clientIDQuery := c.Query("client_id")
	if clientIDQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID клиента обязателен",
			"details": "ID клиента не передан или пустой",
		})
		return
	}

	clientID, err := strconv.Atoi(clientIDQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат client_id"})
		return
	}

	if clientsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Clients service is not configured"})
		return
	}

	clientInfo, err := clientsService.GetInfo(c.Request.Context(), clientID)
	if err != nil {
		clientsErrorResponse(c, err, "Клиент не найден")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"name":        clientInfo.Name,
		"phone":       clientInfo.Phone,
		"email":       clientInfo.Email,
		"categories":  clientInfo.Categories,
		"birth_date":  clientInfo.BirthDate,
		"paid":        clientInfo.Paid,
		"spent":       clientInfo.Spent,
		"gender":      clientInfo.Gender,
		"discount":    clientInfo.Discount,
		"last_visit":  clientInfo.LastVisit,
		"first_visit": clientInfo.FirstVisit,
		"visit_count": clientInfo.VisitCount,
		"comment":     clientInfo.Comment,
	})
}
