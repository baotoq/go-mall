package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	v1 "gomall/api/order/v1"
	"gomall/app/order/internal/biz"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

func (s *OrderService) Checkout(ctx context.Context, req *v1.CheckoutRequest) (*v1.CheckoutResponse, error) {
	if s.sagaCfg == nil || !s.sagaCfg.Enabled {
		return nil, kerrors.New(int(http.StatusNotImplemented), v1.ErrorReason_ER_UNSPECIFIED.String(), "checkout via saga is disabled")
	}

	if _, err := uuid.Parse(req.IdempotencyKey); err != nil {
		return nil, kerrors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "idempotency_key must be a valid UUID")
	}

	if len(req.Items) == 0 {
		return nil, kerrors.BadRequest(v1.ErrorReason_ORDER_EMPTY_ITEMS.String(), "items required")
	}

	items := make([]biz.CheckoutItem, 0, len(req.Items))
	var total int64
	for _, it := range req.Items {
		if it.Quantity <= 0 {
			return nil, kerrors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "item quantity must be > 0")
		}
		if it.PriceCents <= 0 {
			return nil, kerrors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "item price_cents must be > 0")
		}
		total += int64(it.Quantity) * it.PriceCents
		items = append(items, biz.CheckoutItem{
			ProductID:  it.ProductId,
			Quantity:   it.Quantity,
			PriceCents: it.PriceCents,
		})
	}

	checkoutID, orderID, err := s.checkout.Schedule(ctx, biz.CheckoutInput{
		IdempotencyKey: req.IdempotencyKey,
		UserID:         req.UserId,
		SessionID:      req.SessionId,
		Currency:       req.Currency,
		Items:          items,
		TotalCents:     total,
	})
	if err != nil {
		if errors.Is(err, biz.ErrCheckoutDuplicateKey) {
			return nil, kerrors.Conflict(v1.ErrorReason_CHECKOUT_DUPLICATE_KEY.String(), err.Error())
		}
		return nil, err
	}

	return &v1.CheckoutResponse{CheckoutId: checkoutID, OrderId: orderID}, nil
}

func (s *OrderService) GetCheckoutStatus(ctx context.Context, req *v1.GetCheckoutStatusRequest) (*v1.GetCheckoutStatusResponse, error) {
	if req.CheckoutId == "" {
		return nil, kerrors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "checkout_id required")
	}

	res, err := s.checkout.Status(ctx, req.CheckoutId)
	if err != nil {
		if isNotFound(err) {
			return nil, kerrors.NotFound(v1.ErrorReason_CHECKOUT_NOT_FOUND.String(), err.Error())
		}
		return nil, err
	}

	return &v1.GetCheckoutStatusResponse{
		State:     res.State,
		OrderId:   res.OrderID,
		PaymentId: res.PaymentID,
		Attempts:  0,
		Error:     res.LastError,
	}, nil
}

// isNotFound returns true when err looks like a workflow-instance-not-found error.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "ErrInstanceNotFound")
}
