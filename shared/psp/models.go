package psp

type TransactionRequest struct {
	Amount       int64
	CurrencyCode string
}

type TransactionResponse struct {
	PspTransactionID string
	Status           string
	ErrorMessage     string
}
