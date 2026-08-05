package clients

import "time"

type History struct {
	ClientID      int            `json:"client_id"`
	Visits        []Visit        `json:"visits"`
	Payments      []Payment      `json:"payments"`
	Subscriptions []Subscription `json:"subscriptions"`
}

type Visit struct {
	AppointmentID     int       `json:"appointment_id"`
	ServiceID         int       `json:"service_id"`
	ServiceName       string    `json:"service_name"`
	StartTime         time.Time `json:"start_time"`
	AppointmentStatus string    `json:"appointment_status"`
	PaymentStatus     string    `json:"payment_status"`
	Cost              float64   `json:"cost"`
	SubscriptionID    *int      `json:"subscription_id"`
}

type Payment struct {
	ID               int        `json:"id"`
	OperationDate    *time.Time `json:"operation_date"`
	Amount           float64    `json:"amount"`
	Purpose          string     `json:"purpose"`
	ServiceOrProduct string     `json:"service_or_product"`
	AppointmentID    *int       `json:"appointment_id"`
	DocumentNumber   string     `json:"document_number"`
	Cashbox          string     `json:"cashbox"`
}

type Subscription struct {
	ID                int                 `json:"subscription_id"`
	TypeID            int                 `json:"subscription_type_id"`
	TypeName          string              `json:"subscription_type_name"`
	PurchasedAt       *time.Time          `json:"purchased_at"`
	Cost              float64             `json:"cost"`
	SessionsCount     int                 `json:"sessions_count"`
	TypeSessionsCount int                 `json:"type_sessions_count"`
	UsedCount         int                 `json:"used_count"`
	CurrentBalance    int                 `json:"current_balance"`
	Status            string              `json:"status"`
	DeletedAt         *time.Time          `json:"deleted_at"`
	Visits            []SubscriptionVisit `json:"visits"`
}

type SubscriptionVisit struct {
	ID            int       `json:"id"`
	AppointmentID *int      `json:"appointment_id"`
	VisitDate     time.Time `json:"visit_date"`
	ServiceID     *int      `json:"service_id"`
	ServiceName   *string   `json:"service_name"`
}
