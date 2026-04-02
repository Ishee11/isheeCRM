package entity

// Service описывает карточку услуги
type Service struct {
	ID       int     `json:"service_id"`
	Name     string  `json:"name" binding:"required"`
	Duration int     `json:"duration" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
}
