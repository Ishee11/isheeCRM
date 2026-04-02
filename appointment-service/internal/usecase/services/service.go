package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest) (ServiceDTO, error)
	List(ctx context.Context) ([]ServiceDTO, error)
	SoftDelete(ctx context.Context, serviceID int) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateRequest struct {
	Name     string
	Duration int
	Price    float64
}

type ServiceDTO struct {
	ID       int
	Name     string
	Duration int
	Price    float64
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (ServiceDTO, error) {
	if strings.TrimSpace(req.Name) == "" {
		return ServiceDTO{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if req.Duration <= 0 {
		return ServiceDTO{}, fmt.Errorf("%w: duration must be positive", ErrInvalidInput)
	}
	if req.Price <= 0 {
		return ServiceDTO{}, fmt.Errorf("%w: price must be positive", ErrInvalidInput)
	}

	return s.repo.Create(ctx, req)
}

func (s *Service) List(ctx context.Context) ([]ServiceDTO, error) {
	return s.repo.List(ctx)
}

func (s *Service) Delete(ctx context.Context, serviceID int) error {
	if serviceID <= 0 {
		return fmt.Errorf("%w: service_id must be positive", ErrInvalidInput)
	}

	return s.repo.SoftDelete(ctx, serviceID)
}
