package postgres

import (
	"appointment-service/database"
	"appointment-service/internal/usecase/billing"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txContextKey struct{}

type BillingRepository struct {
	pool *pgxpool.Pool
}

func NewBillingRepository(pool *pgxpool.Pool) *BillingRepository {
	return &BillingRepository{pool: pool}
}

func (r *BillingRepository) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *BillingRepository) AddSubscriptionVisit(ctx context.Context, subscriptionID int, visitDate time.Time, appointmentID int) error {
	query := `
		INSERT INTO subscription_visits (subscription_id, visit_date, appointment_id)
		VALUES ($1, $2, $3)
	`
	_, err := r.exec(ctx, query, subscriptionID, visitDate, appointmentID)
	if err != nil {
		return fmt.Errorf("ошибка добавления посещения: %w", err)
	}
	return nil
}

func (r *BillingRepository) DecreaseSubscriptionBalance(ctx context.Context, subscriptionID int) error {
	query := `
		UPDATE subscriptions
		SET current_balance = current_balance - 1
		WHERE subscriptions_id = $1 AND current_balance > 0
		RETURNING current_balance
	`

	var currentBalance int
	err := r.queryRow(ctx, query, subscriptionID).Scan(&currentBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: active subscription with balance not found", billing.ErrNotFound)
		}
		return fmt.Errorf("ошибка уменьшения баланса: %w", err)
	}

	return nil
}

func (r *BillingRepository) FindActiveSubscription(ctx context.Context, clientID, serviceID int) (billing.ActiveSubscription, error) {
	query := `
		SELECT s.subscriptions_id, s.current_balance
		FROM subscriptions s
		JOIN subscription_type_services sts ON sts.subscription_type_id = s.subscription_types_id
		WHERE s.client_id = $1
		  AND sts.service_id = $2
		  AND s.current_balance > 0
		  AND s.deleted_at IS NULL
		LIMIT 1
	`

	var result billing.ActiveSubscription
	err := r.queryRow(ctx, query, clientID, serviceID).Scan(&result.SubscriptionID, &result.CurrentBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return billing.ActiveSubscription{}, fmt.Errorf("%w: active subscription not found", billing.ErrNotFound)
		}
		return billing.ActiveSubscription{}, fmt.Errorf("find active subscription: %w", err)
	}

	return result, nil
}

func (r *BillingRepository) ListClientSubscriptions(ctx context.Context, clientID int) ([]billing.ClientSubscription, error) {
	query := `
		SELECT subscriptions_id, current_balance
		FROM subscriptions
		WHERE client_id = $1
		  AND deleted_at IS NULL
	`

	rows, err := r.query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []billing.ClientSubscription
	for rows.Next() {
		var subscription billing.ClientSubscription
		if err := rows.Scan(&subscription.SubscriptionID, &subscription.CurrentBalance); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, subscription)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", rows.Err())
	}

	return subscriptions, nil
}

func (r *BillingRepository) ListSubscriptionTypes(ctx context.Context) ([]billing.SubscriptionType, error) {
	query := `
		SELECT
			st.subscription_types_id,
			st.name,
			st.cost,
			st.sessions_count,
			COALESCE(
				array_agg(sts.service_id ORDER BY sts.service_id)
				FILTER (WHERE sts.service_id IS NOT NULL),
				'{}'::integer[]
			) AS service_ids
		FROM subscription_types st
		LEFT JOIN subscription_type_services sts
			ON sts.subscription_type_id = st.subscription_types_id
		GROUP BY st.subscription_types_id, st.name, st.cost, st.sessions_count
	`

	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list subscription types: %w", err)
	}
	defer rows.Close()

	var types []billing.SubscriptionType
	for rows.Next() {
		var subscriptionType billing.SubscriptionType
		if err := rows.Scan(
			&subscriptionType.ID,
			&subscriptionType.Name,
			&subscriptionType.Cost,
			&subscriptionType.SessionsCount,
			&subscriptionType.ServiceIDs,
		); err != nil {
			return nil, fmt.Errorf("scan subscription type: %w", err)
		}
		types = append(types, subscriptionType)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate subscription types: %w", rows.Err())
	}

	return types, nil
}

func (r *BillingRepository) GetSubscriptionTypeName(ctx context.Context, subscriptionTypeID int) (string, error) {
	query := `SELECT name FROM subscription_types WHERE subscription_types_id = $1`

	var name string
	err := r.queryRow(ctx, query, subscriptionTypeID).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%w: subscription type not found", billing.ErrNotFound)
		}
		return "", fmt.Errorf("get subscription type name: %w", err)
	}

	return name, nil
}

func (r *BillingRepository) CreateSubscriptionSaleOperation(ctx context.Context, clientID int, amount float64, purpose, documentNumber, cashbox string) error {
	query := `
		INSERT INTO financial_operations (
			client_id, service_or_product, amount, purpose, document_number, cashbox
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.exec(ctx, query, clientID, "subscription", amount, purpose, documentNumber, cashbox)
	if err != nil {
		return fmt.Errorf("create subscription sale operation: %w", err)
	}

	return nil
}

func (r *BillingRepository) CreateSubscription(ctx context.Context, subscriptionTypeID, clientID int, cost float64, currentBalance int) error {
	query := `
		INSERT INTO subscriptions (
			subscription_types_id, client_id, cost, current_balance
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.exec(ctx, query, subscriptionTypeID, clientID, cost, currentBalance)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	return nil
}

func (r *BillingRepository) UpdateAppointmentPaymentStatus(ctx context.Context, appointmentID int, status string) error {
	query := `
		UPDATE appointments
		SET payment_status = $1
		WHERE id = $2
	`

	_, err := r.exec(ctx, query, status, appointmentID)
	if err != nil {
		return fmt.Errorf("update appointment payment status: %w", err)
	}

	return nil
}

func (r *BillingRepository) GetAppointmentCost(ctx context.Context, appointmentID int) (int, error) {
	query := `
		SELECT cost
		FROM appointments
		WHERE id = $1
	`

	var amount int
	err := r.queryRow(ctx, query, appointmentID).Scan(&amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%w: appointment not found", billing.ErrNotFound)
		}
		return 0, fmt.Errorf("get appointment cost: %w", err)
	}

	return amount, nil
}

func (r *BillingRepository) CreateAppointmentPaymentOperation(ctx context.Context, op billing.AppointmentPaymentOperation) error {
	query := `
		INSERT INTO financial_operations (
			document_number, client_id, purpose, cashbox, amount, cashbox_balance, service_or_product, appointment_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.exec(
		ctx,
		query,
		op.DocumentNumber,
		op.ClientID,
		op.Purpose,
		op.Cashbox,
		op.Amount,
		op.CashboxBalance,
		op.ServiceOrProduct,
		op.AppointmentID,
	)
	if err != nil {
		return fmt.Errorf("create appointment payment operation: %w", err)
	}

	return nil
}

func (r *BillingRepository) DeleteFinancialOperationsByAppointment(ctx context.Context, appointmentID int) error {
	query := `
		DELETE FROM financial_operations
		WHERE appointment_id = $1
	`

	_, err := r.exec(ctx, query, appointmentID)
	if err != nil {
		return fmt.Errorf("delete financial operations by appointment: %w", err)
	}

	return nil
}

func (r *BillingRepository) GetSubscriptionIDByAppointment(ctx context.Context, appointmentID int) (int, error) {
	query := `
		SELECT subscription_id
		FROM subscription_visits
		WHERE appointment_id = $1
	`

	var subscriptionID int
	err := r.queryRow(ctx, query, appointmentID).Scan(&subscriptionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%w: subscription visit not found", billing.ErrNotFound)
		}
		return 0, fmt.Errorf("get subscription by appointment: %w", err)
	}

	return subscriptionID, nil
}

func (r *BillingRepository) DeleteSubscriptionVisitByAppointment(ctx context.Context, appointmentID int) error {
	query := `
		DELETE FROM subscription_visits
		WHERE appointment_id = $1
	`

	_, err := r.exec(ctx, query, appointmentID)
	if err != nil {
		return fmt.Errorf("delete subscription visit by appointment: %w", err)
	}

	return nil
}

func (r *BillingRepository) RestoreSubscriptionBalance(ctx context.Context, subscriptionID int) error {
	query := `
		UPDATE subscriptions
		SET current_balance = current_balance + 1
		WHERE subscriptions_id = $1
	`

	_, err := r.exec(ctx, query, subscriptionID)
	if err != nil {
		return fmt.Errorf("restore subscription balance: %w", err)
	}

	return nil
}

func (r *BillingRepository) exec(ctx context.Context, query string, args ...any) (pgconnCommandTag, error) {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Exec(ctx, query, args...)
	}
	return r.pool.Exec(ctx, query, args...)
}

func (r *BillingRepository) query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Query(ctx, query, args...)
	}
	return r.pool.Query(ctx, query, args...)
}

func (r *BillingRepository) queryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if tx, ok := txFromContext(ctx); ok {
		return tx.QueryRow(ctx, query, args...)
	}
	return r.pool.QueryRow(ctx, query, args...)
}

type pgconnCommandTag interface {
	RowsAffected() int64
}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx, ok
}

var _ billing.TxManager = (*BillingRepository)(nil)
var _ billing.SubscriptionVisitWriter = (*BillingRepository)(nil)
var _ billing.SubscriptionBalanceRepository = (*BillingRepository)(nil)
var _ billing.ActiveSubscriptionFinder = (*BillingRepository)(nil)
var _ billing.ClientSubscriptionLister = (*BillingRepository)(nil)
var _ billing.SubscriptionTypeCatalog = (*BillingRepository)(nil)
var _ billing.SubscriptionSeller = (*BillingRepository)(nil)
var _ billing.AppointmentPaymentStatusUpdater = (*BillingRepository)(nil)
var _ billing.AppointmentPaymentOperator = (*BillingRepository)(nil)
var _ billing.SubscriptionPaymentRollbackRepository = (*BillingRepository)(nil)

func NewBillingRepositoryFromGlobal() *BillingRepository {
	return NewBillingRepository(database.Pool)
}
