package handler

import (
	"context"

	"github.com/payment-service/shared/domain"
	"github.com/stretchr/testify/mock"
)

type mockPaymentRepo struct {
	mock.Mock
}

func (m *mockPaymentRepo) Create(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *mockPaymentRepo) Update(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *mockPaymentRepo) Get(ctx context.Context, id string) (*domain.Payment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Payment), args.Error(1)
}
