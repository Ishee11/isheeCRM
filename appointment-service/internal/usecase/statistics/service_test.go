package statistics

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStatisticsRepo struct {
	getCalls int
	start    time.Time
	end      time.Time
}

func (f *fakeStatisticsRepo) GetByPeriod(ctx context.Context, startDate, endDate time.Time) (Summary, error) {
	f.getCalls++
	f.start = startDate
	f.end = endDate
	return Summary{TotalVisits: 5}, nil
}

func TestService_GetByPeriodValidation(t *testing.T) {
	repo := &fakeStatisticsRepo{}
	svc := NewService(repo)
	if _, err := svc.GetByPeriod(context.Background(), time.Time{}, time.Now()); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput when start empty, got %v", err)
	}
	if _, err := svc.GetByPeriod(context.Background(), time.Now(), time.Now().Add(-24*time.Hour)); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput when end before start, got %v", err)
	}
	if repo.getCalls != 0 {
		t.Fatalf("repo should not be called on invalid input")
	}
}

func TestService_GetCurrentMonth(t *testing.T) {
	repo := &fakeStatisticsRepo{}
	svc := NewService(repo)
	svc.now = func() time.Time {
		return time.Date(2026, time.April, 10, 3, 0, 0, 0, time.UTC)
	}
	if _, err := svc.GetCurrentMonth(context.Background()); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	wantStart := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	if !repo.start.Equal(wantStart) {
		t.Fatalf("start = %v, want %v", repo.start, wantStart)
	}
	if !repo.end.Equal(svc.now()) {
		t.Fatalf("end = %v, want now", repo.end)
	}
}
