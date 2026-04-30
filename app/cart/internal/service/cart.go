package service

import (
	"context"

	v1 "gomall/api/cart/v1"
	"gomall/app/cart/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CartService struct {
	v1.UnimplementedCartServiceServer
	uc *biz.CartUsecase
}

func NewCartService(uc *biz.CartUsecase) *CartService {
	return &CartService{uc: uc}
}

func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.Cart, error) {
	if req.SessionId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "session_id is required")
	}
	cart, err := s.uc.GetOrCreate(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return bizToCart(cart), nil
}

func (s *CartService) AddItem(ctx context.Context, req *v1.AddItemRequest) (*v1.Cart, error) {
	if req.SessionId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "session_id is required")
	}
	if req.ProductId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "product_id is required")
	}
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid product_id")
	}
	if req.Quantity <= 0 {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "quantity must be > 0")
	}
	item := &biz.CartItem{
		ProductID:  productID,
		Name:       req.Name,
		PriceCents: req.PriceCents,
		Currency:   req.Currency,
		ImageURL:   req.ImageUrl,
		Quantity:   int(req.Quantity),
	}
	cart, err := s.uc.AddItem(ctx, req.SessionId, item)
	if err != nil {
		return nil, err
	}
	return bizToCart(cart), nil
}

func (s *CartService) UpdateItem(ctx context.Context, req *v1.UpdateItemRequest) (*v1.Cart, error) {
	if req.SessionId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "session_id is required")
	}
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid product_id")
	}
	cart, err := s.uc.UpdateItem(ctx, req.SessionId, productID, int(req.Quantity))
	if err != nil {
		return nil, err
	}
	return bizToCart(cart), nil
}

func (s *CartService) RemoveItem(ctx context.Context, req *v1.RemoveItemRequest) (*v1.Cart, error) {
	if req.SessionId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "session_id is required")
	}
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid product_id")
	}
	cart, err := s.uc.RemoveItem(ctx, req.SessionId, productID)
	if err != nil {
		return nil, err
	}
	return bizToCart(cart), nil
}

func (s *CartService) ClearCart(ctx context.Context, req *v1.ClearCartRequest) (*emptypb.Empty, error) {
	if req.SessionId == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "session_id is required")
	}
	if err := s.uc.Clear(ctx, req.SessionId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func bizToCart(c *biz.Cart) *v1.Cart {
	pb := &v1.Cart{
		Id:         c.ID.String(),
		SessionId:  c.SessionID,
		TotalCents: c.TotalCents,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	for _, item := range c.Items {
		pb.Items = append(pb.Items, bizToCartItem(item))
	}
	return pb
}

func bizToCartItem(ci *biz.CartItem) *v1.CartItem {
	return &v1.CartItem{
		Id:            ci.ID.String(),
		ProductId:     ci.ProductID.String(),
		Name:          ci.Name,
		PriceCents:    ci.PriceCents,
		Currency:      ci.Currency,
		ImageUrl:      ci.ImageURL,
		Quantity:      int32(ci.Quantity),
		SubtotalCents: ci.SubtotalCents,
	}
}
