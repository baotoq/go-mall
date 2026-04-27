// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package payment

import (
	"context"
	"fmt"

	"payment/ent/payment"
	"payment/internal/svc"
	"payment/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List payments by idempotency key
func NewGetPaymentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentsLogic {
	return &GetPaymentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaymentsLogic) GetPayments(req *types.GetPaymentsRequest) (resp *types.GetPaymentsResponse, err error) {
	if req.IdempotencyKey == "" {
		return &types.GetPaymentsResponse{Payments: []types.PaymentResponse{}}, nil
	}

	payments, err := l.svcCtx.Db.Payment.Query().
		Where(payment.IdempotencyKeyEQ(req.IdempotencyKey)).
		All(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	resp = &types.GetPaymentsResponse{
		Payments: make([]types.PaymentResponse, 0, len(payments)),
	}

	for _, p := range payments {
		resp.Payments = append(resp.Payments, types.PaymentResponse{
			Id:          p.ID.String(),
			TotalAmount: p.TotalAmount,
			Currency:    p.Currency,
			Status:      p.Status,
			CreatedAt:   p.CreatedAt.Unix(),
		})
	}

	return resp, nil
}
