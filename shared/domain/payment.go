package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
	JPY Currency = "JPY"
	RUB Currency = "RUB"
)

var availableCurrencies = map[string]Currency{
	"USD": USD,
	"EUR": EUR,
	"GBP": GBP,
	"JPY": JPY,
	"RUB": RUB,
}

func (c Currency) isValid() error {
	_, ok := availableCurrencies[string(c)]
	if !ok {
		return fmt.Errorf("Currency %s is not available", string(c))
	}
	return nil
}

type PaymentStatus string

const (
	PaymentStatusNew        PaymentStatus = "NEW"
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	PaymentStatusCompleted  PaymentStatus = "COMPLETED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"
)

var stateMap = map[PaymentStatus][]PaymentStatus{
	PaymentStatusNew:        []PaymentStatus{PaymentStatusProcessing, PaymentStatusCancelled},
	PaymentStatusProcessing: []PaymentStatus{PaymentStatusCompleted, PaymentStatusFailed, PaymentStatusCancelled},
	PaymentStatusCompleted:  nil,
	PaymentStatusFailed:     []PaymentStatus{PaymentStatusCancelled},
	PaymentStatusCancelled:  nil,
}

type Payment struct {
	Id               uuid.UUID
	Status           PaymentStatus
	Amount           int64
	Currency         Currency
	IdempotencyKey   string
	PSPTransactionID string
	MerchantID       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ErrorMessage     string
}

func (p *Payment) TransitionTo(to PaymentStatus) error {
	transitions, ok := stateMap[p.Status]
	if !ok {
		return &TransitionError{
			From:      p.Status,
			To:        to,
			PaymentId: p.Id,
		}
	}
	for _, transition := range transitions {
		if transition == to {
			p.Status = to
			p.UpdatedAt = time.Now()
			return nil
		}
	}

	return &TransitionError{From: p.Status, To: to, PaymentId: p.Id}
}

func NewPayment(cmd CreatePayment) (*Payment, error) {
	if cmd.Amount <= 0 {
		return nil, &ValidationError{
			Message: "Amount must be greater than zero",
		}
	}
	if cmd.IdempotencyKey == "" {
		return nil, &ValidationError{
			Message: "IdempotencyKey must be set",
		}
	}

	if err := cmd.Currency.isValid(); err != nil {
		return nil, &ValidationError{
			Message: err.Error(),
		}
	}

	if len(cmd.MerchantID) == 0 {
		return nil, &ValidationError{
			Message: "MerchantID must be set",
		}
	}

	return &Payment{
		Id:             uuid.New(),
		Status:         PaymentStatusNew,
		Amount:         cmd.Amount,
		Currency:       cmd.Currency,
		IdempotencyKey: cmd.IdempotencyKey,
		MerchantID:     cmd.MerchantID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}
