package postgres

import (
	"context"
	"fmt"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServicesRepository struct {
	pool *pgxpool.Pool
}

func NewServicesRepository(pool *pgxpool.Pool) *ServicesRepository {
	return &ServicesRepository{pool: pool}
}

func (r *ServicesRepository) Create(ctx context.Context, req services.CreateRequest) (services.ServiceDTO, error) {
	query := `
		INSERT INTO services (name, duration, price)
		VALUES ($1, $2, $3)
		RETURNING service_id, name, duration, price
	`

	var service services.ServiceDTO
	if err := r.pool.QueryRow(ctx, query, req.Name, req.Duration, req.Price).Scan(
		&service.ID,
		&service.Name,
		&service.Duration,
		&service.Price,
	); err != nil {
		return services.ServiceDTO{}, fmt.Errorf("create service: %w", err)
	}

	return service, nil
}

func (r *ServicesRepository) List(ctx context.Context) ([]services.ServiceDTO, error) {
	query := `
		SELECT service_id, name, duration, price
		FROM services
		WHERE deleted_at IS NULL
		ORDER BY name, duration
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	var result []services.ServiceDTO
	for rows.Next() {
		var service services.ServiceDTO
		if err := rows.Scan(&service.ID, &service.Name, &service.Duration, &service.Price); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		result = append(result, service)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate services: %w", rows.Err())
	}

	return result, nil
}

func (r *ServicesRepository) SoftDelete(ctx context.Context, serviceID int) error {
	query := `
		UPDATE services
		SET deleted_at = NOW()
		WHERE service_id = $1 AND deleted_at IS NULL
	`

	tag, err := r.pool.Exec(ctx, query, serviceID)
	if err != nil {
		return fmt.Errorf("soft delete service: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: service not found", services.ErrNotFound)
	}

	return nil
}

var _ services.Repository = (*ServicesRepository)(nil)
