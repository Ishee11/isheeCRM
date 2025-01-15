package models

import (
	"encoding/json"
	"fmt"
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
