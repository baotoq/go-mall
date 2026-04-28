package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type LineItem struct {
	ProductId string  `json:"productId"`
	Quantity  int64   `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreatePaymentRequest struct {
	TotalAmount float64    `json:"totalAmount"`
	Currency    string     `json:"currency"`
	Items       []LineItem `json:"items"`
}

type PaymentResponse struct {
	Id          string  `json:"id"`
	TotalAmount float64 `json:"totalAmount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	CreatedAt   int64   `json:"createdAt"`
}

func (c *Client) CreatePayment(ctx context.Context, idempotencyKey string, req CreatePaymentRequest) (*PaymentResponse, error) {
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/payments", bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call payment: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("payment status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out PaymentResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode payment: %w", err)
	}
	return &out, nil
}
