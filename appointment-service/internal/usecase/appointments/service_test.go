package appointments

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeAppointmentsRepo struct {
	createCalls     int
	lastCreateReq   CreateRequest
	statusCalls     int
	moveCalls       int
	softDeleteCalls int
}

func (f *fakeAppointmentsRepo) Create(ctx context.Context, req CreateRequest) (Appointment, error) {
	f.createCalls++
	f.lastCreateReq = req
	return Appointment{ID: 100, ServiceID: req.ServiceID, ClientID: req.ClientID, StartTime: req.StartTime}, nil
}

func (f *fakeAppointmentsRepo) List(ctx context.Context, filter ListFilter) ([]Appointment, error) {
	return nil, nil
}

func (f *fakeAppointmentsRepo) UpdateStatus(ctx context.Context, appointmentID int, status string) (AppointmentStatusUpdate, error) {
	f.statusCalls++
	return AppointmentStatusUpdate{ID: appointmentID, ClientID: 1, AppointmentStatus: &status}, nil
}

func (f *fakeAppointmentsRepo) Move(ctx context.Context, appointmentID int, startTime time.Time) (AppointmentMoveResult, error) {
	f.moveCalls++
	return AppointmentMoveResult{ID: appointmentID, StartTime: startTime}, nil
}

func (f *fakeAppointmentsRepo) SoftDelete(ctx context.Context, appointmentID int) error {
	f.softDeleteCalls++
	return nil
}

func TestService_CreateValidation(t *testing.T) {
	svc := NewService(&fakeAppointmentsRepo{})
	if _, err := svc.Create(context.Background(), CreateRequest{ServiceID: 0, ClientID: 1, StartTime: time.Now()}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for zero service")
	}
	if _, err := svc.Create(context.Background(), CreateRequest{ServiceID: 1, ClientID: 0, StartTime: time.Now()}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for zero client")
	}
	if _, err := svc.Create(context.Background(), CreateRequest{ServiceID: 1, ClientID: 1}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for zero start time")
	}
}

func TestService_CreateSuccess(t *testing.T) {
	repo := &fakeAppointmentsRepo{}
	svc := NewService(repo)
	req := CreateRequest{ServiceID: 1, ClientID: 2, StartTime: time.Now()}
	if _, err := svc.Create(context.Background(), req); err != nil {
		t.Fatalf("unexpected %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("repo should be called once")
	}
}

func TestService_UpdateStatusValidation(t *testing.T) {
	svc := NewService(&fakeAppointmentsRepo{})
	if _, err := svc.UpdateStatus(context.Background(), 0, "confirmed"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for zero id")
	}
	if _, err := svc.UpdateStatus(context.Background(), 1, ""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty status")
	}
}

func TestService_MoveValidation(t *testing.T) {
	svc := NewService(&fakeAppointmentsRepo{})
	if _, err := svc.Move(context.Background(), 0, time.Now()); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for zero id")
	}
	if _, err := svc.Move(context.Background(), 1, time.Time{}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty time")
	}
}

func TestService_DeleteValidation(t *testing.T) {
	svc := NewService(&fakeAppointmentsRepo{})
	if err := svc.Delete(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for zero id")
	}
}
