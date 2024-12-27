package handlers

import (
	"appointment-service/database"
	"appointment-service/models"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"strconv"
	"time"
)

// ExecuteQueryRow выполняет SQL-запрос с переменным количеством параметров, используя пул соединений pgx
func ExecuteQueryRow(ctx context.Context, pool *pgxpool.Pool, query string, args ...interface{}) (pgx.Row, error) {
	// Выполняем запрос с аргументами и пулом соединений pgx
	return pool.QueryRow(ctx, query, args...), nil
}

// Создать запись
func CreateAppointment(c *gin.Context) {
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		// Используем err.Error(), чтобы вернуть информацию об ошибке
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Неверные данные",
			"details": err.Error(),
		})
		return
	}

	query := `
		INSERT INTO appointments (service_id, client_id, start_time)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := database.Pool.QueryRow(context.Background(), query,
		appointment.ServiceID, appointment.ClientID,
		appointment.StartTime).Scan(&appointment.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Не удалось создать запись",
			"details": map[string]interface{}{
				"service_id": appointment.ServiceID,
				"client_id":  appointment.ClientID,
				"start_time": appointment.StartTime,
			},
		})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

// Получить список записей (если передан интервал - ищем записи в нем, если нет - только будующие записи)
func GetAppointments(c *gin.Context) {
	// Получаем параметры интервала (если они есть)
	startParam := c.DefaultQuery("start", "") // Параметр start, по умолчанию пустая строка
	endParam := c.DefaultQuery("end", "")     // Параметр end, по умолчанию пустая строка

	// Если передан параметр "start", пытаемся его преобразовать в дату
	var startTime time.Time
	var endTime time.Time
	var err error

	// Если startParam не пустой, то парсим дату
	if startParam != "" {
		startTime, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра start"})
			return
		}
	} else {
		// Если start не передан, по умолчанию берём текущее время
		startTime = time.Now()
	}

	// Если endParam не пустой, пытаемся его преобразовать в дату
	if endParam != "" {
		endTime, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра end"})
			return
		}
	} else {
		// Если end не передан, по умолчанию ставим очень далекую дату (или не ставим ограничений)
		endTime = time.Now().Add(24 * time.Hour * 365) // через год
	}

	// Формируем запрос для фильтрации по времени
	query := `
		SELECT id, service_id, client_id, start_time
		FROM appointments
		WHERE start_time >= $1 AND start_time <= $2
	`

	rows, err := database.Pool.Query(context.Background(), query, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить записи"})
		return
	}
	defer rows.Close()

	// Создаём словарь для хранения записей
	appointmentsMap := make(map[string]models.Appointment)
	for rows.Next() {
		var appointment models.Appointment
		if err := rows.Scan(
			&appointment.ID, &appointment.ServiceID, &appointment.ClientID,
			&appointment.StartTime,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных"})
			return
		}
		// Добавляем запись в словарь с ключом = ID
		key := fmt.Sprintf("%v %v", appointment.ServiceID, appointment.StartTime)
		appointmentsMap[key] = appointment
	}

	c.JSON(http.StatusOK, appointmentsMap)
}

// Перенести запись по ID на другое время
func UpdateAppointment(c *gin.Context) {
	// Читаем ID записи из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Читаем данные из тела запроса
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Выполняем обновление записи
	query := `
		UPDATE appointments
		SET start_time = $1
		WHERE id = $2
		RETURNING id, start_time;
	`
	// row := database.Pool.QueryRow(context.Background(), query, appointment.StartTime, id)

	// Выполняем запрос с использованием утилитной функции и передаем пул соединений pgx
	row, err := ExecuteQueryRow(context.Background(), database.Pool, query, appointment.StartTime, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось выполнить запрос"})
		return
	}

	var updatedAppointment models.Appointment
	if err := row.Scan(&updatedAppointment.ID, &updatedAppointment.StartTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить запись"})
		return
	}

	// Возвращаем обновленную запись
	c.JSON(http.StatusOK, updatedAppointment)
}

// Удалить запись по ID
func DeleteAppointment(c *gin.Context) {
	// Читаем ID записи из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Выполняем запрос на удаление записи
	query := `DELETE FROM appointments WHERE id = $1`

	_, err = database.Pool.Exec(context.Background(), query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить запись"})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "Запись успешно удалена"})
}
