package storage

import "github.com/payment-service/shared/domain"

func fromDomain(payment *domain.Payment) *PaymentsModel {
	paymentsModel := &PaymentsModel{
		Id:             payment.Id,
		Status:         string(payment.Status),
		Amount:         payment.Amount,
		Currency:       string(payment.Currency),
		MerchantId:     payment.MerchantID,
		IdempotencyKey: payment.IdempotencyKey,
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
	}
	if payment.PSPTransactionID != "" {
		paymentsModel.PspTransactionId = &payment.PSPTransactionID
	}
	if payment.ErrorMessage != "" {
		paymentsModel.Error = &payment.ErrorMessage
	}
	return paymentsModel
}

func toDomain(model *PaymentsModel) *domain.Payment {
	payment := &domain.Payment{
		Id:             model.Id,
		Status:         domain.PaymentStatus(model.Status),
		Amount:         model.Amount,
		Currency:       domain.Currency(model.Currency),
		MerchantID:     model.MerchantId,
		IdempotencyKey: model.IdempotencyKey,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
	if model.PspTransactionId != nil {
		payment.PSPTransactionID = *model.PspTransactionId
	}
	if model.Error != nil {
		payment.ErrorMessage = *model.Error
	}
	return payment
}
