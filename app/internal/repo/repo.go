// Package repo implements the data access layer of the application.
// It handles all database queries, transactions, and data mapping,
// providing a clean interface for the service layer to interact with PostgreSQL.
package repo

import (
	"context"
	"hotel.com/app/internal/models"
)

type ServiceRepository interface {
	DbPing() error
	CreatePayment(ctx context.Context, payment *models.Payment) error
	UpdatePaymentStatus(ctx context.Context, id string, status string) error
	GetPaymentByBookingID(ctx context.Context, bookingID string) (*models.Payment, error)
}
