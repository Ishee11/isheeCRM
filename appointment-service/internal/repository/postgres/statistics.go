package postgres

import (
	"appointment-service/internal/usecase/statistics"
	"context"
	"fmt"
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

var _ statistics.Repository = (*StatisticsRepository)(nil)
