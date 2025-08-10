package models

import (
	"appointment-service/database"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"strconv"
	"time"
)

// IntString - тип для кастомной обработки чисел, переданных в виде строки
type IntString int

// Метод для кастомного парсинга чисел из строки
func (i *IntString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Если это строка, пытаемся преобразовать в int
		num, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", str)
		}
		*i = IntString(num)
		return nil
	}

	// Если это не строка, пробуем как число
	var num int
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("invalid data format: %s", string(data))
	}
	*i = IntString(num)
	return nil
}

// Appointment - структура для записи на прием
type Appointment struct {
	ID                int        `json:"id"`
	Cost              int        `json:"cost"`
	ServiceID         IntString  `json:"service_id"`         // Используем указатель для возможности проверки, передано ли значение
	ClientID          IntString  `json:"client_id"`          // Используем указатель
	StartTime         *time.Time `json:"start_time"`         // Используем указатель для времени
	Price             IntString  `json:"price"`              // Используем указатель для цены
	PaymentStatus     *string    `json:"payment_status"`     // Используем указатель для статуса оплаты
	AppointmentStatus *string    `json:"appointment_status"` // Используем указатель для статуса записи
	Amount            int        `json:"amount"`             // Сумма оплаты
	ServiceName       string     `json:"service_name"`
	ClientName        string     `json:"client_name"`
}

// Save сохраняет запись в базе и заполняет ID
func (a *Appointment) Save(ctx context.Context) error {
	query := `
		INSERT INTO appointments (service_id, client_id, start_time)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	return database.Pool.QueryRow(ctx, query, a.ServiceID, a.ClientID, a.StartTime).Scan(&a.ID)
}

// Helper для обработки ошибок PG (можно вынести в отдельный пакет, если надо)
func InterpretPgError(err error) (int, string) {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23503":
			return http.StatusBadRequest, "Указанная услуга или клиент не существуют"
		case "23514":
			return http.StatusBadRequest, "Некорректные данные (нарушение ограничения CHECK)"
		case "23505":
			return http.StatusConflict, "Такая запись уже существует"
		default:
			return http.StatusInternalServerError, "Ошибка базы данных"
		}
	}
	return http.StatusInternalServerError, "Не удалось создать запись"
}

func GetAppointments(ctx context.Context, onlyUnpaid bool, start *time.Time) ([]Appointment, error) {
	query := `
		SELECT id, service_id, client_id, start_time, payment_status
		FROM appointments
		WHERE 1=1
	`
	params := []interface{}{}
	paramIndex := 1

	if onlyUnpaid {
		query += " AND payment_status = 'unpaid'"
	}
	if start != nil {
		query += fmt.Sprintf(" AND start_time >= $%d", paramIndex)
		params = append(params, *start)
		paramIndex++
	}

	rows, err := database.Pool.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var a Appointment
		if err := rows.Scan(&a.ID, &a.ServiceID, &a.ClientID, &a.StartTime, &a.PaymentStatus); err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (a *Appointment) LoadServiceName(ctx context.Context) error {
	query := `SELECT name FROM services WHERE service_id = $1`
	return database.Pool.QueryRow(ctx, query, a.ServiceID).Scan(&a.ServiceName)
}

func (a *Appointment) LoadClientName(ctx context.Context) error {
	query := `SELECT name FROM clients WHERE clients_id = $1`
	return database.Pool.QueryRow(ctx, query, a.ClientID).Scan(&a.ClientName)
}
