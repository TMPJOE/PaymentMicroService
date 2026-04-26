// Package handler provides HTTP request handlers, routing, and middleware.
// It handles incoming HTTP requests, delegates to the service layer for
// business logic, and returns JSON responses with appropriate status codes.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"hotel.com/app/internal/helper"
	"hotel.com/app/internal/models"
	"hotel.com/app/internal/service"
)

type Handler struct {
	s       service.Service
	l       *slog.Logger
	jwtAuth *JWTAuthenticator
}

func New(s service.Service, l *slog.Logger, jwtAuth *JWTAuthenticator) *Handler {
	return &Handler{
		s:       s,
		l:       l,
		jwtAuth: jwtAuth,
	}
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// readinessCheck verifies if the service is ready to accept traffic
// by pinging the database and other critical dependencies.
func (h *Handler) readinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Check(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "reason": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
		"db":     "ok",
	})
}

// processPayment handles POST /payments/process
func (h *Handler) processPayment(w http.ResponseWriter, r *http.Request) {
	var req models.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error("Failed to decode process payment request", "error", err)
		helper.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validator := helper.NewValidator()
	if err := validator.Validate(&req); err != nil {
		h.l.Error("Validation failed", "err", err)
		helper.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.s.ProcessPayment(r.Context(), &req)
	if err != nil {
		h.l.Error("Payment processing failed", "error", err)
		helper.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
