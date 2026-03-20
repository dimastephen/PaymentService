package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPayment(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *CreatePayment
		wantErr bool
	}{
		{
			name: "success",
			cmd: &CreatePayment{
				Amount:         100,
				Currency:       "USD",
				IdempotencyKey: "test",
				MerchantID:     "test",
			},
			wantErr: false,
		},
		{
			name: "fail currency",
			cmd: &CreatePayment{
				Amount:         100,
				Currency:       "fake",
				IdempotencyKey: "test",
				MerchantID:     "test",
			},
			wantErr: true,
		},
		{
			name: "fail amount",
			cmd: &CreatePayment{
				Amount:         0,
				Currency:       "USD",
				IdempotencyKey: "test",
				MerchantID:     "test",
			},
			wantErr: true,
		},
		{
			name: "fail idempotency_key and merchant_id",
			cmd: &CreatePayment{
				Amount:         100,
				Currency:       "USD",
				IdempotencyKey: "",
				MerchantID:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := NewPayment(tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, PaymentStatusNew, payment.Status)
			}
		})
	}
}

func TestTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		from    PaymentStatus
		to      PaymentStatus
		wantErr bool
	}{
		{"new to processing", PaymentStatusNew, PaymentStatusProcessing, false},
		{"new to completed", PaymentStatusNew, PaymentStatusCompleted, true},
		{"processing to completed", PaymentStatusProcessing, PaymentStatusCompleted, false},
		{"processing to failed", PaymentStatusProcessing, PaymentStatusFailed, false},
		{"completed to failed", PaymentStatusCompleted, PaymentStatusFailed, true},
		{"cancelled to processing", PaymentStatusCancelled, PaymentStatusProcessing, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.from}
			err := payment.TransitionTo(tt.to)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.from, payment.Status) // статус не изменился
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.to, payment.Status)
			}
		})
	}
}
