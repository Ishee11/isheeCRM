package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func Run() error {
	router := newRouter()

	if err := database.ConnectDB(); err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer database.Close()

	setupServices()

	addr := fmt.Sprintf(":%s", getenv("APP_PORT", "8080"))
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	errs := make(chan error, 1)
	go func() {
		fmt.Printf("Запуск сервера на %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errs <- err
			return
		}
		errs <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		fmt.Printf("получен сигнал %s, завершение...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	case err := <-errs:
		return err
	}
}

func newRouter() *gin.Engine {
	router := gin.Default()
	router.StaticFile("/", "./test.html")
	router.GET("/healthz", healthHandler())
	router.HEAD("/healthz", healthHandler())
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"image_tag": getenv("IMAGE_TAG", "unknown"),
			"app_image": getenv("APP_IMAGE", "unknown"),
		})
	})
	setupRoutes(router)
	return router
}

func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	}
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

func setupServices() {
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
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
