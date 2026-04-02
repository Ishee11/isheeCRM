package entity

// Statistics содержит агрегаты для отчётов
type Statistics struct {
	TotalVisits        int     `json:"total_visits"`
	TotalEarnings      float64 `json:"total_earnings"`
	TotalServices      float64 `json:"total_services"`
	TotalSubscriptions float64 `json:"total_subscriptions"`
}
