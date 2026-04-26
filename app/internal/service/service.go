// Package service contains the business logic layer of the application.
// It defines service interfaces and implements use cases by orchestrating
// repositories, applying business rules, and returning results to handlers.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"hotel.com/app/internal/models"
	"hotel.com/app/internal/repo"
)

type Service interface {
	Check(ctx context.Context) error
	ProcessPayment(ctx context.Context, req *models.ProcessPaymentRequest) (*models.PaymentResponse, error)
}

type paymentService struct {
	l *slog.Logger
	r repo.ServiceRepository
}

func (s *paymentService) Check(ctx context.Context) error {
	s.l.Info("Pinging db...")
	err := s.r.DbPing()
	if err != nil {
		s.l.Error("DbPing failed", "error", err)
	}
	return err
}

func (s *paymentService) ProcessPayment(ctx context.Context, req *models.ProcessPaymentRequest) (*models.PaymentResponse, error) {
	s.l.Info("Processing payment", "booking_id", req.BookingID, "amount", req.Amount)

	// Create a new Payment record in pending status
	paymentID := uuid.New().String()
	paymentRecord := &models.Payment{
		ID:        paymentID,
		BookingID: req.BookingID,
		Amount:    req.Amount,
		Status:    "pending",
	}

	err := s.r.CreatePayment(ctx, paymentRecord)
	if err != nil {
		s.l.Error("Failed to create payment record", "error", err)
		return nil, fmt.Errorf("failed to initialize payment record")
	}

	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create and confirm PaymentIntent
	// Amount needs to be converted to cents for Stripe
	amountCents := int64(req.Amount * 100)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		PaymentMethod: stripe.String(req.PaymentMethodID),
		Confirm: stripe.Bool(true),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
		Description: stripe.String(fmt.Sprintf("Booking ID: %s", req.BookingID)),
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		s.l.Error("Stripe PaymentIntent failed", "error", err)
		
		// Update DB record to failed
		paymentRecord.Status = "failed"
		_ = s.r.UpdatePaymentStatus(ctx, paymentID, "failed")
		
		if stripeErr, ok := err.(*stripe.Error); ok {
			return nil, fmt.Errorf("payment failed: %s", stripeErr.Msg)
		}
		return nil, fmt.Errorf("payment processing failed")
	}

	status := "pending"
	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		status = "succeeded"
	} else if pi.Status == stripe.PaymentIntentStatusRequiresAction || pi.Status == stripe.PaymentIntentStatusRequiresConfirmation {
		// Since we set AllowRedirects to never, this shouldn't happen for basic cards, 
		// but handled just in case.
		status = "requires_action"
	} else {
		status = "failed"
	}

	// Update record with intent ID and new status
	paymentRecord.StripePaymentIntentID = pi.ID
	paymentRecord.Status = status

	// We only have UpdatePaymentStatus in repo, let's just update the status for now
	// Ideally we should have an UpdatePayment that updates all mutable fields
	err = s.r.UpdatePaymentStatus(ctx, paymentID, status)
	if err != nil {
		s.l.Error("Failed to update payment status", "error", err)
	}

	response := &models.PaymentResponse{
		PaymentID: paymentID,
		BookingID: req.BookingID,
		Amount:    req.Amount,
		Status:    status,
	}

	if status == "failed" {
		return response, fmt.Errorf("payment failed with status: %s", pi.Status)
	}

	return response, nil
}

func New(l *slog.Logger, r repo.ServiceRepository) Service {
	return &paymentService{
		l: l,
		r: r,
	}
}
