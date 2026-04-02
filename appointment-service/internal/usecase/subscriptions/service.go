package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest) (SubscriptionType, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateRequest struct {
	Name          string
	Cost          float64
	SessionsCount int
	ServiceIDs    []int
}

type SubscriptionType struct {
	ID            int
	Name          string
	Cost          float64
	SessionsCount int
	ServiceIDs    []int
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (SubscriptionType, error) {
	if strings.TrimSpace(req.Name) == "" {
		return SubscriptionType{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if req.Cost <= 0 {
		return SubscriptionType{}, fmt.Errorf("%w: cost must be positive", ErrInvalidInput)
	}
	if req.SessionsCount <= 0 {
		return SubscriptionType{}, fmt.Errorf("%w: sessions_count must be positive", ErrInvalidInput)
	}
	if len(req.ServiceIDs) == 0 {
		return SubscriptionType{}, fmt.Errorf("%w: service_ids must not be empty", ErrInvalidInput)
	}

	return s.repo.Create(ctx, req)
}
