// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package payment

import (
	"context"
	"fmt"
	"time"

	"payment/ent"
	"payment/ent/payment"
	"payment/internal/provider"
	"payment/internal/svc"
	"payment/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create a payment
func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type PaymentSucceededEvent struct {
	OccurredAt     time.Time `json:"occurred_at"`
	PaymentID      string    `json:"payment_id"`
	Amount         float64   `json:"amount"`
	IdempotencyKey string    `json:"idempotency_key"`
	TransactionID  string    `json:"transaction_id"`
}

func (e PaymentSucceededEvent) EventID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

type PaymentFailedEvent struct {
	OccurredAt     time.Time `json:"occurred_at"`
	PaymentID      string    `json:"payment_id"`
	Amount         float64   `json:"amount"`
	IdempotencyKey string    `json:"idempotency_key"`
	ErrorMessage   string    `json:"error_message"`
}

func (e PaymentFailedEvent) EventID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func (l *CreatePaymentLogic) CreatePayment(req *types.CreatePaymentRequest, idempotencyKey string) (resp *types.PaymentResponse, err error) {
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotency_key is required")
	}

	l.Logger.Infow("handling create payment", logx.Field("req", req), logx.Field("idempotency_key", idempotencyKey))

	p, err := l.svcCtx.Db.Payment.Create().
		SetIdempotencyKey(idempotencyKey).
		SetTotalAmount(req.TotalAmount).
		SetCurrency(req.Currency).
		SetStatus("pending").
		Save(l.ctx)

	if err != nil {
		if ent.IsConstraintError(err) {
			existing, qErr := l.svcCtx.Db.Payment.Query().
				Where(payment.IdempotencyKeyEQ(idempotencyKey)).
				First(l.ctx)
			if qErr != nil {
				return nil, fmt.Errorf("failed to query existing payment: %w", qErr)
			}
			l.Logger.Infow("idempotent payment replay", logx.Field("id", existing.ID))
			return &types.PaymentResponse{
				Id:          existing.ID.String(),
				TotalAmount: existing.TotalAmount,
				Currency:    existing.Currency,
				Status:      existing.Status,
				CreatedAt:   existing.CreatedAt.Unix(),
			}, nil
		}
		return nil, fmt.Errorf("failed to create pending payment: %w", err)
	}

	chargeResp, chargeErr := l.svcCtx.PaymentProvider.Charge(l.ctx, provider.ChargeRequest{
		Amount:   req.TotalAmount,
		Currency: req.Currency,
	})

	if chargeErr != nil {
		_, updateErr := l.svcCtx.Db.Payment.UpdateOneID(p.ID).SetStatus("failed").Save(l.ctx)
		if updateErr != nil {
			l.Logger.Errorw("failed to update payment status", logx.Field("error", updateErr))
		}
		return nil, fmt.Errorf("payment provider charge failed: %w", chargeErr)
	}

	status := "failed"
	if chargeResp.Success {
		status = "succeeded"
	}

	updatedP, err := l.svcCtx.Db.Payment.UpdateOneID(p.ID).
		SetStatus(status).
		Save(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	if chargeResp.Success {
		evt := PaymentSucceededEvent{
			OccurredAt:     time.Now().UTC(),
			PaymentID:      updatedP.ID.String(),
			Amount:         updatedP.TotalAmount,
			IdempotencyKey: updatedP.IdempotencyKey,
			TransactionID:  chargeResp.TransactionId,
		}
		if dispatchErr := l.svcCtx.Dispatcher.PublishEvent(l.ctx, evt); dispatchErr != nil {
			l.Logger.Errorw("failed to dispatch payment.succeeded", logx.Field("error", dispatchErr))
		}
	} else {
		evt := PaymentFailedEvent{
			OccurredAt:     time.Now().UTC(),
			PaymentID:      updatedP.ID.String(),
			Amount:         updatedP.TotalAmount,
			IdempotencyKey: updatedP.IdempotencyKey,
			ErrorMessage:   chargeResp.ErrorMessage,
		}
		if dispatchErr := l.svcCtx.Dispatcher.PublishEvent(l.ctx, evt); dispatchErr != nil {
			l.Logger.Errorw("failed to dispatch payment.failed", logx.Field("error", dispatchErr))
		}
	}

	return &types.PaymentResponse{
		Id:          updatedP.ID.String(),
		TotalAmount: updatedP.TotalAmount,
		Currency:    updatedP.Currency,
		Status:      updatedP.Status,
		CreatedAt:   updatedP.CreatedAt.Unix(),
	}, nil
}
