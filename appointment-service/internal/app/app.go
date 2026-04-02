package app

import (
	"fmt"
	"os"

	"appointment-service/database"
	controllerhttp "appointment-service/internal/controller/http"
	postgresrepo "appointment-service/internal/repository/postgres"
	appointmentsuc "appointment-service/internal/usecase/appointments"
	"appointment-service/internal/usecase/billing"
	clientsuc "appointment-service/internal/usecase/clients"
	servicesuc "appointment-service/internal/usecase/services"
	statisticsuc "appointment-service/internal/usecase/statistics"
	subscriptionsuc "appointment-service/internal/usecase/subscriptions"
	"github.com/gin-gonic/gin"
)

// Run bootstraps the HTTP stack, connects to the database and starts the server.
func Run() error {
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

	if err := database.ConnectDB(); err != nil {
		return fmt.Errorf("vault db: %w", err)
	}
	defer database.Close()

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

	setupRoutes(router)
	addr := fmt.Sprintf(":%s", getenv("APP_PORT", "8080"))
	fmt.Printf("Запуск сервера на %s\n", addr)
	return router.Run(addr)
}

func setupRoutes(router *gin.Engine) {
	visitsGroup := router.Group("/visits")
	{
		visitsGroup.POST("/", controllerhttp.CreateVisit)
		visitsGroup.GET("/", controllerhttp.GetVisits)
		visitsGroup.PUT("/status/:id", controllerhttp.UpdateVisitStatus)
		visitsGroup.PUT("/move/:id", controllerhttp.MoveVisit)
		visitsGroup.DELETE("/:id", controllerhttp.DeleteVisit)
	}

	clientsGroup := router.Group("/clients")
	{
		clientsGroup.POST("/", controllerhttp.CreateClient)
		clientsGroup.GET("/find", controllerhttp.FindClientByPhoneHandler)
		clientsGroup.GET("/info", controllerhttp.GetClientInfoHandler)
	}

	subscriptionsGroup := router.Group("/subscriptions")
	{
		subscriptionsGroup.POST("/search", controllerhttp.GetActiveSubscriptionHandler)
		subscriptionsGroup.POST("/sell", controllerhttp.SellSubscription)
		subscriptionsGroup.POST("/add", controllerhttp.AddSubscriptionType)
		subscriptionsGroup.GET("/types", controllerhttp.GetSubscriptionTypes)
		subscriptionsGroup.GET("/client", controllerhttp.GetSubscriptionsHandler)
	}

	serviceGroup := router.Group("/services")
	{
		serviceGroup.POST("/add", controllerhttp.AddService)
		serviceGroup.GET("/", controllerhttp.GetServices)
		serviceGroup.DELETE("/:id", controllerhttp.DeleteService)
	}

	paymentsGroup := router.Group("/payments")
	{
		paymentsGroup.POST("/subscription", controllerhttp.AddVisitTransaction)
		paymentsGroup.PUT("/visits/:id", controllerhttp.UpdatePaymentStatusMain)
	}

	statisticsGroup := router.Group("/statistics")
	{
		statisticsGroup.POST("/", controllerhttp.GetStatisticsHandler)
		statisticsGroup.GET("/current-month", controllerhttp.GetCurrentMonthStatisticsHandler)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
