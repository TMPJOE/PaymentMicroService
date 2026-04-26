package models

import "time"

// Payment represents a payment transaction in the system
type Payment struct {
	ID                    string    `json:"id" db:"id"`
	BookingID             string    `json:"booking_id" db:"booking_id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id" db:"stripe_payment_intent_id"`
	Amount                float64   `json:"amount" db:"amount"`
	Status                string    `json:"status" db:"status"` // e.g., "pending", "succeeded", "failed"
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// ProcessPaymentRequest is the request to process a new payment
type ProcessPaymentRequest struct {
	BookingID       string  `json:"booking_id" validate:"required,uuid"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	PaymentMethodID string  `json:"payment_method_id" validate:"required"` // e.g., "pm_card_visa"
}

// PaymentResponse represents the response sent to the client
type PaymentResponse struct {
	PaymentID string  `json:"payment_id"`
	BookingID string  `json:"booking_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

// ErrorResponse represents an error sent to the client
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
