package appointments

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest) (Appointment, error)
	List(ctx context.Context, filter ListFilter) ([]Appointment, error)
	UpdateStatus(ctx context.Context, appointmentID int, status string) (AppointmentStatusUpdate, error)
	Move(ctx context.Context, appointmentID int, startTime time.Time) (AppointmentMoveResult, error)
	SoftDelete(ctx context.Context, appointmentID int) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateRequest struct {
	ServiceID int
	ClientID  int
	StartTime time.Time
}

type ListFilter struct {
	OnlyUnpaid bool
	Start      *time.Time
}

type Appointment struct {
	ID                int
	ServiceID         int
	ClientID          int
	StartTime         time.Time
	PaymentStatus     *string
	AppointmentStatus *string
	ServiceName       string
	ClientName        string
}

type AppointmentStatusUpdate struct {
	ID                int
	ClientID          int
	AppointmentStatus *string
}

type AppointmentMoveResult struct {
	ID        int
	StartTime time.Time
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Appointment, error) {
	if req.ServiceID <= 0 || req.ClientID <= 0 || req.StartTime.IsZero() {
		return Appointment{}, fmt.Errorf("%w: service_id, client_id and start_time are required", ErrInvalidInput)
	}

	return s.repo.Create(ctx, req)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Appointment, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) UpdateStatus(ctx context.Context, appointmentID int, status string) (AppointmentStatusUpdate, error) {
	if appointmentID <= 0 || status == "" {
		return AppointmentStatusUpdate{}, fmt.Errorf("%w: appointment_id and appointment_status are required", ErrInvalidInput)
	}

	return s.repo.UpdateStatus(ctx, appointmentID, status)
}

func (s *Service) Move(ctx context.Context, appointmentID int, startTime time.Time) (AppointmentMoveResult, error) {
	if appointmentID <= 0 || startTime.IsZero() {
		return AppointmentMoveResult{}, fmt.Errorf("%w: appointment_id and start_time are required", ErrInvalidInput)
	}

	return s.repo.Move(ctx, appointmentID, startTime)
}

func (s *Service) Delete(ctx context.Context, appointmentID int) error {
	if appointmentID <= 0 {
		return fmt.Errorf("%w: appointment_id must be positive", ErrInvalidInput)
	}

	return s.repo.SoftDelete(ctx, appointmentID)
}
