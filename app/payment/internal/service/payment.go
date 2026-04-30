package service

import (
	"context"

	v1 "gomall/api/payment/v1"
	"gomall/app/payment/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

type PaymentService struct {
	v1.UnimplementedPaymentServiceServer
	uc *biz.PaymentUsecase
}

func NewPaymentService(uc *biz.PaymentUsecase) *PaymentService {
	return &PaymentService{uc: uc}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *v1.CreatePaymentRequest) (*v1.Payment, error) {
	if req.OrderId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "order_id is required")
	}
	if req.UserId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "user_id is required")
	}
	if req.AmountCents <= 0 {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "amount_cents must be > 0")
	}
	if req.Provider == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "provider is required")
	}
	p := &biz.Payment{
		OrderID:     req.OrderId,
		UserID:      req.UserId,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Provider:    req.Provider,
	}
	result, err := s.uc.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return bizToPayment(result), nil
}

func (s *PaymentService) GetPayment(ctx context.Context, req *v1.GetPaymentRequest) (*v1.Payment, error) {
	if req.Id == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "id is required")
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	result, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return bizToPayment(result), nil
}

func (s *PaymentService) ListPayments(ctx context.Context, req *v1.ListPaymentsRequest) (*v1.ListPaymentsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)
	payments, total, err := s.uc.ListPayments(ctx, req.UserId, req.OrderId, page, pageSize)
	if err != nil {
		return nil, err
	}
	resp := &v1.ListPaymentsResponse{
		Total: int32(total),
	}
	for _, p := range payments {
		resp.Payments = append(resp.Payments, bizToPayment(p))
	}
	return resp, nil
}

func (s *PaymentService) RefundPayment(ctx context.Context, req *v1.RefundPaymentRequest) (*v1.Payment, error) {
	if req.Id == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "id is required")
	}
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	result, err := s.uc.Refund(ctx, id)
	if err != nil {
		return nil, err
	}
	return bizToPayment(result), nil
}

func bizToPayment(p *biz.Payment) *v1.Payment {
	statusVal, ok := v1.PaymentStatus_value[p.Status]
	if !ok {
		statusVal = int32(v1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED)
	}
	return &v1.Payment{
		Id:          p.ID.String(),
		OrderId:     p.OrderID,
		UserId:      p.UserID,
		AmountCents: p.AmountCents,
		Currency:    p.Currency,
		Status:      v1.PaymentStatus(statusVal),
		Provider:    p.Provider,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
