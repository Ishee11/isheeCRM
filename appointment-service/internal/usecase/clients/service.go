package clients

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

type Repository interface {
	Create(ctx context.Context, phone, name string) (int, error)
	FindByPhone(ctx context.Context, phone string) (int, error)
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	GetInfo(ctx context.Context, clientID int) (Info, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type Info struct {
	Name       *string
	Phone      *string
	Email      *string
	Categories *string
	BirthDate  *time.Time
	Paid       *float64
	Spent      *float64
	Gender     *string
	Discount   *float64
	LastVisit  *time.Time
	FirstVisit *time.Time
	VisitCount *int
	Comment    *string
}

type SearchResult struct {
	ID         int
	Name       *string
	Phone      *string
	LastVisit  *time.Time
	VisitCount *int
}

func (s *Service) Create(ctx context.Context, phone, name string) (int, string, error) {
	if strings.TrimSpace(name) == "" {
		return 0, "", fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	normalizedPhone := normalizePhone(phone)
	if !isValidPhone(normalizedPhone) {
		return 0, "", fmt.Errorf("%w: invalid phone format", ErrInvalidInput)
	}

	clientID, err := s.repo.Create(ctx, normalizedPhone, name)
	if err != nil {
		return 0, "", err
	}

	return clientID, normalizedPhone, nil
}

func (s *Service) FindByPhone(ctx context.Context, phone string) (int, string, error) {
	normalizedPhone := normalizePhone(phone)
	if !isValidPhone(normalizedPhone) {
		return 0, "", fmt.Errorf("%w: invalid phone format", ErrInvalidInput)
	}

	clientID, err := s.repo.FindByPhone(ctx, normalizedPhone)
	if err != nil {
		return 0, "", err
	}

	return clientID, normalizedPhone, nil
}

func (s *Service) GetInfo(ctx context.Context, clientID int) (Info, error) {
	if clientID <= 0 {
		return Info{}, fmt.Errorf("%w: client_id must be positive", ErrInvalidInput)
	}

	return s.repo.GetInfo(ctx, clientID)
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return nil, fmt.Errorf("%w: query must contain at least 2 characters", ErrInvalidInput)
	}
	if limit <= 0 || limit > 20 {
		limit = 8
	}

	return s.repo.Search(ctx, query, limit)
}

func normalizePhone(phone string) string {
	re := regexp.MustCompile(`\D`)
	normalized := re.ReplaceAllString(phone, "")
	if strings.HasPrefix(normalized, "8") {
		normalized = "7" + normalized[1:]
	}
	return normalized
}

func isValidPhone(phone string) bool {
	matched, _ := regexp.MatchString(`^7[0-9]{10}$`, phone)
	return matched
}
