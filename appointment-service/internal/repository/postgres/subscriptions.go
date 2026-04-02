package postgres

import (
	"appointment-service/internal/usecase/subscriptions"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionsRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionsRepository(pool *pgxpool.Pool) *SubscriptionsRepository {
	return &SubscriptionsRepository{pool: pool}
}

func (r *SubscriptionsRepository) Create(ctx context.Context, req subscriptions.CreateRequest) (subscriptions.SubscriptionType, error) {
	query := `
		INSERT INTO subscription_types (name, cost, sessions_count, service_ids)
		VALUES ($1, $2, $3, $4)
		RETURNING subscription_types_id, name, cost, sessions_count, service_ids
	`

	var result subscriptions.SubscriptionType
	err := r.pool.QueryRow(ctx, query, req.Name, req.Cost, req.SessionsCount, req.ServiceIDs).Scan(
		&result.ID,
		&result.Name,
		&result.Cost,
		&result.SessionsCount,
		&result.ServiceIDs,
	)
	if err != nil {
		return subscriptions.SubscriptionType{}, fmt.Errorf("create subscription type: %w", err)
	}

	return result, nil
}

var _ subscriptions.Repository = (*SubscriptionsRepository)(nil)
