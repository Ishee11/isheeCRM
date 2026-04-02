package entity

import "time"

// ClientInfo описывает базовую информацию о клиенте и рассчитанные поля
type ClientInfo struct {
	Name       string     `json:"name"`
	Phone      string     `json:"phone"`
	Email      string     `json:"email"`
	Categories string     `json:"categories"`
	BirthDate  *time.Time `json:"birth_date"`
	Paid       float64    `json:"paid"`
	Spent      float64    `json:"spent"`
	Gender     *string    `json:"gender"`
	Discount   float64    `json:"discount"`
	LastVisit  *time.Time `json:"last_visit"`
	FirstVisit *time.Time `json:"first_visit"`
	VisitCount int        `json:"visit_count"`
	Comment    *string    `json:"comment"`
}
