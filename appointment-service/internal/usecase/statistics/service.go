package statistics

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
	GetByPeriod(ctx context.Context, startDate, endDate time.Time) (Summary, error)
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

type Summary struct {
	TotalVisits        int
	TotalEarnings      float64
	TotalServices      float64
	TotalSubscriptions float64
}

func (s *Service) GetByPeriod(ctx context.Context, startDate, endDate time.Time) (Summary, error) {
	if startDate.IsZero() || endDate.IsZero() {
		return Summary{}, fmt.Errorf("%w: start_date and end_date are required", ErrInvalidInput)
	}
	if endDate.Before(startDate) {
		return Summary{}, fmt.Errorf("%w: end_date must not be before start_date", ErrInvalidInput)
	}

	return s.repo.GetByPeriod(ctx, startDate, endDate)
}

func (s *Service) GetCurrentMonth(ctx context.Context) (Summary, error) {
	now := s.now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return s.repo.GetByPeriod(ctx, startDate, now)
}
