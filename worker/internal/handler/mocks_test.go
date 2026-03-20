package handler

import (
	"context"

	"github.com/google/uuid"
	"github.com/payment-service/shared/domain"
	"github.com/payment-service/shared/psp"
	"github.com/payment-service/shared/storage"
	"github.com/stretchr/testify/mock"
)

type mockPSP struct {
	mock.Mock
}

func (m *mockPSP) Charge(ctx context.Context, req *psp.TransactionRequest) (*psp.TransactionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*psp.TransactionResponse), args.Error(1)
}

type mockCommandRepo struct {
	mock.Mock
}

func (m *mockCommandRepo) IsProcessed(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockCommandRepo) Save(ctx context.Context, key string, id uuid.UUID) error {
	args := m.Called(ctx, key, id)
	return args.Error(0)
}

type mockEventRepo struct {
	mock.Mock
}

func (m *mockEventRepo) Add(ctx context.Context, paymentId uuid.UUID, event string, payload []byte) error {
	args := m.Called(ctx, paymentId, event, payload)
	return args.Error(0)
}

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

func (m *mockPaymentRepo) Get(ctx context.Context, key string) (*domain.Payment, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*domain.Payment), args.Error(1)
}

type mockProducer struct {
	mock.Mock
}

func (m *mockProducer) SendMessage(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	args := m.Called(ctx, topic, key, value, headers)
	return args.Error(0)
}

func (m *mockProducer) Close() error { return nil }

type mockStore struct {
	mock.Mock
}

func (m *mockStore) WithTransaction(ctx context.Context, fn func(tx storage.TXDB) error) error {
	m.Called(ctx, mock.AnythingOfType("func(storage.TXDB) error"))
	return nil
}
