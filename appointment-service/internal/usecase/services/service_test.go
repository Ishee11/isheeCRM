package services

import (
	"context"
	"errors"
	"testing"
)

type fakeServicesRepo struct {
	createCalls int
	lastReq     CreateRequest
	deleteCalls int
	lastID      int
}

func (f *fakeServicesRepo) Create(ctx context.Context, req CreateRequest) (ServiceDTO, error) {
	f.createCalls++
	f.lastReq = req
	return ServiceDTO{ID: 42, Name: req.Name, Duration: req.Duration, Price: req.Price}, nil
}

func (f *fakeServicesRepo) List(ctx context.Context) ([]ServiceDTO, error) {
	return nil, nil
}

func (f *fakeServicesRepo) SoftDelete(ctx context.Context, serviceID int) error {
	f.deleteCalls++
	f.lastID = serviceID
	return nil
}

func TestService_CreateValidation(t *testing.T) {
	repo := &fakeServicesRepo{}
	svc := NewService(repo)
	tests := []struct {
		name    string
		req     CreateRequest
		wantErr error
	}{
		{"empty name", CreateRequest{Name: "", Duration: 30, Price: 100}, ErrInvalidInput},
		{"zero duration", CreateRequest{Name: "A", Duration: 0, Price: 100}, ErrInvalidInput},
		{"negative price", CreateRequest{Name: "B", Duration: 30, Price: -1}, ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.Create(context.Background(), tt.req); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if repo.createCalls != 0 {
		t.Fatalf("repo should not be called when request is invalid, got %d", repo.createCalls)
	}
}

func TestService_CreateSuccess(t *testing.T) {
	repo := &fakeServicesRepo{}
	svc := NewService(repo)
	req := CreateRequest{Name: "Cut", Duration: 45, Price: 120}
	got, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if got.ID != 42 {
		t.Fatalf("unexpected ID %d", got.ID)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repo called once, got %d", repo.createCalls)
	}
}

func TestService_DeleteValidation(t *testing.T) {
	repo := &fakeServicesRepo{}
	svc := NewService(repo)
	if err := svc.Delete(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Delete() error = %v, want ErrInvalidInput", err)
	}
	if repo.deleteCalls != 0 {
		t.Fatalf("repo should not be invoked on invalid delete, got %d", repo.deleteCalls)
	}
}
