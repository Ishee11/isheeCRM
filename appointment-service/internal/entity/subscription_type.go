package entity

// SubscriptionType отражает тип абонемента для интерфейса
type SubscriptionType struct {
	ID            int     `json:"subscription_types_id"`
	Name          string  `json:"name"`
	Cost          float64 `json:"cost"`
	SessionsCount int     `json:"sessions_count"`
	ServiceIDs    []int   `json:"service_ids"`
}
