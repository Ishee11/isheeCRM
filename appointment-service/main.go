package main

import (
	"appointment-service/database"
	controllerhttp "appointment-service/internal/controller/http"
	postgresrepo "appointment-service/internal/repository/postgres"
	appointmentsuc "appointment-service/internal/usecase/appointments"
	"appointment-service/internal/usecase/billing"
	clientsuc "appointment-service/internal/usecase/clients"
	servicesuc "appointment-service/internal/usecase/services"
	statisticsuc "appointment-service/internal/usecase/statistics"
	subscriptionsuc "appointment-service/internal/usecase/subscriptions"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	router := gin.Default()
	router.StaticFile("/", "./test.html")
	healthHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	}
	router.GET("/healthz", healthHandler)
	router.HEAD("/healthz", healthHandler)
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"image_tag": getenv("IMAGE_TAG", "unknown"),
			"app_image": getenv("APP_IMAGE", "unknown"),
		})
	})
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
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer database.Close() // Закрываем соединение при завершении программы

	billingRepository := postgresrepo.NewBillingRepository(database.Pool)
	billingService := billing.NewService(billing.Dependencies{
		TxManager:                  billingRepository,
		SubscriptionVisitWriter:    billingRepository,
		SubscriptionBalanceRepo:    billingRepository,
		ActiveSubscriptionFinder:   billingRepository,
		ClientSubscriptionLister:   billingRepository,
		SubscriptionTypeCatalog:    billingRepository,
		SubscriptionSeller:         billingRepository,
		PaymentStatusUpdater:       billingRepository,
		AppointmentPaymentOperator: billingRepository,
		PaymentRollbackRepo:        billingRepository,
	})
	appointmentsRepository := postgresrepo.NewAppointmentsRepository(database.Pool)
	appointmentsService := appointmentsuc.NewService(appointmentsRepository)
	clientsRepository := postgresrepo.NewClientsRepository(database.Pool)
	clientsService := clientsuc.NewService(clientsRepository)
	servicesRepository := postgresrepo.NewServicesRepository(database.Pool)
	servicesService := servicesuc.NewService(servicesRepository)
	statisticsRepository := postgresrepo.NewStatisticsRepository(database.Pool)
	statisticsService := statisticsuc.NewService(statisticsRepository)
	subscriptionsRepository := postgresrepo.NewSubscriptionsRepository(database.Pool)
	subscriptionsService := subscriptionsuc.NewService(subscriptionsRepository)
	controllerhttp.SetBillingService(billingService)
	controllerhttp.SetAppointmentsService(appointmentsService)
	controllerhttp.SetClientsService(clientsService)
	controllerhttp.SetServicesService(servicesService)
	controllerhttp.SetStatisticsService(statisticsService)
	controllerhttp.SetSubscriptionService(subscriptionsService)

	// Настраиваем обработчик для добавления активности
	// ЗАПИСИ
	visitsGroup := router.Group("/visits")
	{
		visitsGroup.POST("/", controllerhttp.CreateVisit)                // создание записи
		visitsGroup.GET("/", controllerhttp.GetVisits)                   // получение списка записей с фильтрами
		visitsGroup.PUT("/status/:id", controllerhttp.UpdateVisitStatus) // обновление статуса записи
		visitsGroup.PUT("/move/:id", controllerhttp.MoveVisit)           // перенос записи
		visitsGroup.DELETE("/:id", controllerhttp.DeleteVisit)           // удаление записи
	}

	// КЛИЕНТЫ
	clientsGroup := router.Group("/clients")
	{
		clientsGroup.POST("/", controllerhttp.CreateClient)                // создание клиента
		clientsGroup.GET("/find", controllerhttp.FindClientByPhoneHandler) // поиск клиента по номеру телефона
		clientsGroup.GET("/info", controllerhttp.GetClientInfoHandler)     // получение информации о клиенте по id
	}

	// АБОНЕМЕНТЫ
	subscriptionsGroup := router.Group("/subscriptions")
	{
		subscriptionsGroup.POST("/search", controllerhttp.GetActiveSubscriptionHandler) // найти абонемент с положительным балансом
		subscriptionsGroup.POST("/sell", controllerhttp.SellSubscription)               // продажа абонемента
		subscriptionsGroup.POST("/add", controllerhttp.AddSubscriptionType)             // добавить тип абонемента
		subscriptionsGroup.GET("/types", controllerhttp.GetSubscriptionTypes)           // получить список типов абонементов
		subscriptionsGroup.GET("/client", controllerhttp.GetSubscriptionsHandler)       // получение списка абонементов клиента
	}

	// УСЛУГИ
	serviceGroup := router.Group("/services")
	{
		serviceGroup.POST("/add", controllerhttp.AddService)      // добавить услугу
		serviceGroup.GET("/", controllerhttp.GetServices)         // список услуг
		serviceGroup.DELETE("/:id", controllerhttp.DeleteService) // удалить услугу
	}

	// ОПЛАТА
	paymentsGroup := router.Group("/payments")
	{
		paymentsGroup.POST("/subscription", controllerhttp.AddVisitTransaction)  // оплата абонементом
		paymentsGroup.PUT("/visits/:id", controllerhttp.UpdatePaymentStatusMain) // обновление статуса оплаты для записи
	}

	// СТАТИСТИКА
	statisticsGroup := router.Group("/statistics")
	{
		statisticsGroup.POST("/", controllerhttp.GetStatisticsHandler)                         // статистика
		statisticsGroup.GET("/current-month", controllerhttp.GetCurrentMonthStatisticsHandler) // за этот месяц
	}

	// Запуск сервера
	fmt.Println("Запуск сервера на :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
