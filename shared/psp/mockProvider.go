package psp

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type FakeProvider struct {
}

func NewFakeProvider() *FakeProvider {
	return &FakeProvider{}
}

func (p *FakeProvider) Charge(ctx context.Context, request *TransactionRequest) (*TransactionResponse, error) {
	num := rand.Intn(2)
	time.Sleep(time.Millisecond * 800)
	switch num {
	case 0:
		return &TransactionResponse{
			PspTransactionID: uuid.New().String(),
			Status:           StatusSuccess,
			ErrorMessage:     "",
		}, nil
	case 1:
		return &TransactionResponse{
			PspTransactionID: "",
			Status:           StatusFailed,
			ErrorMessage:     "Failed to charge",
		}, nil
	}
	return nil, errors.New("failed to charge")
}
