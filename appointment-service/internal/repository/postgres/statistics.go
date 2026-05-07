package postgres

import (
	"context"
	"fmt"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/statistics"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatisticsRepository struct {
	pool *pgxpool.Pool
}

func NewStatisticsRepository(pool *pgxpool.Pool) *StatisticsRepository {
	return &StatisticsRepository{pool: pool}
}

func (r *StatisticsRepository) GetByPeriod(ctx context.Context, startDate, endDate time.Time) (statistics.Summary, error) {
	query := `
		WITH all_visits AS (
			SELECT operation_date::date AS visit_date
			FROM financial_operations
			WHERE operation_date::date BETWEEN $1::date AND $2::date
		
			UNION ALL
		
			SELECT sv.visit_date
			FROM subscription_visits sv
			LEFT JOIN appointments a ON a.id = sv.appointment_id
			WHERE sv.visit_date BETWEEN $1::date AND $2::date
			  AND (sv.appointment_id IS NULL OR a.deleted_at IS NULL)
		)
		SELECT
			(SELECT COUNT(*) FROM all_visits) AS total_visits,
			COALESCE(SUM(fo.amount), 0) AS total_earnings,
			COALESCE(SUM(CASE WHEN fo.service_or_product = 'service' THEN fo.amount ELSE 0 END), 0) AS total_services,
			COALESCE(SUM(CASE WHEN fo.service_or_product = 'subscription' THEN fo.amount ELSE 0 END), 0) AS total_subscriptions
		FROM financial_operations fo
		WHERE fo.operation_date::date BETWEEN $1::date AND $2::date
		  AND (
			fo.appointment_id IS NULL OR EXISTS (
				SELECT 1
				FROM appointments a
				WHERE a.id = fo.appointment_id
				  AND a.deleted_at IS NULL
			)
		  );
	`

	var summary statistics.Summary
	if err := r.pool.QueryRow(ctx, query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Scan(
		&summary.TotalVisits,
		&summary.TotalEarnings,
		&summary.TotalServices,
		&summary.TotalSubscriptions,
	); err != nil {
		return statistics.Summary{}, fmt.Errorf("get statistics by period: %w", err)
	}

	return summary, nil
}

func (r *StatisticsRepository) GetClientPayments(ctx context.Context) ([]statistics.ClientPaymentStats, error) {
	query := `
		SELECT
			cl.clients_id,
			COALESCE(cl.name, '') AS name,
			COUNT(*) AS count,
			AVG(fo.amount) AS avg_amount,
			SUM(fo.amount) AS paid
		FROM clients AS cl
		JOIN appointments AS ap ON ap.client_id = cl.clients_id
		JOIN financial_operations AS fo ON fo.appointment_id = ap.id
		GROUP BY cl.clients_id, cl.name
		ORDER BY paid DESC, avg_amount DESC;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get client payments: %w", err)
	}
	defer rows.Close()

	result := make([]statistics.ClientPaymentStats, 0)
	for rows.Next() {
		var item statistics.ClientPaymentStats
		if err := rows.Scan(
			&item.ClientID,
			&item.Name,
			&item.Count,
			&item.AvgAmount,
			&item.Paid,
		); err != nil {
			return nil, fmt.Errorf("scan client payment stats: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate client payment stats: %w", err)
	}

	return result, nil
}

var _ statistics.Repository = (*StatisticsRepository)(nil)
