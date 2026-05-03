package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/appointments"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppointmentsRepository struct {
	pool *pgxpool.Pool
}

func NewAppointmentsRepository(pool *pgxpool.Pool) *AppointmentsRepository {
	return &AppointmentsRepository{pool: pool}
}

func (r *AppointmentsRepository) Create(ctx context.Context, req appointments.CreateRequest) (appointments.Appointment, error) {
	query := `
		INSERT INTO appointments (service_id, client_id, start_time)
		VALUES ($1, $2, $3)
		RETURNING id, service_id, client_id, start_time, payment_status, appointment_status
	`

	var result appointments.Appointment
	err := r.pool.QueryRow(ctx, query, req.ServiceID, req.ClientID, req.StartTime).Scan(
		&result.ID,
		&result.ServiceID,
		&result.ClientID,
		&result.StartTime,
		&result.PaymentStatus,
		&result.AppointmentStatus,
	)
	if err != nil {
		return appointments.Appointment{}, err
	}

	return result, nil
}

func (r *AppointmentsRepository) List(ctx context.Context, filter appointments.ListFilter) ([]appointments.Appointment, error) {
	query := `
		SELECT
			a.id,
			a.service_id,
			a.client_id,
			a.start_time,
			a.payment_status,
			a.appointment_status,
			s.name,
			c.name
		FROM appointments a
		JOIN services s ON s.service_id = a.service_id
		LEFT JOIN clients c ON c.clients_id = a.client_id
		WHERE a.deleted_at IS NULL
		  AND s.deleted_at IS NULL
	`

	args := make([]any, 0, 4)
	argIndex := 1

	if filter.OnlyUnpaid {
		query += " AND a.payment_status = 'unpaid'"
	}
	if filter.ClientID > 0 {
		query += fmt.Sprintf(" AND a.client_id = $%d", argIndex)
		args = append(args, filter.ClientID)
		argIndex++
	}
	if filter.AppointmentStatus != "" {
		query += fmt.Sprintf(" AND a.appointment_status = $%d", argIndex)
		args = append(args, filter.AppointmentStatus)
		argIndex++
	}
	if filter.From != nil {
		query += fmt.Sprintf(" AND a.start_time >= $%d", argIndex)
		args = append(args, *filter.From)
		argIndex++
	}
	if filter.To != nil {
		query += fmt.Sprintf(" AND a.start_time < $%d", argIndex)
		args = append(args, filter.To.AddDate(0, 0, 1))
		argIndex++
	}

	query += " ORDER BY a.start_time"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}
	defer rows.Close()

	var result []appointments.Appointment
	for rows.Next() {
		var appointment appointments.Appointment
		if err := rows.Scan(
			&appointment.ID,
			&appointment.ServiceID,
			&appointment.ClientID,
			&appointment.StartTime,
			&appointment.PaymentStatus,
			&appointment.AppointmentStatus,
			&appointment.ServiceName,
			&appointment.ClientName,
		); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		result = append(result, appointment)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate appointments: %w", rows.Err())
	}

	return result, nil
}

func (r *AppointmentsRepository) UpdateStatus(ctx context.Context, appointmentID int, status string) (appointments.AppointmentStatusUpdate, error) {
	query := `
		UPDATE appointments
		SET appointment_status = $1
		WHERE id = $2
		  AND deleted_at IS NULL
		RETURNING id, client_id, appointment_status
	`

	var result appointments.AppointmentStatusUpdate
	err := r.pool.QueryRow(ctx, query, status, appointmentID).Scan(
		&result.ID,
		&result.ClientID,
		&result.AppointmentStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appointments.AppointmentStatusUpdate{}, fmt.Errorf("%w: appointment not found", appointments.ErrNotFound)
		}
		return appointments.AppointmentStatusUpdate{}, fmt.Errorf("update appointment status: %w", err)
	}

	return result, nil
}

func (r *AppointmentsRepository) Move(ctx context.Context, appointmentID int, startTime time.Time) (appointments.AppointmentMoveResult, error) {
	query := `
		UPDATE appointments
		SET start_time = $1
		WHERE id = $2
		  AND deleted_at IS NULL
		RETURNING id, start_time
	`

	var result appointments.AppointmentMoveResult
	err := r.pool.QueryRow(ctx, query, startTime, appointmentID).Scan(&result.ID, &result.StartTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appointments.AppointmentMoveResult{}, fmt.Errorf("%w: appointment not found", appointments.ErrNotFound)
		}
		return appointments.AppointmentMoveResult{}, fmt.Errorf("move appointment: %w", err)
	}

	return result, nil
}

func (r *AppointmentsRepository) SoftDelete(ctx context.Context, appointmentID int) error {
	query := `
		UPDATE appointments
		SET deleted_at = NOW()
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	tag, err := r.pool.Exec(ctx, query, appointmentID)
	if err != nil {
		return fmt.Errorf("soft delete appointment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: appointment not found", appointments.ErrNotFound)
	}

	return nil
}

var _ appointments.Repository = (*AppointmentsRepository)(nil)
