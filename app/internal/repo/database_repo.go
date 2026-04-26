package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"hotel.com/app/internal/models"
)

type databaseRepo struct {
	db *pgxpool.Pool
}

func NewDatabaseRepo(conn *pgxpool.Pool) ServiceRepository {
	return &databaseRepo{
		db: conn,
	}
}

func (dbr *databaseRepo) DbPing() error {
	err := dbr.db.Ping(context.Background())
	return err
}

func (dbr *databaseRepo) CreatePayment(ctx context.Context, p *models.Payment) error {
	query := `
		INSERT INTO payments (id, booking_id, stripe_payment_intent_id, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := dbr.db.Exec(ctx, query,
		p.ID, p.BookingID, p.StripePaymentIntentID, p.Amount, p.Status, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (dbr *databaseRepo) UpdatePaymentStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := dbr.db.Exec(ctx, query, status, id)
	return err
}

func (dbr *databaseRepo) GetPaymentByBookingID(ctx context.Context, bookingID string) (*models.Payment, error) {
	query := `
		SELECT id, booking_id, stripe_payment_intent_id, amount, status, created_at, updated_at
		FROM payments
		WHERE booking_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var p models.Payment
	err := dbr.db.QueryRow(ctx, query, bookingID).Scan(
		&p.ID, &p.BookingID, &p.StripePaymentIntentID, &p.Amount, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
