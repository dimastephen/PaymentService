package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type TransitionError struct {
	From      PaymentStatus `json:"from"`
	To        PaymentStatus `json:"to"`
	PaymentId uuid.UUID     `json:"payment_id"`
}

type ValidationError struct {
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

func (e TransitionError) Error() string {
	return fmt.Sprintf("cannot transition from %s to %s for payment id:%v \n", e.From, e.To, e.PaymentId)
}
