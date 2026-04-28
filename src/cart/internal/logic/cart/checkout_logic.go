package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cart/ent"
	"cart/ent/cartitem"
	entschema "cart/ent/schema"
	catalogclient "cart/internal/clients/catalog"
	paymentclient "cart/internal/clients/payment"
	"cart/internal/svc"
	"cart/internal/types"

	"github.com/google/uuid"
	sharedevent "shared/event"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckoutLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	sessionID string
}

func NewCheckoutLogic(ctx context.Context, svcCtx *svc.ServiceContext, sessionID string) *CheckoutLogic {
	return &CheckoutLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		sessionID: sessionID,
	}
}

type OrderPlacedEvent struct {
	OccurredAt    time.Time `json:"occurred_at"`
	OrderID       string    `json:"order_id"`
	SessionID     string    `json:"session_id"`
	TotalAmount   int64     `json:"total_amount"`
	ReservationID string    `json:"reservation_id"`
	PaymentID     string    `json:"payment_id"`
	TransactionID string    `json:"transaction_id"`
}

func (e OrderPlacedEvent) EventID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

type OrderFailedEvent struct {
	OccurredAt    time.Time `json:"occurred_at"`
	OrderID       string    `json:"order_id"`
	SessionID     string    `json:"session_id"`
	TotalAmount   int64     `json:"total_amount"`
	FailureReason string    `json:"failure_reason"`
}

func (e OrderFailedEvent) EventID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func (l *CheckoutLogic) Checkout(req *types.CheckoutRequest) (*types.CheckoutResponse, error) {
	if l.sessionID == "" {
		return nil, errors.New("missing session")
	}

	// 1. Load cart items.
	cartItems, err := l.svcCtx.Db.CartItem.Query().
		Where(cartitem.SessionID(l.sessionID)).
		All(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("query cart: %w", err)
	}
	if len(cartItems) == 0 {
		return nil, errors.New("cart is empty")
	}

	// 2. Fetch product details from catalog; build snapshot + reservation request.
	orderItems := make([]entschema.OrderItem, 0, len(cartItems))
	respItems := make([]types.OrderItemInfo, 0, len(cartItems))
	resvItems := make([]catalogclient.ReservationItemInput, 0, len(cartItems))
	var totalAmount int64

	for _, ci := range cartItems {
		productID := ci.ProductID.String()
		product, err := l.svcCtx.CatalogClient.GetProduct(l.ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("fetch product %s: %w", productID, err)
		}
		unitPrice := int64(product.Price)
		orderItems = append(orderItems, entschema.OrderItem{
			ProductID: productID,
			Quantity:  ci.Quantity,
			UnitPrice: unitPrice,
		})
		respItems = append(respItems, types.OrderItemInfo{
			ProductId: productID,
			Quantity:  ci.Quantity,
			UnitPrice: unitPrice,
		})
		resvItems = append(resvItems, catalogclient.ReservationItemInput{
			ProductId: productID,
			Quantity:  ci.Quantity,
		})
		totalAmount += unitPrice * ci.Quantity
	}

	// 3. Create pending order (so we have a stable id for idempotency key).
	order, err := l.svcCtx.Db.Order.Create().
		SetSessionID(l.sessionID).
		SetStatus("pending").
		SetTotalAmount(totalAmount).
		SetItems(orderItems).
		Save(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	orderID := order.ID.String()

	// 4. Reserve stock in catalog.
	reservation, err := l.svcCtx.CatalogClient.CreateReservation(l.ctx, catalogclient.CreateReservationRequest{
		SessionId: l.sessionID,
		Items:     resvItems,
	})
	if err != nil {
		l.failOrder(order, "reservation failed: "+err.Error(), totalAmount)
		return nil, fmt.Errorf("reserve stock: %w", err)
	}

	order, err = order.Update().SetReservationID(reservation.Id).Save(l.ctx)
	if err != nil {
		// Best-effort cancel reservation, then fail.
		_, _ = l.svcCtx.CatalogClient.CancelReservation(l.ctx, reservation.Id)
		l.failOrder(order, "persist reservation: "+err.Error(), totalAmount)
		return nil, fmt.Errorf("persist reservation id: %w", err)
	}

	// 5. Charge payment with idempotency key derived from order id.
	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = "order-" + orderID
	}

	payItems := make([]paymentclient.LineItem, 0, len(orderItems))
	for _, it := range orderItems {
		payItems = append(payItems, paymentclient.LineItem{
			ProductId: it.ProductID,
			Quantity:  it.Quantity,
			Price:     float64(it.UnitPrice),
		})
	}

	payment, err := l.svcCtx.PaymentClient.CreatePayment(l.ctx, idempotencyKey, paymentclient.CreatePaymentRequest{
		TotalAmount: float64(totalAmount),
		Currency:    "USD",
		Items:       payItems,
	})
	if err != nil {
		_, cancelErr := l.svcCtx.CatalogClient.CancelReservation(l.ctx, reservation.Id)
		if cancelErr != nil {
			l.Logger.Errorw("cancel reservation after payment error",
				logx.Field("reservation_id", reservation.Id),
				logx.Field("error", cancelErr))
		}
		l.failOrder(order, "payment call failed: "+err.Error(), totalAmount)
		return nil, fmt.Errorf("payment: %w", err)
	}

	// 6. Branch on payment status.
	if payment.Status != "succeeded" {
		_, cancelErr := l.svcCtx.CatalogClient.CancelReservation(l.ctx, reservation.Id)
		if cancelErr != nil {
			l.Logger.Errorw("cancel reservation after payment failure",
				logx.Field("reservation_id", reservation.Id),
				logx.Field("error", cancelErr))
		}
		reason := "payment status: " + payment.Status
		updated, updErr := order.Update().
			SetStatus("failed").
			SetPaymentID(payment.Id).
			SetFailureReason(reason).
			Save(l.ctx)
		if updErr != nil {
			l.Logger.Errorw("persist failed order", logx.Field("error", updErr))
			updated = order
		}
		l.publishFailed(updated, reason, totalAmount)
		return l.toResponse(updated, respItems), nil
	}

	// 7. Confirm reservation.
	if _, err := l.svcCtx.CatalogClient.ConfirmReservation(l.ctx, reservation.Id); err != nil {
		// Payment captured but confirmation failed — surface as failed for visibility.
		reason := "confirm reservation failed: " + err.Error()
		updated, updErr := order.Update().
			SetStatus("failed").
			SetPaymentID(payment.Id).
			SetFailureReason(reason).
			Save(l.ctx)
		if updErr != nil {
			l.Logger.Errorw("persist failed order after confirm error", logx.Field("error", updErr))
			updated = order
		}
		l.publishFailed(updated, reason, totalAmount)
		return l.toResponse(updated, respItems), nil
	}

	// 8. Mark order placed and clear cart.
	updated, err := order.Update().
		SetStatus("placed").
		SetPaymentID(payment.Id).
		Save(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("persist placed order: %w", err)
	}

	if _, err := l.svcCtx.Db.CartItem.Delete().
		Where(cartitem.SessionID(l.sessionID)).
		Exec(l.ctx); err != nil {
		l.Logger.Errorw("clear cart after checkout",
			logx.Field("session_id", l.sessionID),
			logx.Field("error", err))
	}

	evt := OrderPlacedEvent{
		OccurredAt:    time.Now().UTC(),
		OrderID:       updated.ID.String(),
		SessionID:     l.sessionID,
		TotalAmount:   updated.TotalAmount,
		ReservationID: reservation.Id,
		PaymentID:     payment.Id,
	}
	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, sharedevent.Event(evt)); err != nil {
		l.Logger.Errorw("dispatch order.placed", logx.Field("error", err))
	}

	return l.toResponse(updated, respItems), nil
}

func (l *CheckoutLogic) failOrder(order *ent.Order, reason string, totalAmount int64) {
	updated, err := order.Update().
		SetStatus("failed").
		SetFailureReason(reason).
		Save(l.ctx)
	if err != nil {
		l.Logger.Errorw("persist failed order",
			logx.Field("order_id", order.ID),
			logx.Field("error", err))
		updated = order
	}
	l.publishFailed(updated, reason, totalAmount)
}

func (l *CheckoutLogic) publishFailed(order *ent.Order, reason string, totalAmount int64) {
	evt := OrderFailedEvent{
		OccurredAt:    time.Now().UTC(),
		OrderID:       order.ID.String(),
		SessionID:     l.sessionID,
		TotalAmount:   totalAmount,
		FailureReason: reason,
	}
	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, sharedevent.Event(evt)); err != nil {
		l.Logger.Errorw("dispatch order.failed", logx.Field("error", err))
	}
}

func (l *CheckoutLogic) toResponse(order *ent.Order, items []types.OrderItemInfo) *types.CheckoutResponse {
	return &types.CheckoutResponse{
		OrderId:       order.ID.String(),
		Status:        order.Status,
		TotalAmount:   order.TotalAmount,
		ReservationId: order.ReservationID,
		PaymentId:     order.PaymentID,
		TransactionId: order.TransactionID,
		FailureReason: order.FailureReason,
		Items:         items,
	}
}
