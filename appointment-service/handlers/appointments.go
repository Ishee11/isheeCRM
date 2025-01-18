package handlers

import (
	"appointment-service/database"
	"appointment-service/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
	ЗАПИСЬ
	ОПЛАТА
	КЛИЕНТЫ
	АБОНЕМЕНТЫ
	СТАТИСТИКА
	ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
*/

// ЗАПИСЬ -------------------------------------------------------------------------------------------------------------

// Создать запись
func CreateVisit(c *gin.Context) {
	var appointment models.Appointment

	// Попытка привязки данных JSON к структуре
	if err := c.ShouldBindJSON(&appointment); err != nil {
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

	// Выполнение SQL-запроса
	err := database.Pool.QueryRow(context.Background(), query,
		appointment.ServiceID, appointment.ClientID,
		appointment.StartTime).Scan(&appointment.ID)

	// Если произошла ошибка
	if err != nil {
		var statusCode int
		var errorMessage string

		// Проверка типа ошибки
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "23503": // foreign_key_violation
				statusCode = http.StatusBadRequest
				errorMessage = "Указанная услуга или клиент не существуют"
			case "23514": // check_violation
				statusCode = http.StatusBadRequest
				errorMessage = "Некорректные данные (нарушение ограничения CHECK)"
			case "23505": // unique_violation
				statusCode = http.StatusConflict
				errorMessage = "Такая запись уже существует"
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Ошибка базы данных"
			}
		} else {
			statusCode = http.StatusInternalServerError
			errorMessage = "Не удалось создать запись"
		}

		// Возвращаем детализированный ответ об ошибке
		c.JSON(statusCode, gin.H{
			"error":   errorMessage,
			"details": err.Error(),
			"input": map[string]interface{}{
				"service_id": appointment.ServiceID,
				"client_id":  appointment.ClientID,
				"start_time": appointment.StartTime,
			},
		})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusCreated, appointment)
}

// Получить список записей (если передан интервал - ищем записи в нем, если нет - только будующие записи)
/*func GetVisits(c *gin.Context) {
	// Получаем параметры интервала (если они есть)
	startParam := c.DefaultQuery("start", "") // Параметр start, по умолчанию пустая строка
	endParam := c.DefaultQuery("end", "")     // Параметр end, по умолчанию пустая строка

	// Если передан параметр "start", пытаемся его преобразовать в дату
	var startTime time.Time
	var endTime time.Time
	var err error

	// Если startParam не пустой, то парсим дату
	// fmt.Println(startParam)
	if startParam != "" {
		startTime, err = time.Parse("2006-01-02", startParam)
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

		// Запрашиваем название услуги по ServiceID
		var serviceName string
		serviceQuery := `SELECT name FROM services WHERE service_id = $1`
		err = database.Pool.QueryRow(context.Background(), serviceQuery, appointment.ServiceID).Scan(&serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения названия услуги"})
			return
		}

		// Запрашиваем имя клиента по ClientID
		var clientName string
		clientQuery := `SELECT name FROM clients WHERE clients.clients_id = $1`
		err = database.Pool.QueryRow(context.Background(), clientQuery, appointment.ClientID).Scan(&clientName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения имени клиента"})
			return
		}

		// Добавляем запись в словарь с ключом = ID
		appointment.ServiceName = serviceName
		appointment.ClientName = clientName
		key := fmt.Sprintf("%v в %v \n%v \n%v", appointment.StartTime.Format("02.01.2006"), appointment.StartTime.Format("15:04"), appointment.ServiceName, appointment.ClientName)

		appointmentsMap[key] = appointment
		// fmt.Println(appointmentsMap)
	}
	fmt.Println(appointmentsMap)
	c.JSON(http.StatusOK, appointmentsMap)
}*/

// GetVisits получает записи с фильтрацией по статусу оплаты и/или времени (возвращает мапу)
/*func GetVisits(c *gin.Context) {
	// Получаем параметры фильтрации
	startParam := c.DefaultQuery("start", "")  // Параметр start (опционально)
	endParam := c.DefaultQuery("end", "")      // Параметр end (опционально)
	onlyUnpaid := c.DefaultQuery("unpaid", "") // Флаг, показывать только неоплаченные

	var startTime, endTime time.Time
	var err error

	// Обработка параметра start
	if startParam != "" {
		startTime, err = time.Parse("2006-01-02", startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра start"})
			return
		}
	} else {
		startTime = time.Time{} // Нулевое значение, если параметр отсутствует
	}

	// Обработка параметра end
	if endParam != "" {
		endTime, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра end"})
			return
		}
	} else {
		endTime = time.Now().Add(24 * time.Hour * 365) // Через год, если параметр отсутствует
	}

	// Начало запроса SQL
	query := `
		SELECT id, service_id, client_id, start_time, payment_status
		FROM appointments
		WHERE ($1::timestamptz IS NULL OR start_time >= $1)
		  AND ($2::timestamptz IS NULL OR start_time <= $2)
	`

	// Если требуется показать только неоплаченные записи
	if onlyUnpaid == "true" {
		query += " AND payment_status = 'unpaid'"
	}

	// Выполнение SQL-запроса
	rows, err := database.Pool.Query(context.Background(), query, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить записи"})
		return
	}
	defer rows.Close()

	// Создаем итоговый словарь
	result := make(map[string]models.Appointment)

	for rows.Next() {
		var appointment models.Appointment
		if err := rows.Scan(&appointment.ID, &appointment.ServiceID, &appointment.ClientID, &appointment.StartTime, &appointment.PaymentStatus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных"})
			return
		}

		// Получаем название услуги
		var serviceName string
		serviceQuery := `SELECT name FROM services WHERE service_id = $1`
		err = database.Pool.QueryRow(context.Background(), serviceQuery, appointment.ServiceID).Scan(&serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения названия услуги"})
			return
		}

		// Получаем имя клиента
		var clientName string
		clientQuery := `SELECT name FROM clients WHERE clients_id = $1`
		err = database.Pool.QueryRow(context.Background(), clientQuery, appointment.ClientID).Scan(&clientName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения имени клиента"})
			return
		}

		appointment.ServiceName = serviceName
		appointment.ClientName = clientName

		// Формируем ключ
		key := fmt.Sprintf("%v в %v %v %v",
			appointment.StartTime.Format("02.01.2006"),
			appointment.StartTime.Format("15:04"),
			appointment.ServiceName,
			appointment.ClientName,
		)

		// Добавляем запись в итоговый словарь
		result[key] = appointment
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, result)
}*/

// GetVisits получает записи с фильтрацией по статусу оплаты и/или времени (возвращает список с мапами)
/*func GetVisits(c *gin.Context) {
	// Получаем параметры фильтрации
	startParam := c.DefaultQuery("start", "")  // Параметр start (опционально)
	endParam := c.DefaultQuery("end", "")      // Параметр end (опционально)
	onlyUnpaid := c.DefaultQuery("unpaid", "") // Флаг, показывать только неоплаченные

	var startTime, endTime time.Time
	var err error

	// Обработка параметра start
	if startParam != "" {
		startTime, err = time.Parse("2006-01-02", startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра start"})
			return
		}
	} else {
		startTime = time.Time{} // Нулевое значение, если параметр отсутствует
	}

	// Обработка параметра end
	if endParam != "" {
		endTime, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра end"})
			return
		}
	} else {
		endTime = time.Now().Add(24 * time.Hour * 365) // Через год, если параметр отсутствует
	}

	// Начало запроса SQL
	query := `
		SELECT id, service_id, client_id, start_time, payment_status
		FROM appointments
		WHERE ($1::timestamptz IS NULL OR start_time >= $1)
		  AND ($2::timestamptz IS NULL OR start_time <= $2)
	`

	// Если требуется показать только неоплаченные записи
	if onlyUnpaid == "true" {
		query += " AND payment_status = 'unpaid'"
	}

	// Выполнение SQL-запроса
	rows, err := database.Pool.Query(context.Background(), query, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить записи"})
		return
	}
	defer rows.Close()

	// Создаем список словарей
	result := []map[string]models.Appointment{}

	for rows.Next() {
		var appointment models.Appointment
		if err := rows.Scan(&appointment.ID, &appointment.ServiceID, &appointment.ClientID, &appointment.StartTime, &appointment.PaymentStatus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных"})
			return
		}

		// Получаем название услуги
		var serviceName string
		serviceQuery := `SELECT name FROM services WHERE service_id = $1`
		err = database.Pool.QueryRow(context.Background(), serviceQuery, appointment.ServiceID).Scan(&serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения названия услуги"})
			return
		}

		// Получаем имя клиента
		var clientName string
		clientQuery := `SELECT name FROM clients WHERE clients_id = $1`
		err = database.Pool.QueryRow(context.Background(), clientQuery, appointment.ClientID).Scan(&clientName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения имени клиента"})
			return
		}

		appointment.ServiceName = serviceName
		appointment.ClientName = clientName

		// Формируем ключ
		key := fmt.Sprintf("%v в %v %v %v",
			appointment.StartTime.Format("02.01.2006"),
			appointment.StartTime.Format("15:04"),
			appointment.ServiceName,
			appointment.ClientName,
		)

		// Добавляем запись в итоговый список
		result = append(result, map[string]models.Appointment{key: appointment})
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, result)
}*/

func GetVisits(c *gin.Context) {
	startParam := c.DefaultQuery("start", "")
	onlyUnpaid := c.DefaultQuery("unpaid", "")

	var startTime *time.Time
	var err error

	// Обработка параметра start
	if startParam != "" {
		parsedStart, err := time.Parse("2006-01-02", startParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат параметра start"})
			return
		}
		startTime = &parsedStart
	}

	// Формируем SQL-запрос
	query := `
        SELECT id, service_id, client_id, start_time, payment_status
        FROM appointments
        WHERE 1 = 1
    `
	params := []interface{}{}

	if onlyUnpaid == "true" {
		query += " AND payment_status = 'unpaid'"
	}
	fmt.Println(query)
	if startTime != nil {
		query += " AND start_time >= $1"
		params = append(params, *startTime)
	}
	fmt.Println(query)

	// Выполняем запрос
	rows, err := database.Pool.Query(context.Background(), query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить записи"})
		return
	}
	defer rows.Close()

	result := make(map[string]models.Appointment)

	for rows.Next() {
		var appointment models.Appointment
		if err := rows.Scan(&appointment.ID, &appointment.ServiceID, &appointment.ClientID, &appointment.StartTime, &appointment.PaymentStatus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных"})
			return
		}

		var serviceName string
		serviceQuery := `SELECT name FROM services WHERE service_id = $1`
		err = database.Pool.QueryRow(context.Background(), serviceQuery, appointment.ServiceID).Scan(&serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения названия услуги"})
			return
		}

		var clientName string
		clientQuery := `SELECT name FROM clients WHERE clients_id = $1`
		err = database.Pool.QueryRow(context.Background(), clientQuery, appointment.ClientID).Scan(&clientName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения имени клиента"})
			return
		}

		appointment.ServiceName = serviceName
		appointment.ClientName = clientName

		key := fmt.Sprintf("%v в %v %v %v",
			appointment.StartTime.Format("02.01.2006"),
			appointment.StartTime.Format("15:04"),
			appointment.ServiceName,
			appointment.ClientName,
		)
		result[key] = appointment
	}

	c.JSON(http.StatusOK, result)
}

// Перенести запись
func MoveVisit(c *gin.Context) {
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
		SET 
			start_time = COALESCE($1, start_time)
		WHERE id = $2
		RETURNING id, start_time;
	`
	var updatedAppointment models.Appointment
	row := database.Pool.QueryRow(
		context.Background(),
		query,
		appointment.StartTime,
		id,
	)

	// Чтение обновленной записи
	if err := row.Scan(&updatedAppointment.ID, &updatedAppointment.StartTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленную запись
	c.JSON(http.StatusOK, updatedAppointment)
}

// Удалить запись по ID
func DeleteVisit(c *gin.Context) {
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

// Обновить статус записи
func UpdateVisitStatus(c *gin.Context) {
	// Читаем ID записи из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Читаем данные из тела запроса
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Выполняем обновление записи
	query := `
		UPDATE appointments
		SET
			appointment_status = COALESCE($1, appointment_status)
		WHERE id = $2
		RETURNING id, client_id, appointment_status;
	`
	var updatedAppointment models.Appointment
	row := database.Pool.QueryRow(
		context.Background(),
		query,
		appointment.AppointmentStatus,
		id,
	)

	// Чтение обновленной записи
	if err := row.Scan(&updatedAppointment.ID, &updatedAppointment.ClientID, &updatedAppointment.AppointmentStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// Возвращаем обновленную запись
	c.JSON(http.StatusOK, updatedAppointment)
}

// ОПЛАТА -------------------------------------------------------------------------------------------------------------

// Основная функция изменения статуса оплаты
func UpdatePaymentStatusMain(c *gin.Context) {
	// Получаем ID записи из параметров
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Читаем данные из тела запроса
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	// Открываем транзакцию
	tx, err := database.Pool.BeginTx(context.Background(), pgx.TxOptions{}) // Передаем правильные параметры
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала транзакции", "details": err.Error()})
		return
	}
	defer func() {
		if p := recover(); p != nil || err != nil {
			_ = tx.Rollback(context.Background()) // Передаем контекст в Rollback
		} else {
			_ = tx.Commit(context.Background()) // Передаем контекст в Commit
		}
	}()

	// Обновляем статус оплаты
	err = UpdatePaymentStatus(tx, id, *appointment.PaymentStatus) // Передаем транзакцию в функцию
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления статуса", "details": err.Error()})
		return
	}

	// Если статус оплаты изменяется на "paid", добавляем финансовую операцию
	if *appointment.PaymentStatus == "paid" {
		err = UpdatePaymentAmount(tx, id, appointment.ClientID, 0) // Передаем транзакцию
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления финансовой операции", "details": err.Error()})
			return
		}
	}

	// Если статус оплаты изменяется на "partially_paid", добавляем финансовую операцию
	if *appointment.PaymentStatus == "partially_paid" {
		err = UpdatePaymentAmount(tx, id, appointment.ClientID, appointment.Amount) // Передаем транзакцию
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления финансовой операции", "details": err.Error()})
			return
		}
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "Статус и данные обновлены успешно"})
}

// Обновить сумму платежа и создать финансовую операцию
func UpdatePaymentAmount(tx pgx.Tx, appointmentID int, clientID models.IntString, amount int) error {
	// Преобразуем clientID в int
	clientIDInt := IntStringToInt(clientID)

	if amount == 0 {
		// Сначала извлекаем стоимость записи (cost) из таблицы appointments
		query := `
		SELECT cost
		FROM appointments
		WHERE id = $1
	`
		err := tx.QueryRow(context.Background(), query, appointmentID).Scan(&amount)
		if err != nil {
			return fmt.Errorf("не удалось получить стоимость записи: %w", err)
		}
	}

	// Теперь создаём финансовую операцию с этой суммой
	paymentQuery := `
		INSERT INTO financial_operations (
			document_number, client_id, purpose, cashbox, amount, cashbox_balance, service_or_product, appointment_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	documentNumber := fmt.Sprintf("PAY-%d-%d", appointmentID, time.Now().Unix())
	purpose := "Оплата за услугу"
	cashbox := "Основная касса"
	cashboxBalance := amount      // Пример расчёта остатка кассы
	serviceOrProduct := "service" // Для оплаты услуги

	// Выполнение запроса с добавлением appointment_id
	_, err := tx.Exec(
		context.Background(),
		paymentQuery,
		documentNumber,
		clientIDInt, // Теперь передаем clientID как обычное int
		purpose,
		cashbox,
		amount,
		cashboxBalance,
		serviceOrProduct,
		appointmentID, // Добавляем связку с appointment_id
	)
	if err != nil {
		return fmt.Errorf("не удалось добавить финансовую операцию: %w", err)
	}
	return nil
}

// Обновить статус оплаты
/*func UpdatePaymentStatus(tx pgx.Tx, appointmentID int, paymentStatus string) error {
	// Запрос для обновления статуса оплаты
	query := `
		UPDATE appointments
		SET payment_status = $1
		WHERE id = $2
	`

	// Выполнение запроса
	_, err := tx.Exec(
		context.Background(),
		query,
		paymentStatus, // Статус оплаты, например 'paid'
		appointmentID,
	)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус оплаты: %w", err)
	}
	return nil
}*/

// UpdatePaymentStatus обновляет статус оплаты и удаляет финансовую операцию, если статус оплаты "не оплачено"
func UpdatePaymentStatus(tx pgx.Tx, appointmentID int, newStatus string) error {
	// Обновляем статус оплаты в таблице appointments
	_, err := tx.Exec(context.Background(), `
		UPDATE appointments
		SET payment_status = $1
		WHERE id = $2`, newStatus, appointmentID)
	if err != nil {
		return err
	}

	// Если статус "не оплачено", ищем и удаляем связанную финансовую операцию
	if newStatus == "unpaid" {
		// Удаляем финансовую операцию, если она существует
		_, err = tx.Exec(context.Background(), `
			DELETE FROM financial_operations
			WHERE appointment_id = $1`, appointmentID)
		if err != nil {
			return err
		}
	}

	// Нет необходимости явно делать tx.Commit, так как это сделает вызывающая функция
	// Если ошибка была, транзакция откатится, если нет — она будет зафиксирована в вызывающей функции
	return nil
}

// КЛИЕНТЫ ------------------------------------------------------------------------------------------------------------

// Создание клиента
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
	fmt.Println(req)
	// Форматируем номер телефона
	normalizedPhone := normalizePhone(req.Phone)
	if len(normalizedPhone) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат номера телефона"})
		return
	}

	// SQL-запрос для добавления клиента
	query := `
		INSERT INTO clients (phone, name)
		VALUES ($1, $2)
		RETURNING clients.clients_id
	`

	// Выполняем запрос к базе данных
	var clientID int
	err := database.Pool.QueryRow(context.Background(), query, normalizedPhone, req.Name).Scan(&clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать клиента", "details": err.Error()})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"message":   "Клиент успешно создан",
		"client_id": clientID,
		"phone":     normalizedPhone,
	})
}

// Поиск клиента по номеру телефона (возвращает id клиента)
func FindClientByPhoneHandler(c *gin.Context) {
	// Получаем номер телефона из запроса
	phone := c.Query("phone")
	fmt.Println(phone)
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Номер телефона обязателен",
			"details": "Параметр phone не передан или пустой",
		})
		return
	}

	// Нормализуем номер телефона
	normalizedPhone := normalizePhone(phone)

	var clientID int
	query := `SELECT clients.clients_id FROM clients WHERE phone = $1`

	// Выполняем запрос к базе данных с нормализованным телефоном
	err := database.Pool.QueryRow(context.Background(), query, normalizedPhone).Scan(&clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Клиент не найден",
				"details": fmt.Sprintf("Телефон: %s (нормализованный: %s)", phone, normalizedPhone),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Ошибка при поиске клиента",
			"details": err.Error(),
		})
		return
	}

	// Если клиент найден, возвращаем его ID
	c.JSON(http.StatusOK, gin.H{
		"client_id": clientID,
		"phone":     normalizedPhone,
	})
}

// GetClientInfoByID - функция для получения информации о клиенте по его ID
func GetClientInfoByID(clientID int) (map[string]interface{}, error) {
	// Запрос для получения информации о клиенте
	query := `
		SELECT 
			name, phone, email, categories, birth_date, paid, spent, gender, discount, 
			last_visit, first_visit, visit_count, comment 
		FROM clients 
		WHERE clients_id = $1
	`

	// Структура для хранения результатов запроса
	var (
		name       *string
		phone      *string
		email      *string
		categories *string
		birthDate  *time.Time
		paid       *float64
		spent      *float64
		gender     *string
		discount   *float64
		lastVisit  *time.Time
		firstVisit *time.Time
		visitCount *int
		comment    *string
	)

	// Выполнение запроса
	err := database.Pool.QueryRow(context.Background(), query, clientID).Scan(
		&name, &phone, &email, &categories, &birthDate, &paid, &spent, &gender, &discount,
		&lastVisit, &firstVisit, &visitCount, &comment,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить данные клиента: %w", err)
	}

	// Формирование мапы для ответа
	clientInfo := map[string]interface{}{
		"name":        name,
		"phone":       phone,
		"email":       email,
		"categories":  categories,
		"birth_date":  birthDate,
		"paid":        paid,
		"spent":       spent,
		"gender":      gender,
		"discount":    discount,
		"last_visit":  lastVisit,
		"first_visit": firstVisit,
		"visit_count": visitCount,
		"comment":     comment,
	}

	return clientInfo, nil
}

// GetClientInfoHandler - хендлер для получения информации о клиенте
func GetClientInfoHandler(c *gin.Context) {
	// Получаем clientID из параметров запроса, который передается через путь
	clientIDQuery := c.Query("client_id") // Параметр client_id передается в строке запроса, если он существует.
	if clientIDQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID клиента обязателен",
			"details": "ID клиента не передан или пустой",
		})
	}

	// Преобразуем clientID в int
	clientID, err := strconv.Atoi(clientIDQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат client_id"})
		return
	}
	// Получаем информацию о клиенте
	clientInfo, err := GetClientInfoByID(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Клиент не найден", "details": err.Error()})
		return
	}

	// Успешный ответ с информацией о клиенте
	c.JSON(http.StatusOK, clientInfo)
}

// АБОНЕМЕНТЫ ---------------------------------------------------------------------------------------------------------

// AddSubscriptionVisit добавляет посещение в таблицу subscription_visits
func AddSubscriptionVisit(tx pgx.Tx, subscriptionID int, visitDate time.Time) error {
	query := `
		INSERT INTO subscription_visits (subscription_id, visit_date)
		VALUES ($1, $2)
	`
	_, err := tx.Exec(context.Background(), query, subscriptionID, visitDate)
	if err != nil {
		return fmt.Errorf("ошибка добавления посещения: %w", err)
	}
	return nil
}

// DecreaseSubscriptionBalance уменьшает current_balance в таблице subscription
func DecreaseSubscriptionBalance(tx pgx.Tx, subscriptionID int) error {
	query := `
		UPDATE subscriptions
		SET current_balance = current_balance - 1
		WHERE subscriptions_id = $1 AND current_balance > 0
		RETURNING current_balance
	`

	var currentBalance int
	err := tx.QueryRow(context.Background(), query, subscriptionID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("ошибка уменьшения баланса: %w", err)
	}

	if currentBalance < 0 {
		return fmt.Errorf("баланс абонемента отрицательный")
	}

	return nil
}

// AddVisitTransaction выполняет транзакцию для добавления посещения по абонементу
// передаем аргументы: subscription_id и visit_date
func AddVisitTransaction(c *gin.Context) {
	// Структура для входящих данных
	type Request struct {
		SubscriptionID int       `json:"subscription_id" binding:"required"`
		VisitDate      time.Time `json:"visit_date" binding:"required"`
	}

	var req Request

	// Парсим входящие данные
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	// Открываем транзакцию
	tx, err := database.Pool.Begin(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала транзакции", "details": err.Error()})
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
		} else {
			_ = tx.Commit(context.Background())
		}
	}()

	// Добавляем посещение
	err = AddSubscriptionVisit(tx, req.SubscriptionID, req.VisitDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления посещения", "details": err.Error()})
		return
	}

	// Уменьшаем баланс абонемента
	err = DecreaseSubscriptionBalance(tx, req.SubscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка уменьшения баланса", "details": err.Error()})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "Посещение добавлено успешно"})
}

// GetActiveSubscription возвращает subscription_id и текущий баланс для клиента с положительным балансом
func GetActiveSubscriptionID(clientID, serviceID int) (int, int, error) {
	query := `
		SELECT s.subscriptions_id, s.current_balance
		FROM subscriptions s
		JOIN subscription_types st ON s.subscription_types_id = st.subscription_types_id
		WHERE s.client_id = $1
		  AND $2 = ANY(st.service_ids)
		  AND s.current_balance > 0
		LIMIT 1
	`

	var subscriptionID, currentBalance int
	err := database.Pool.QueryRow(context.Background(), query, clientID, serviceID).Scan(&subscriptionID, &currentBalance)
	if err != nil {
		return 0, 0, fmt.Errorf("не удалось найти активный абонемент: %w", err)
	}

	return subscriptionID, currentBalance, nil
}

// Handler для вызова GetActiveSubscription через HTTP-запрос
// передаем аргументы: client_id и service_id
func GetActiveSubscriptionHandler(c *gin.Context) {
	// Структура для входящих данных
	type Request struct {
		ClientID  int `json:"client_id" binding:"required"`
		ServiceID int `json:"service_id" binding:"required"`
	}

	var req Request

	// Парсим входящие данные
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	// Получаем ID активного абонемента и его текущий баланс
	subscriptionID, currentBalance, err := GetActiveSubscriptionID(req.ClientID, req.ServiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Активный абонемент не найден", "details": err.Error()})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"subscription_id": subscriptionID,
		"current_balance": currentBalance,
	})
}

// GetSubscriptionsByClientID возвращает список абонементов клиента с их балансами
func GetSubscriptionsByClientID(clientID int) ([]map[string]interface{}, error) {
	query := `
		SELECT subscriptions_id, current_balance
		FROM subscriptions
		WHERE client_id = $1
	`

	rows, err := database.Pool.Query(context.Background(), query, clientID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var subscriptions []map[string]interface{}

	for rows.Next() {
		var subscriptionID int
		var currentBalance int

		if err := rows.Scan(&subscriptionID, &currentBalance); err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %w", err)
		}

		subscriptions = append(subscriptions, map[string]interface{}{
			"subscription_id": subscriptionID,
			"current_balance": currentBalance,
		})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("ошибка чтения строк из результата запроса: %w", rows.Err())
	}

	return subscriptions, nil
}

// GetSubscriptionsHandler обрабатывает запрос на получение абонементов клиента через GET
func GetSubscriptionsHandler(c *gin.Context) {
	// Получаем client_id из параметров запроса
	clientIDParam := c.Query("client_id")
	if clientIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр client_id обязателен"})
		return
	}

	// Преобразуем client_id в int
	clientID, err := strconv.Atoi(clientIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат client_id", "details": err.Error()})
		return
	}

	// Получаем список абонементов клиента
	subscriptions, err := GetSubscriptionsByClientID(clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить абонементы клиента", "details": err.Error()})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
	})
}

// Получить список типов абонементов
func GetSubscriptionTypes(c *gin.Context) {
	query := `
		SELECT subscription_types_id, name, cost, sessions_count, service_ids
		FROM subscription_types
	`

	rows, err := database.Pool.Query(context.Background(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить типы абонементов", "details": err.Error()})
		return
	}
	defer rows.Close()

	// Используем map для формирования результата
	subscriptionTypes := make(map[string]models.SubscriptionType)

	for rows.Next() {
		var subType models.SubscriptionType
		if err := rows.Scan(&subType.ID, &subType.Name, &subType.Cost, &subType.SessionsCount, &subType.ServiceIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных", "details": err.Error()})
			return
		}
		// Добавляем запись в map
		key := fmt.Sprintf("%v", subType.Name)
		subscriptionTypes[key] = subType
	}
	fmt.Println(subscriptionTypes)
	c.JSON(http.StatusOK, subscriptionTypes)
}

// Продажа абонемента
// аргументы: client_id, subscription_types_id, cost
func SellSubscription(c *gin.Context) {
	var request struct {
		ClientID           int     `json:"client_id"`
		SubscriptionTypeID int     `json:"subscription_types_id"`
		Cost               float64 `json:"cost"`
		CurrentBalance     float64 `json:"sessions_count"`
	}

	// Читаем данные из запроса
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные", "details": err.Error()})
		return
	}

	// Проверяем, существует ли абонемент с указанным ID
	var subscriptionTypeName string
	checkQuery := `SELECT name FROM subscription_types WHERE subscription_types_id = $1`
	err := database.Pool.QueryRow(context.Background(), checkQuery, request.SubscriptionTypeID).Scan(&subscriptionTypeName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Абонемент с указанным ID не найден"})
		return
	}

	// Создаем финансовую операцию
	financeQuery := `
		INSERT INTO financial_operations (
			client_id, service_or_product, amount, purpose, document_number, cashbox
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	documentNumber := fmt.Sprintf("SUB-%d-%d", request.ClientID, time.Now().Unix())
	purpose := fmt.Sprintf("Покупка абонемента: %s", subscriptionTypeName)
	cashbox := "Основная касса"

	_, err = database.Pool.Exec(
		context.Background(),
		financeQuery,
		request.ClientID,
		"subscription",
		request.Cost,
		purpose,
		documentNumber,
		cashbox,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания финансовой операции", "details": err.Error()})
		return
	}

	// Добавляем запись в таблицу subscriptions
	subscriptionQuery := `
		INSERT INTO subscriptions (
			subscription_types_id, client_id, cost, current_balance
		)
		VALUES ($1, $2, $3, $4)
	`
	_, err = database.Pool.Exec(
		context.Background(),
		subscriptionQuery,
		request.SubscriptionTypeID,
		request.ClientID,
		request.Cost,
		request.CurrentBalance,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления абонемента", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Абонемент успешно продан и добавлен", "subscription": subscriptionTypeName})
}

// СТАТИСТИКА ---------------------------------------------------------------------------------------------------------

// Получение статистики за период (количество посещений, сумма по услугам, товарам)
func GetStatisticsHandler(c *gin.Context) {
	type Request struct {
		StartDate string `json:"start_date" binding:"required"`
		EndDate   string `json:"end_date" binding:"required"`
	}

	var req Request

	// Парсим параметры
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные", "details": err.Error()})
		return
	}

	// Преобразуем строки в дату
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты для start_date"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты для end_date"})
		return
	}

	// Получаем статистику
	stats, err := GetStatistics(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения статистики", "details": err.Error()})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, stats)
}
func GetStatistics(startDate, endDate time.Time) (models.Statistics, error) {
	// Преобразуем дату в строку формата YYYY-MM-DD
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	query := `
		WITH all_visits AS (
			SELECT operation_date::date AS visit_date
			FROM financial_operations
			WHERE operation_date::date BETWEEN $1::date AND $2::date
		
			UNION ALL
		
			SELECT visit_date
			FROM subscription_visits
			WHERE visit_date BETWEEN $1::date AND $2::date
		)
		SELECT
			(SELECT COUNT(*) FROM all_visits) AS total_visits, -- Теперь считаем все посещения из объединенной выборки
			COALESCE(SUM(fo.amount), 0) AS total_earnings,
			COALESCE(SUM(CASE WHEN fo.service_or_product = 'service' THEN fo.amount ELSE 0 END), 0) AS total_services,
			COALESCE(SUM(CASE WHEN fo.service_or_product = 'subscription' THEN fo.amount ELSE 0 END), 0) AS total_subscriptions
		FROM
			financial_operations fo
		WHERE
			fo.operation_date::date BETWEEN $1::date AND $2::date;
	`
	// Используем указатели для обработки возможных NULL
	var totalVisits int
	var totalEarnings, totalServices, totalSubscriptions *float64

	// Выполняем запрос
	err := database.Pool.QueryRow(context.Background(), query, startDateStr, endDateStr).Scan(
		&totalVisits,
		&totalEarnings,
		&totalServices,
		&totalSubscriptions,
	)
	if err != nil {
		return models.Statistics{}, fmt.Errorf("не удалось получить статистику: %w", err)
	}

	// Преобразуем NULL в нулевые значения
	stats := models.Statistics{
		TotalVisits:        totalVisits,
		TotalEarnings:      coalesceFloat64(totalEarnings),
		TotalServices:      coalesceFloat64(totalServices),
		TotalSubscriptions: coalesceFloat64(totalSubscriptions),
	}
	fmt.Println(stats)
	return stats, nil
}

// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ --------------------------------------------------------------------------------------------

// Функция для преобразования IntString в int
func IntStringToInt(value models.IntString) int {
	return int(value)
}

// Вспомогательная функция для обработки NULL
func coalesceFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

// normalizePhone форматирует телефон в стандартный вид, соответствующий базе данных
func normalizePhone(phone string) string {
	// Удаляем все нецифровые символы
	re := regexp.MustCompile(`\D`)
	normalized := re.ReplaceAllString(phone, "")
	// Если номер начинается с "8", заменяем на "7"
	if strings.HasPrefix(normalized, "8") {
		normalized = "7" + normalized[1:]
	}
	return normalized
}
