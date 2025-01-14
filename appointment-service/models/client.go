package models

import (
	"time"
)

// ClientInfo - структура для хранения полной информации о клиенте
type ClientInfo struct {
	Name       string     `json:"name"`
	Phone      string     `json:"phone"`
	Email      string     `json:"email"`
	Categories string     `json:"categories"`  // Пример: категории клиента (VIP, обычный и т.д.)
	BirthDate  *time.Time `json:"birth_date"`  // Используем указатель, если поле может быть NULL
	Paid       float64    `json:"paid"`        // Сумма оплаченных средств
	Spent      float64    `json:"spent"`       // Сумма потраченных средств
	Gender     *string    `json:"gender"`      // Пол клиента, указатель если может быть NULL
	Discount   float64    `json:"discount"`    // Размер скидки
	LastVisit  *time.Time `json:"last_visit"`  // Дата последнего посещения
	FirstVisit *time.Time `json:"first_visit"` // Дата первого посещения
	VisitCount int        `json:"visit_count"` // Общее количество посещений
	Comment    *string    `json:"comment"`     // Комментарий о клиенте
}
