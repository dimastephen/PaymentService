package storage

import "github.com/payment-service/shared/domain"

func fromDomain(payment *domain.Payment) *PaymentsModel {
	paymentsModel := &PaymentsModel{
		Id:               payment.Id,
		Status:           string(payment.Status),
		Amount:           payment.Amount,
		Currency:         string(payment.Currency),
		MerchantId:       payment.MerchantID,
		IdempotencyKey:   payment.IdempotencyKey,
		PspTransactionId: payment.PSPTransactionID,
		CreatedAt:        payment.CreatedAt,
		UpdatedAt:        payment.UpdatedAt,
		Error:            payment.ErrorMessage,
	}
	return paymentsModel
}

func toDomain(model *PaymentsModel) *domain.Payment {
	payment := &domain.Payment{
		Id:               model.Id,
		Status:           domain.PaymentStatus(model.Status),
		Amount:           model.Amount,
		Currency:         domain.Currency(model.Currency),
		MerchantID:       model.MerchantId,
		IdempotencyKey:   model.IdempotencyKey,
		PSPTransactionID: model.PspTransactionId,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		ErrorMessage:     model.Error,
	}
	return payment
}
