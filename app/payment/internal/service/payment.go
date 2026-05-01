package service

import (
	"context"

	v1 "gomall/api/payment/v1"
	"gomall/app/payment/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	kratosJWT "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/google/uuid"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type PaymentService struct {
	v1.UnimplementedPaymentServiceServer
	uc *biz.PaymentUsecase
}

func NewPaymentService(uc *biz.PaymentUsecase) *PaymentService {
	return &PaymentService{uc: uc}
}

// callerID extracts the JWT subject from the request context (empty string if unavailable).
func callerID(ctx context.Context) string {
	claims, ok := kratosJWT.FromContext(ctx)
	if !ok {
		return ""
	}
	mc, ok := claims.(jwtv5.MapClaims)
	if !ok {
		return ""
	}
	sub, _ := mc.GetSubject()
	return sub
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
	// JWT subject overrides request body to prevent forging ledger entries for other users.
	userID := req.UserId
	if caller := callerID(ctx); caller != "" {
		userID = caller
	}
	p := &biz.Payment{
		OrderID:     req.OrderId,
		UserID:      userID,
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
	if caller := callerID(ctx); caller != "" && caller != result.UserID {
		return nil, errors.Forbidden("FORBIDDEN", "forbidden")
	}
	return bizToPayment(result), nil
}

func (s *PaymentService) ListPayments(ctx context.Context, req *v1.ListPaymentsRequest) (*v1.ListPaymentsResponse, error) {
	// Callers may only list their own payments.
	if caller := callerID(ctx); caller != "" && req.UserId != "" && caller != req.UserId {
		return nil, errors.Forbidden("FORBIDDEN", "forbidden")
	}
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
	p, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if caller := callerID(ctx); caller != "" && caller != p.UserID {
		return nil, errors.Forbidden("FORBIDDEN", "forbidden")
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
