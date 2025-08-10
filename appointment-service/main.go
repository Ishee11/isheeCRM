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
	router.StaticFile("/", "./test.html")
	/*router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})*/

	// Подключение к базе данных
	err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer database.Close() // Закрываем соединение при завершении программы

	// Настраиваем обработчик для добавления активности
	// ЗАПИСИ
	visitsGroup := router.Group("/visits")
	{
		visitsGroup.POST("/", handlers.CreateVisit)                // создание записи
		visitsGroup.GET("/", handlers.GetVisits)                   // получение списка записей с фильтрами
		visitsGroup.PUT("/status/:id", handlers.UpdateVisitStatus) // обновление статуса записи
		visitsGroup.PUT("/move/:id", handlers.MoveVisit)           // перенос записи
		visitsGroup.DELETE("/:id", handlers.DeleteVisit)           // удаление записи
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
		subscriptionsGroup.POST("/add", handlers.AddSubscriptionType)             // добавить тип абонемента
		subscriptionsGroup.GET("/types", handlers.GetSubscriptionTypes)           // получить список типов абонементов
		subscriptionsGroup.GET("/client", handlers.GetSubscriptionsHandler)       // получение списка абонементов клиента
	}

	// УСЛУГИ
	serviceGroup := router.Group("/services")
	{
		serviceGroup.POST("/add", handlers.AddService)      // добавить услугу
		serviceGroup.GET("/", handlers.GetServices)         // список услуг
		serviceGroup.DELETE("/:id", handlers.DeleteService) // удалить услугу
	}

	// ОПЛАТА
	paymentsGroup := router.Group("/payments")
	{
		paymentsGroup.POST("/subscription", handlers.AddVisitTransaction)  // оплата абонементом
		paymentsGroup.PUT("/visits/:id", handlers.UpdatePaymentStatusMain) // обновление статуса оплаты для записи
	}

	// СТАТИСТИКА
	statisticsGroup := router.Group("/statistics")
	{
		statisticsGroup.POST("/", handlers.GetStatisticsHandler)                         // статистика
		statisticsGroup.GET("/current-month", handlers.GetCurrentMonthStatisticsHandler) // за этот месяц
	}

	// Запуск сервера
	fmt.Println("Сервер запущен на :8080")
	log.Fatal(router.Run("0.0.0.0:8080"))
}
