package provider

import (
	"context"

	"github.com/google/uuid"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error) {
	if req.Amount <= 0 {
		return &ChargeResponse{
			Success:      false,
			ErrorMessage: "invalid amount",
		}, nil
	}

	if req.Amount > 1000000 {
		return &ChargeResponse{
			Success:      false,
			ErrorMessage: "amount exceeds limit",
		}, nil
	}

	txId, err := uuid.NewV7()
	if err != nil {
		txId = uuid.New()
	}

	return &ChargeResponse{
		Success:       true,
		TransactionId: "mock_" + txId.String(),
	}, nil
}
