package entity

// Statistics содержит агрегаты для отчётов
type Statistics struct {
	TotalVisits        int     `json:"total_visits"`
	TotalEarnings      float64 `json:"total_earnings"`
	TotalServices      float64 `json:"total_services"`
	TotalSubscriptions float64 `json:"total_subscriptions"`
}

// ClientPaymentStatistics содержит платежные агрегаты по клиенту.
type ClientPaymentStatistics struct {
	ClientID  int     `json:"client_id"`
	Name      string  `json:"name"`
	Count     int64   `json:"count"`
	AvgAmount float64 `json:"avg_amount"`
	Paid      float64 `json:"paid"`
}
