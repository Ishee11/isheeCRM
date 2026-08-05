package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/clients"
	"github.com/jackc/pgx/v5"
)

func (r *ClientsRepository) GetHistory(ctx context.Context, clientID int) (clients.History, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return clients.History{}, fmt.Errorf("begin client history transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var foundClientID int
	if err := tx.QueryRow(ctx, `
		SELECT clients_id FROM clients
		WHERE clients_id = $1 AND deleted_at IS NULL
	`, clientID).Scan(&foundClientID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return clients.History{}, fmt.Errorf("%w: client not found", clients.ErrNotFound)
		}
		return clients.History{}, fmt.Errorf("find client for history: %w", err)
	}

	history := clients.History{
		ClientID:      foundClientID,
		Visits:        make([]clients.Visit, 0),
		Payments:      make([]clients.Payment, 0),
		Subscriptions: make([]clients.Subscription, 0),
	}

	if err := loadClientVisits(ctx, tx, clientID, &history); err != nil {
		return clients.History{}, err
	}
	if err := loadClientPayments(ctx, tx, clientID, &history); err != nil {
		return clients.History{}, err
	}
	if err := loadClientSubscriptions(ctx, tx, clientID, &history); err != nil {
		return clients.History{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return clients.History{}, fmt.Errorf("commit client history transaction: %w", err)
	}
	return history, nil
}

func loadClientVisits(ctx context.Context, tx pgx.Tx, clientID int, history *clients.History) error {
	rows, err := tx.Query(ctx, `
		SELECT a.id, a.service_id, COALESCE(s.name, ''), a.start_time,
		       COALESCE(a.appointment_status, ''), COALESCE(a.payment_status, ''),
		       COALESCE(a.cost, 0), sv.subscription_id
		FROM appointments a
		LEFT JOIN services s ON s.service_id = a.service_id
		LEFT JOIN subscription_visits sv ON sv.appointment_id = a.id
		WHERE a.client_id = $1 AND a.deleted_at IS NULL
		ORDER BY a.start_time DESC, a.id DESC
	`, clientID)
	if err != nil {
		return fmt.Errorf("query client visits: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var visit clients.Visit
		if err := rows.Scan(&visit.AppointmentID, &visit.ServiceID, &visit.ServiceName,
			&visit.StartTime, &visit.AppointmentStatus, &visit.PaymentStatus,
			&visit.Cost, &visit.SubscriptionID); err != nil {
			return fmt.Errorf("scan client visit: %w", err)
		}
		history.Visits = append(history.Visits, visit)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate client visits: %w", err)
	}
	return nil
}

func loadClientPayments(ctx context.Context, tx pgx.Tx, clientID int, history *clients.History) error {
	rows, err := tx.Query(ctx, `
		SELECT id, operation_date, amount, COALESCE(purpose, ''),
		       COALESCE(service_or_product, ''), appointment_id,
		       document_number, COALESCE(cashbox, '')
		FROM financial_operations
		WHERE client_id = $1
		ORDER BY operation_date DESC NULLS LAST, id DESC
	`, clientID)
	if err != nil {
		return fmt.Errorf("query client payments: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var payment clients.Payment
		if err := rows.Scan(&payment.ID, &payment.OperationDate, &payment.Amount,
			&payment.Purpose, &payment.ServiceOrProduct, &payment.AppointmentID,
			&payment.DocumentNumber, &payment.Cashbox); err != nil {
			return fmt.Errorf("scan client payment: %w", err)
		}
		history.Payments = append(history.Payments, payment)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate client payments: %w", err)
	}
	return nil
}

func loadClientSubscriptions(ctx context.Context, tx pgx.Tx, clientID int, history *clients.History) error {
	rows, err := tx.Query(ctx, `
		SELECT s.subscriptions_id, st.subscription_types_id, st.name, s.sale_date,
		       COALESCE(s.cost, st.cost), st.sessions_count, COUNT(sv.id)::integer,
		       s.current_balance, COALESCE(s.status, ''), s.deleted_at
		FROM subscriptions s
		JOIN subscription_types st ON st.subscription_types_id = s.subscription_types_id
		LEFT JOIN subscription_visits sv ON sv.subscription_id = s.subscriptions_id
		WHERE s.client_id = $1
		GROUP BY s.subscriptions_id, st.subscription_types_id, st.name, st.cost, st.sessions_count
		ORDER BY s.sale_date DESC NULLS LAST, s.subscriptions_id DESC
	`, clientID)
	if err != nil {
		return fmt.Errorf("query client subscriptions: %w", err)
	}

	subscriptionIndexes := make(map[int]int)
	for rows.Next() {
		var subscription clients.Subscription
		if err := rows.Scan(&subscription.ID, &subscription.TypeID, &subscription.TypeName,
			&subscription.PurchasedAt, &subscription.Cost, &subscription.TypeSessionsCount,
			&subscription.UsedCount, &subscription.CurrentBalance, &subscription.Status,
			&subscription.DeletedAt); err != nil {
			rows.Close()
			return fmt.Errorf("scan client subscription: %w", err)
		}
		subscription.SessionsCount = subscription.UsedCount + subscription.CurrentBalance
		subscription.Visits = make([]clients.SubscriptionVisit, 0)
		subscriptionIndexes[subscription.ID] = len(history.Subscriptions)
		history.Subscriptions = append(history.Subscriptions, subscription)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate client subscriptions: %w", err)
	}
	rows.Close()

	visitRows, err := tx.Query(ctx, `
		SELECT sv.id, sv.subscription_id, sv.appointment_id, sv.visit_date,
		       a.service_id, ser.name
		FROM subscription_visits sv
		JOIN subscriptions s ON s.subscriptions_id = sv.subscription_id
		LEFT JOIN appointments a ON a.id = sv.appointment_id
		LEFT JOIN services ser ON ser.service_id = a.service_id
		WHERE s.client_id = $1
		ORDER BY sv.visit_date, sv.id
	`, clientID)
	if err != nil {
		return fmt.Errorf("query client subscription visits: %w", err)
	}
	defer visitRows.Close()
	for visitRows.Next() {
		var subscriptionID int
		var visit clients.SubscriptionVisit
		if err := visitRows.Scan(&visit.ID, &subscriptionID, &visit.AppointmentID,
			&visit.VisitDate, &visit.ServiceID, &visit.ServiceName); err != nil {
			return fmt.Errorf("scan client subscription visit: %w", err)
		}
		if index, ok := subscriptionIndexes[subscriptionID]; ok {
			history.Subscriptions[index].Visits = append(history.Subscriptions[index].Visits, visit)
		}
	}
	if err := visitRows.Err(); err != nil {
		return fmt.Errorf("iterate client subscription visits: %w", err)
	}
	return nil
}
