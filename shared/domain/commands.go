package domain

type CreatePayment struct {
	Amount         int64
	Currency       Currency
	IdempotencyKey string
	MerchantID     string
}
