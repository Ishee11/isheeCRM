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
	// получение списка записей с фильтрами.
	router.GET("/appointments", handlers.GetAppointments)
	// обновление записи.
	router.PUT("/appointments/:id", handlers.UpdateAppointment)
	// удаление записи.
	router.DELETE("/appointments/:id", handlers.DeleteAppointment)

	// Запуск сервера
	fmt.Println("Сервер запущен на :8080")
	log.Fatal(router.Run("0.0.0.0:8080"))
}
