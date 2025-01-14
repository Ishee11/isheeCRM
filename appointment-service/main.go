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
	// создание записи.
	router.POST("/appointments", handlers.CreateAppointment)
	// создание клиента.
	router.POST("/new_client", handlers.CreateClient)
	// найти абонемент с положительным балансом (по id услуги).
	router.POST("/search_subscription", handlers.GetActiveSubscriptionHandler)
	// оплата абонементом.
	router.POST("/pay_subscription", handlers.AddVisitTransaction)
	// продажа абонемента.
	router.POST("/sell_subscription", handlers.SellSubscription)
	// получить список типов абонементов.
	router.GET("/get_subscription_type", handlers.GetSubscriptionTypes)
	// статистика.
	router.POST("/statistics", handlers.GetStatisticsHandler)
	// получение списка записей с фильтрами.
	router.GET("/appointments", handlers.GetAppointments)
	// получение списка абонементов клиента.
	router.GET("/client_subscription", handlers.GetSubscriptionsHandler)
	// обновление записи.
	router.PUT("/appointments/status/:id", handlers.UpdateAppointmentStatus)
	// обновление статуса записи.
	router.PUT("/appointments/move/:id", handlers.MoveAppointment)
	// обновление статуса оплаты.
	router.PUT("/appointments/payment/:id", handlers.UpdatePaymentStatusMain)
	// удаление записи.
	router.DELETE("/appointments/:id", handlers.DeleteAppointment)
	// получение id клиента по номеру.
	router.GET("/clients/find", handlers.FindClientByPhoneHandler)
	// получение информации о клиенте по id.
	router.GET("/clients/info", handlers.GetClientInfoHandler)

	// Запуск сервера
	fmt.Println("Сервер запущен на :8080")
	log.Fatal(router.Run("0.0.0.0:8080"))
}

/*package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Appointment struct {
	ID                string     `json:"id"`
	StartTime         *time.Time `json:"start_time"`         // Используем указатель для времени
	PaymentStatus     *string    `json:"payment_status"`     // Используем указатель для статуса оплаты
	AppointmentStatus *string    `json:"appointment_status"` // Используем указатель для статуса записи
	Amount            string     `json:"amount"`             // Сумма оплаты
	ServiceName       string     `json:"service_name"`
	ClientName        string     `json:"client_name"`
}

func main() {
	r := gin.Default()

	r.POST("/echo", func(c *gin.Context) {
		var updatedAppointment Appointment
		if err := c.BindJSON(&updatedAppointment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		fmt.Println(updatedAppointment)
		c.JSON(http.StatusOK, gin.H{
			"received": updatedAppointment,
		})
	})

	r.Run(":8080") // Сервер запустится на http://localhost:8080
}
*/
