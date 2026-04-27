// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package payment

import (
	"context"
	"fmt"

	"payment/internal/svc"
	"payment/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get payment status
func NewGetPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentLogic {
	return &GetPaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaymentLogic) GetPayment(req *types.GetPaymentRequest) (resp *types.PaymentResponse, err error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid payment id format: %w", err)
	}

	p, err := l.svcCtx.Db.Payment.Get(l.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &types.PaymentResponse{
		Id:          p.ID.String(),
		TotalAmount: p.TotalAmount,
		Currency:    p.Currency,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt.Unix(),
	}, nil
}
