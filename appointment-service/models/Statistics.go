package models

type Statistics struct {
	TotalVisits        int     `json:"total_visits"`        // Общее количество посещений
	TotalEarnings      float64 `json:"total_earnings"`      // Общая заработанная сумма
	TotalServices      float64 `json:"total_services"`      // Общая сумма по услугам
	TotalSubscriptions float64 `json:"total_subscriptions"` // Общая сумма по абонементам
}
