package domain

type CreatePayment struct {
	Amount         int64
	Currency       Currency
	IdempotencyKey string
	MerchantID     string
}

func (c *CreatePayment) Validate() error {
	if c.Amount <= 0 {
		return &ValidationError{Message: "amount must be greater than zero"}
	}
	if c.IdempotencyKey == "" {
		return &ValidationError{Message: "idempotency key must be set"}
	}
	if err := c.Currency.isValid(); err != nil {
		return &ValidationError{Message: err.Error()}
	}
	if c.MerchantID == "" {
		return &ValidationError{Message: "merchant id must be set"}
	}
	return nil
}
