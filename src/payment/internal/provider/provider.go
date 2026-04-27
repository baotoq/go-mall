package provider

import "context"

type ChargeRequest struct {
	Amount   float64
	Currency string
}

type ChargeResponse struct {
	Success       bool
	TransactionId string
	ErrorMessage  string
}

type PaymentProvider interface {
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error)
}
