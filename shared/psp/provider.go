package psp

import "context"

type Provider interface {
	Charge(ctx context.Context, request *TransactionRequest) (*TransactionResponse, error)
}
