package handler

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	pb "github.com/payment-service/shared/proto/gen/proto/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name           string
		created        *pb.PaymentCreatedEvent
		completed      *pb.PaymentCompletedEvent
		failed         *pb.PaymentFailedEvent
		repoErrCreated error
		repoErrUpdated error
		wantErrCreated bool
		wantErrUpdated bool
	}{
		{
			name: "success",
			created: &pb.PaymentCreatedEvent{
				PaymentId:      "21cb3b13-4fff-4d0e-91c3-03c7a6b05069",
				Amount:         gofakeit.Int64(),
				Currency:       "USD",
				IdempotencyKey: "test",
				MerchantId:     "shop",
				OccuredAt:      time.Now().Unix(),
			},
			completed: &pb.PaymentCompletedEvent{
				PaymentId:        "21cb3b13-4fff-4d0e-91c3-03c7a6b05069",
				PspTransactionId: gofakeit.UUID(),
				OccurredAt:       time.Now().Unix(),
			},
			repoErrCreated: nil,
			repoErrUpdated: nil,
			wantErrCreated: false,
			wantErrUpdated: false,
		},
		{
			name: "fake psp id",
			created: &pb.PaymentCreatedEvent{
				PaymentId:      "21cb3b13-4fff-4d0e-91c3-03c7a6b05069",
				Amount:         gofakeit.Int64(),
				Currency:       "USD",
				IdempotencyKey: "test",
				MerchantId:     "shop",
				OccuredAt:      time.Now().Unix(),
			},
			completed: &pb.PaymentCompletedEvent{
				PaymentId:        "21cb4b13-4fff-4d0e-91c3-03c7a6b05069",
				PspTransactionId: gofakeit.UUID(),
				OccurredAt:       time.Now().Unix(),
			},
			repoErrCreated: nil,
			repoErrUpdated: pgx.ErrNoRows,
			wantErrCreated: false,
			wantErrUpdated: true,
		},
		{
			name: "invalid amount and currency",
			created: &pb.PaymentCreatedEvent{
				PaymentId:      "21cb3b13-4fff-4d0e-91c3-03c7a6b05069",
				Amount:         0,
				Currency:       "smth",
				IdempotencyKey: "test",
				MerchantId:     "shop",
				OccuredAt:      time.Now().Unix(),
			},
			completed: &pb.PaymentCompletedEvent{
				PaymentId:        "21cb3b13-4fff-4d0e-91c3-03c7a6b05069",
				PspTransactionId: gofakeit.UUID(),
				OccurredAt:       time.Now().Unix(),
			},
			repoErrCreated: nil,
			repoErrUpdated: nil,
			wantErrCreated: true,
			wantErrUpdated: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockPaymentRepo)
			mockRepo.On("Create", mock.Anything, mock.Anything).Return(tt.repoErrCreated)
			mockRepo.On("Update", mock.Anything, mock.Anything).Return(tt.repoErrUpdated)
			projection := NewProjectionHandler(mockRepo)
			ctx := context.Background()
			createdVal, err := proto.Marshal(tt.created)
			updatedVal, err := proto.Marshal(tt.completed)
			err = projection.Handle(ctx, []byte("test"), createdVal, map[string]string{"event_type": "payment_created"})
			if tt.wantErrCreated {
				assert.Error(t, err)
				return
			}
			err = projection.Handle(ctx, []byte("test"), updatedVal, map[string]string{"event_type": "COMPLETED"})
			if tt.wantErrUpdated {
				assert.Error(t, err)
			}
		})
	}
}
