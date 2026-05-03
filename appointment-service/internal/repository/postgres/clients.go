package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/clients"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientsRepository struct {
	pool *pgxpool.Pool
}

func NewClientsRepository(pool *pgxpool.Pool) *ClientsRepository {
	return &ClientsRepository{pool: pool}
}

func (r *ClientsRepository) Create(ctx context.Context, phone, name string) (int, error) {
	query := `
		INSERT INTO clients (phone, name)
		VALUES ($1, $2)
		RETURNING clients_id
	`

	var clientID int
	if err := r.pool.QueryRow(ctx, query, phone, name).Scan(&clientID); err != nil {
		return 0, fmt.Errorf("create client: %w", err)
	}

	return clientID, nil
}

func (r *ClientsRepository) FindByPhone(ctx context.Context, phone string) (int, error) {
	query := `
		SELECT clients_id
		FROM clients
		WHERE phone = $1
		  AND deleted_at IS NULL
	`

	var clientID int
	if err := r.pool.QueryRow(ctx, query, phone).Scan(&clientID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%w: client not found", clients.ErrNotFound)
		}
		return 0, fmt.Errorf("find client by phone: %w", err)
	}

	return clientID, nil
}

func (r *ClientsRepository) GetInfo(ctx context.Context, clientID int) (clients.Info, error) {
	query := `
		SELECT 
			c.name,
			c.phone,
			c.email,
			c.categories,
			c.birth_date,
			cs.paid,
			cs.spent,
			c.gender,
			c.discount,
			cs.last_visit,
			cs.first_visit,
			cs.visit_count,
			c.comment
		FROM clients c
		LEFT JOIN client_stats cs ON cs.clients_id = c.clients_id
		WHERE c.clients_id = $1
		  AND c.deleted_at IS NULL
	`

	var info clients.Info
	err := r.pool.QueryRow(ctx, query, clientID).Scan(
		&info.Name,
		&info.Phone,
		&info.Email,
		&info.Categories,
		&info.BirthDate,
		&info.Paid,
		&info.Spent,
		&info.Gender,
		&info.Discount,
		&info.LastVisit,
		&info.FirstVisit,
		&info.VisitCount,
		&info.Comment,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return clients.Info{}, fmt.Errorf("%w: client not found", clients.ErrNotFound)
		}
		return clients.Info{}, fmt.Errorf("get client info: %w", err)
	}

	return info, nil
}

var _ clients.Repository = (*ClientsRepository)(nil)

var _ = time.Time{}
