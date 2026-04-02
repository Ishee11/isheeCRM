package clients

import (
	"context"
	"errors"
	"testing"
)

type fakeClientsRepo struct {
	createCalls     int
	createdPhone    string
	createdName     string
	findCalls       int
	searchedPhone   string
	infoCalls       int
	queriedClientID int
}

func (f *fakeClientsRepo) Create(ctx context.Context, phone, name string) (int, error) {
	f.createCalls++
	f.createdPhone = phone
	f.createdName = name
	return 99, nil
}

func (f *fakeClientsRepo) FindByPhone(ctx context.Context, phone string) (int, error) {
	f.findCalls++
	f.searchedPhone = phone
	return 88, nil
}

func (f *fakeClientsRepo) GetInfo(ctx context.Context, clientID int) (Info, error) {
	f.infoCalls++
	f.queriedClientID = clientID
	return Info{Name: ptrString("John")}, nil
}

func ptrString(s string) *string {
	return &s
}

func TestService_CreateValidation(t *testing.T) {
	repo := &fakeClientsRepo{}
	svc := NewService(repo)
	if _, _, err := svc.Create(context.Background(), "79000000000", ""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Create() error %v, want ErrInvalidInput", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("repo should not be called when name empty")
	}
}

func TestService_CreateNormalizesPhone(t *testing.T) {
	repo := &fakeClientsRepo{}
	svc := NewService(repo)
	if _, normalized, err := svc.Create(context.Background(), "8 (912) 345-67-89", "Anna"); err != nil {
		t.Fatalf("Create() unexpected error %v", err)
	} else if normalized != "79123456789" {
		t.Fatalf("normalized phone = %s", normalized)
	}
	if repo.createdPhone != "79123456789" {
		t.Fatalf("repo received %s", repo.createdPhone)
	}
}

func TestService_FindByPhoneValidation(t *testing.T) {
	repo := &fakeClientsRepo{}
	svc := NewService(repo)
	if _, _, err := svc.FindByPhone(context.Background(), "123"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("FindByPhone() error %v, want ErrInvalidInput", err)
	}
	if repo.findCalls != 0 {
		t.Fatalf("repo should not be called on invalid phone")
	}
}

func TestService_FindByPhoneNormalizes(t *testing.T) {
	repo := &fakeClientsRepo{}
	svc := NewService(repo)
	if _, normalized, err := svc.FindByPhone(context.Background(), "8 912 000 00 00"); err != nil {
		t.Fatalf("unexpected error %v", err)
	} else if normalized != "79120000000" {
		t.Fatalf("normalized phone = %s", normalized)
	}
	if repo.searchedPhone != "79120000000" {
		t.Fatalf("repo got %s", repo.searchedPhone)
	}
}

func TestService_GetInfoValidation(t *testing.T) {
	repo := &fakeClientsRepo{}
	svc := NewService(repo)
	if _, err := svc.GetInfo(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("GetInfo() error %v, want ErrInvalidInput", err)
	}
	if repo.infoCalls != 0 {
		t.Fatalf("repo should not be called for invalid id")
	}
}
