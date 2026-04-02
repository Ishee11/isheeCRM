package entity

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// IntString помогает корректно парсить числа, приходящие из JSON, как строку или число
type IntString int

// UnmarshalJSON позволяет принимать как строку, так и число в JSON
func (i *IntString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		num, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", str)
		}
		*i = IntString(num)
		return nil
	}

	var num int
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("invalid data format: %s", string(data))
	}
	*i = IntString(num)
	return nil
}

// Appointment описывает запись на приём в терминах бизнес-логики
type Appointment struct {
	ID                uint       `json:"id"`
	Cost              uint       `json:"cost"`
	ServiceID         IntString  `json:"service_id"`
	ClientID          IntString  `json:"client_id"`
	StartTime         *time.Time `json:"start_time"`
	Price             IntString  `json:"price"`
	PaymentStatus     *string    `json:"payment_status"`
	AppointmentStatus *string    `json:"appointment_status"`
	Amount            int        `json:"amount"`
	ServiceName       string     `json:"service_name"`
	ClientName        string     `json:"client_name"`
}
