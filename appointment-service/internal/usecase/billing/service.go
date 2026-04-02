package billing

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

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type SubscriptionVisitWriter interface {
	AddSubscriptionVisit(ctx context.Context, subscriptionID int, visitDate time.Time, appointmentID int) error
}

type SubscriptionBalanceRepository interface {
	DecreaseSubscriptionBalance(ctx context.Context, subscriptionID int) error
	RestoreSubscriptionBalance(ctx context.Context, subscriptionID int) error
}

type ActiveSubscriptionFinder interface {
	FindActiveSubscription(ctx context.Context, clientID, serviceID int) (ActiveSubscription, error)
}

type ClientSubscriptionLister interface {
	ListClientSubscriptions(ctx context.Context, clientID int) ([]ClientSubscription, error)
}

type SubscriptionTypeCatalog interface {
	ListSubscriptionTypes(ctx context.Context) ([]SubscriptionType, error)
	GetSubscriptionTypeName(ctx context.Context, subscriptionTypeID int) (string, error)
}

type SubscriptionSeller interface {
	CreateSubscriptionSaleOperation(ctx context.Context, clientID int, amount float64, purpose, documentNumber, cashbox string) error
	CreateSubscription(ctx context.Context, subscriptionTypeID, clientID int, cost float64, currentBalance int) error
}

type AppointmentPaymentStatusUpdater interface {
	UpdateAppointmentPaymentStatus(ctx context.Context, appointmentID int, status string) error
}

type AppointmentPaymentOperator interface {
	GetAppointmentCost(ctx context.Context, appointmentID int) (int, error)
	CreateAppointmentPaymentOperation(ctx context.Context, operation AppointmentPaymentOperation) error
}

type SubscriptionPaymentRollbackRepository interface {
	DeleteFinancialOperationsByAppointment(ctx context.Context, appointmentID int) error
	GetSubscriptionIDByAppointment(ctx context.Context, appointmentID int) (int, error)
	DeleteSubscriptionVisitByAppointment(ctx context.Context, appointmentID int) error
}

type Service struct {
	txManager                  TxManager
	subscriptionVisitWriter    SubscriptionVisitWriter
	subscriptionBalanceRepo    SubscriptionBalanceRepository
	activeSubscriptionFinder   ActiveSubscriptionFinder
	clientSubscriptionLister   ClientSubscriptionLister
	subscriptionTypeCatalog    SubscriptionTypeCatalog
	subscriptionSeller         SubscriptionSeller
	paymentStatusUpdater       AppointmentPaymentStatusUpdater
	appointmentPaymentOperator AppointmentPaymentOperator
	paymentRollbackRepo        SubscriptionPaymentRollbackRepository
}

type Dependencies struct {
	TxManager                  TxManager
	SubscriptionVisitWriter    SubscriptionVisitWriter
	SubscriptionBalanceRepo    SubscriptionBalanceRepository
	ActiveSubscriptionFinder   ActiveSubscriptionFinder
	ClientSubscriptionLister   ClientSubscriptionLister
	SubscriptionTypeCatalog    SubscriptionTypeCatalog
	SubscriptionSeller         SubscriptionSeller
	PaymentStatusUpdater       AppointmentPaymentStatusUpdater
	AppointmentPaymentOperator AppointmentPaymentOperator
	PaymentRollbackRepo        SubscriptionPaymentRollbackRepository
}

func NewService(deps Dependencies) *Service {
	return &Service{
		txManager:                  deps.TxManager,
		subscriptionVisitWriter:    deps.SubscriptionVisitWriter,
		subscriptionBalanceRepo:    deps.SubscriptionBalanceRepo,
		activeSubscriptionFinder:   deps.ActiveSubscriptionFinder,
		clientSubscriptionLister:   deps.ClientSubscriptionLister,
		subscriptionTypeCatalog:    deps.SubscriptionTypeCatalog,
		subscriptionSeller:         deps.SubscriptionSeller,
		paymentStatusUpdater:       deps.PaymentStatusUpdater,
		appointmentPaymentOperator: deps.AppointmentPaymentOperator,
		paymentRollbackRepo:        deps.PaymentRollbackRepo,
	}
}

type AddSubscriptionVisitRequest struct {
	SubscriptionID int
	AppointmentID  int
	VisitDate      time.Time
}

type ActiveSubscription struct {
	SubscriptionID int
	CurrentBalance int
}

type ClientSubscription struct {
	SubscriptionID int
	CurrentBalance int
}

type SubscriptionType struct {
	ID            int
	Name          string
	Cost          float64
	SessionsCount int
	ServiceIDs    []int
}

type SellSubscriptionRequest struct {
	ClientID           int
	SubscriptionTypeID int
	Cost               float64
	CurrentBalance     int
}

type UpdateAppointmentPaymentRequest struct {
	AppointmentID int
	ClientID      int
	PaymentStatus string
	Amount        int
}

type AppointmentPaymentOperation struct {
	DocumentNumber   string
	ClientID         int
	Purpose          string
	Cashbox          string
	Amount           int
	CashboxBalance   int
	ServiceOrProduct string
	AppointmentID    int
}

func (s *Service) AddSubscriptionVisit(ctx context.Context, req AddSubscriptionVisitRequest) error {
	if req.SubscriptionID <= 0 || req.AppointmentID <= 0 || req.VisitDate.IsZero() {
		return fmt.Errorf("%w: subscription_id, appointment_id and visit_date are required", ErrInvalidInput)
	}

	return s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.subscriptionVisitWriter.AddSubscriptionVisit(txCtx, req.SubscriptionID, req.VisitDate, req.AppointmentID); err != nil {
			return err
		}

		if err := s.subscriptionBalanceRepo.DecreaseSubscriptionBalance(txCtx, req.SubscriptionID); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) GetActiveSubscription(ctx context.Context, clientID, serviceID int) (ActiveSubscription, error) {
	if clientID <= 0 || serviceID <= 0 {
		return ActiveSubscription{}, fmt.Errorf("%w: client_id and service_id must be positive", ErrInvalidInput)
	}

	return s.activeSubscriptionFinder.FindActiveSubscription(ctx, clientID, serviceID)
}

func (s *Service) GetClientSubscriptions(ctx context.Context, clientID int) ([]ClientSubscription, error) {
	if clientID <= 0 {
		return nil, fmt.Errorf("%w: client_id must be positive", ErrInvalidInput)
	}

	return s.clientSubscriptionLister.ListClientSubscriptions(ctx, clientID)
}

func (s *Service) GetSubscriptionTypes(ctx context.Context) ([]SubscriptionType, error) {
	return s.subscriptionTypeCatalog.ListSubscriptionTypes(ctx)
}

func (s *Service) SellSubscription(ctx context.Context, req SellSubscriptionRequest) (string, error) {
	if req.ClientID <= 0 || req.SubscriptionTypeID <= 0 || req.Cost <= 0 || req.CurrentBalance <= 0 {
		return "", fmt.Errorf("%w: invalid sell subscription request", ErrInvalidInput)
	}

	subscriptionTypeName, err := s.subscriptionTypeCatalog.GetSubscriptionTypeName(ctx, req.SubscriptionTypeID)
	if err != nil {
		return "", err
	}

	documentNumber := fmt.Sprintf("SUB-%d-%d", req.ClientID, time.Now().Unix())
	purpose := fmt.Sprintf("Покупка абонемента: %s", subscriptionTypeName)
	cashbox := "Основная касса"

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.subscriptionSeller.CreateSubscriptionSaleOperation(txCtx, req.ClientID, req.Cost, purpose, documentNumber, cashbox); err != nil {
			return err
		}

		if err := s.subscriptionSeller.CreateSubscription(txCtx, req.SubscriptionTypeID, req.ClientID, req.Cost, req.CurrentBalance); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return subscriptionTypeName, nil
}

func (s *Service) UpdateAppointmentPaymentStatus(ctx context.Context, req UpdateAppointmentPaymentRequest) error {
	if req.AppointmentID <= 0 {
		return fmt.Errorf("%w: appointment_id must be positive", ErrInvalidInput)
	}
	if req.PaymentStatus == "" {
		return fmt.Errorf("%w: payment_status is required", ErrInvalidInput)
	}
	if req.PaymentStatus != "unpaid" && req.PaymentStatus != "paid" && req.PaymentStatus != "partially_paid" {
		return fmt.Errorf("%w: unsupported payment_status", ErrInvalidInput)
	}
	if req.PaymentStatus == "partially_paid" && req.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive for partially_paid", ErrInvalidInput)
	}

	return s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.paymentStatusUpdater.UpdateAppointmentPaymentStatus(txCtx, req.AppointmentID, req.PaymentStatus); err != nil {
			return err
		}

		switch req.PaymentStatus {
		case "paid":
			amount, err := s.appointmentPaymentOperator.GetAppointmentCost(txCtx, req.AppointmentID)
			if err != nil {
				return err
			}

			return s.appointmentPaymentOperator.CreateAppointmentPaymentOperation(txCtx, AppointmentPaymentOperation{
				DocumentNumber:   fmt.Sprintf("PAY-%d-%d", req.AppointmentID, time.Now().Unix()),
				ClientID:         req.ClientID,
				Purpose:          "Оплата за услугу",
				Cashbox:          "Основная касса",
				Amount:           amount,
				CashboxBalance:   amount,
				ServiceOrProduct: "service",
				AppointmentID:    req.AppointmentID,
			})
		case "partially_paid":
			return s.appointmentPaymentOperator.CreateAppointmentPaymentOperation(txCtx, AppointmentPaymentOperation{
				DocumentNumber:   fmt.Sprintf("PAY-%d-%d", req.AppointmentID, time.Now().Unix()),
				ClientID:         req.ClientID,
				Purpose:          "Оплата за услугу",
				Cashbox:          "Основная касса",
				Amount:           req.Amount,
				CashboxBalance:   req.Amount,
				ServiceOrProduct: "service",
				AppointmentID:    req.AppointmentID,
			})
		case "unpaid":
			if err := s.paymentRollbackRepo.DeleteFinancialOperationsByAppointment(txCtx, req.AppointmentID); err != nil {
				return err
			}

			subscriptionID, err := s.paymentRollbackRepo.GetSubscriptionIDByAppointment(txCtx, req.AppointmentID)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					return nil
				}
				return err
			}

			if err := s.paymentRollbackRepo.DeleteSubscriptionVisitByAppointment(txCtx, req.AppointmentID); err != nil {
				return err
			}

			return s.subscriptionBalanceRepo.RestoreSubscriptionBalance(txCtx, subscriptionID)
		default:
			return nil
		}
	})
}
