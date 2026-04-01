package entity

import "time"

type VisitStatus string

const (
	StatusScheduled  VisitStatus = "scheduled"
	StatusConfirmed  VisitStatus = "confirmed"
	StatusInProgress VisitStatus = "in_progress"
	StatusCompleted  VisitStatus = "completed"
	StatusCancelled  VisitStatus = "cancelled"
	StatusNoShow     VisitStatus = "no_show"
)

type Visit struct {
	ID        string
	ClientID  string
	MasterID  string
	ServiceID string

	StartAt time.Time
	EndAt   time.Time

	PriceCents int64
	Currency   string

	Status VisitStatus

	Comment         string
	CancelledReason *string

	CreatedAt time.Time
	UpdatedAt time.Time
}
