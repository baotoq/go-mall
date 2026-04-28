package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL   string
	authToken string
	http      *http.Client
}

func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

func New(daprHTTPPort string, appID string) *Client {
	if daprHTTPPort == "" {
		daprHTTPPort = os.Getenv("DAPR_HTTP_PORT")
	}
	if daprHTTPPort == "" {
		daprHTTPPort = "3500"
	}
	return &Client{
		baseURL: fmt.Sprintf("http://localhost:%s/v1.0/invoke/%s/method", daprHTTPPort, appID),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type ProductInfo struct {
	Id             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Description    string  `json:"description"`
	ImageUrl       string  `json:"imageUrl"`
	Price          float64 `json:"price"`
	TotalStock     int64   `json:"totalStock"`
	RemainingStock int64   `json:"remainingStock"`
	CategoryId     string  `json:"categoryId"`
}

type ReservationItemInput struct {
	ProductId string `json:"productId"`
	Quantity  int64  `json:"quantity"`
}

type ReservationInfo struct {
	Id        string                 `json:"id"`
	SessionId string                 `json:"sessionId"`
	Status    string                 `json:"status"`
	Items     []ReservationItemInput `json:"items"`
	CreatedAt int64                  `json:"createdAt"`
	UpdatedAt int64                  `json:"updatedAt"`
}

type CreateReservationRequest struct {
	SessionId string                 `json:"sessionId"`
	Items     []ReservationItemInput `json:"items"`
}

type ReservationActionResponse struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

func (c *Client) GetProduct(ctx context.Context, id string) (*ProductInfo, error) {
	var out ProductInfo
	if err := c.do(ctx, http.MethodGet, "/api/v1/products/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateReservation(ctx context.Context, req CreateReservationRequest) (*ReservationInfo, error) {
	var out ReservationInfo
	if err := c.do(ctx, http.MethodPost, "/api/v1/reservations", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ConfirmReservation(ctx context.Context, id string) (*ReservationActionResponse, error) {
	var out ReservationActionResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/reservations/"+id+"/confirm", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CancelReservation(ctx context.Context, id string) (*ReservationActionResponse, error) {
	var out ReservationActionResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/reservations/"+id+"/cancel", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("catalog %s %s: status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
