package service

import (
	"context"

	v1 "gomall/api/order/v1"
	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

type OrderService struct {
	v1.UnimplementedOrderServiceServer
	uc       *biz.OrderUsecase
	checkout *biz.CheckoutUsecase
	sagaCfg  *conf.Saga
}

func NewOrderService(uc *biz.OrderUsecase, checkout *biz.CheckoutUsecase, sagaCfg *conf.Saga) *OrderService {
	return &OrderService{uc: uc, checkout: checkout, sagaCfg: sagaCfg}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.Order, error) {
	if req.UserId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "user_id required")
	}
	if req.SessionId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "session_id required")
	}
	if len(req.Items) == 0 {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "items required")
	}
	items := make([]biz.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, biz.OrderItem{
			ProductID:  it.ProductId,
			Name:       it.Name,
			PriceCents: it.PriceCents,
			ImageURL:   it.ImageUrl,
			Quantity:   it.Quantity,
		})
	}
	o := &biz.Order{
		UserID:    req.UserId,
		SessionID: req.SessionId,
		Items:     items,
		Currency:  req.Currency,
	}
	res, err := s.uc.Create(ctx, o)
	if err != nil {
		return nil, err
	}
	return bizToOrder(res), nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.Order, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	res, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return bizToOrder(res), nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	var statusStr string
	if req.Status != v1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		statusStr = v1.OrderStatus_name[int32(req.Status)]
	}
	orders, total, err := s.uc.ListOrders(ctx, req.UserId, statusStr, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	resp := &v1.ListOrdersResponse{Total: int32(total)}
	for _, o := range orders {
		resp.Orders = append(resp.Orders, bizToOrder(o))
	}
	return resp, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *v1.UpdateOrderStatusRequest) (*v1.Order, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	statusStr := v1.OrderStatus_name[int32(req.Status)]
	res, err := s.uc.UpdateStatus(ctx, id, statusStr)
	if err != nil {
		return nil, err
	}
	return bizToOrder(res), nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.Order, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	res, err := s.uc.Cancel(ctx, id)
	if err != nil {
		return nil, err
	}
	return bizToOrder(res), nil
}

func bizToOrder(o *biz.Order) *v1.Order {
	statusVal, ok := v1.OrderStatus_value[o.Status]
	if !ok {
		statusVal = int32(v1.OrderStatus_ORDER_STATUS_UNSPECIFIED)
	}
	items := make([]*v1.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &v1.OrderItem{
			ProductId:     it.ProductID,
			Name:          it.Name,
			PriceCents:    it.PriceCents,
			Currency:      it.Currency,
			ImageUrl:      it.ImageURL,
			Quantity:      it.Quantity,
			SubtotalCents: it.SubtotalCents,
		})
	}
	return &v1.Order{
		Id:        o.ID.String(),
		UserId:    o.UserID,
		SessionId: o.SessionID,
		Items:     items,
		TotalCents: o.TotalCents,
		Currency:  o.Currency,
		Status:    v1.OrderStatus(statusVal),
		PaymentId: o.PaymentID,
		CreatedAt: o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
