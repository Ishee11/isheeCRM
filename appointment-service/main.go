package main

import (
	"appointment-service/database"
	"appointment-service/handlers"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	router := gin.Default()

	// Подключение к базе данных
	err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer database.Close() // Закрываем соединение при завершении программы

	// Настраиваем обработчик для добавления активности
	// ЗАПИСИ
	appointmentsGroup := router.Group("/appointments")
	{
		appointmentsGroup.POST("/", handlers.CreateAppointment)                // создание записи
		appointmentsGroup.GET("/", handlers.GetAppointments)                   // получение списка записей с фильтрами
		appointmentsGroup.PUT("/status/:id", handlers.UpdateAppointmentStatus) // обновление статуса записи
		appointmentsGroup.PUT("/move/:id", handlers.MoveAppointment)           // перенос записи
		appointmentsGroup.DELETE("/:id", handlers.DeleteAppointment)           // удаление записи
	}

	// КЛИЕНТЫ
	clientsGroup := router.Group("/clients")
	{
		clientsGroup.POST("/", handlers.CreateClient)                // создание клиента
		clientsGroup.GET("/find", handlers.FindClientByPhoneHandler) // поиск клиента по номеру телефона
		clientsGroup.GET("/info", handlers.GetClientInfoHandler)     // получение информации о клиенте по id
	}

	// АБОНЕМЕНТЫ
	subscriptionsGroup := router.Group("/subscriptions")
	{
		subscriptionsGroup.POST("/search", handlers.GetActiveSubscriptionHandler) // найти абонемент с положительным балансом
		subscriptionsGroup.POST("/sell", handlers.SellSubscription)               // продажа абонемента
		subscriptionsGroup.GET("/types", handlers.GetSubscriptionTypes)           // получить список типов абонементов
		subscriptionsGroup.GET("/client", handlers.GetSubscriptionsHandler)       // получение списка абонементов клиента
	}

	// ОПЛАТА
	paymentsGroup := router.Group("/payments")
	{
		paymentsGroup.POST("/subscription", handlers.AddVisitTransaction)       // оплата абонементом
		paymentsGroup.PUT("/appointment/:id", handlers.UpdatePaymentStatusMain) // обновление статуса оплаты для записи
	}

	// СТАТИСТИКА
	statisticsGroup := router.Group("/statistics")
	{
		statisticsGroup.POST("/", handlers.GetStatisticsHandler) // статистика
	}

	// Запуск сервера
	fmt.Println("Сервер запущен на :8080")
	log.Fatal(router.Run("0.0.0.0:8080"))
}
