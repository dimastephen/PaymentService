package handler

import (
	"context"
	"log/slog"
	"os"
	"testing"

	pb "github.com/payment-service/shared/proto/gen/proto/payment"
	"github.com/payment-service/shared/psp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestHandle(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)
	tests := []struct {
		name         string
		key          string
		cmd          *pb.CreatePaymentCommand
		isProccessed bool
		pspStatus    string
		pspError     error
		wantErr      bool
	}{
		{
			name: "success",
			key:  "payment_uuid",
			cmd: &pb.CreatePaymentCommand{
				Amount:         100,
				Currency:       "USD",
				IdempotencyKey: "payment_uuid",
				MerchantId:     "shop",
			},
			isProccessed: false,
			pspStatus:    psp.StatusSuccess,
			pspError:     nil,
			wantErr:      false,
		},
		{
			name: "pspFailure",
			key:  "payment_uuid",
			cmd: &pb.CreatePaymentCommand{
				Amount:         100,
				Currency:       "USD",
				IdempotencyKey: "payment_uuid",
				MerchantId:     "shop",
			},
			isProccessed: false,
			pspStatus:    psp.StatusFailed,
			pspError:     nil,
			wantErr:      false,
		},
		{
			name: "Unknown psp status",
			key:  "payment_uuid",
			cmd: &pb.CreatePaymentCommand{
				Amount:         100,
				Currency:       "USD",
				IdempotencyKey: "payment_uuid",
				MerchantId:     "shop",
			},
			isProccessed: false,
			pspStatus:    "test",
			pspError:     nil,
			wantErr:      true,
		},
		{
			name: "dublicate_idempotency_key",
			key:  "payment_uuid",
			cmd: &pb.CreatePaymentCommand{
				Amount:         100,
				Currency:       "USD",
				IdempotencyKey: "payment_uuid",
				MerchantId:     "shop",
			},
			isProccessed: true,
			pspStatus:    "success",
			pspError:     nil,
			wantErr:      true,
		},
		{
			name: "invalid_amount_and_currency",
			key:  "payment_uuid",
			cmd: &pb.CreatePaymentCommand{
				Amount:         0,
				Currency:       "",
				IdempotencyKey: "payment_uuid",
				MerchantId:     "shop",
			},
			isProccessed: false,
			pspStatus:    "success",
			pspError:     nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPSP := new(mockPSP)
			mockCommandRepo := new(mockCommandRepo)
			mockStore := new(mockStore)
			mockEventRepo := new(mockEventRepo)
			mockPaymentRepo := new(mockPaymentRepo)
			mockProducer := new(mockProducer)

			mockCommandRepo.On("IsProcessed", mock.Anything, mock.Anything).Return(tt.isProccessed, nil)
			if !tt.isProccessed {
				mockEventRepo.On("Add", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockProducer.On("SendMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				if tt.pspError != nil {
					mockPSP.On("Charge", mock.Anything, mock.Anything).Return((*psp.TransactionResponse)(nil), tt.pspError)
				} else if tt.pspStatus != "" {
					mockPSP.On("Charge", mock.Anything, mock.Anything).Return(&psp.TransactionResponse{
						PspTransactionID: "test",
						Status:           tt.pspStatus,
						ErrorMessage:     "",
					}, nil)
					mockStore.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
				}
			}

			handler := NewWorkerHandler(mockPSP, mockEventRepo, mockCommandRepo, mockPaymentRepo, mockStore, "payment_events")
			value, _ := proto.Marshal(tt.cmd)
			headers := make(map[string]string)
			err := handler.Handle(context.Background(), []byte(tt.key), value, headers)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
