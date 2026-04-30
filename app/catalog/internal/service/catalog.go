package service

import (
	"context"

	v1 "gomall/api/catalog/v1"
	"gomall/app/catalog/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CatalogService struct {
	v1.UnimplementedCatalogServiceServer
	puc *biz.ProductUsecase
	cuc *biz.CategoryUsecase
	ruc *biz.ReservationUsecase
}

func NewCatalogService(puc *biz.ProductUsecase, cuc *biz.CategoryUsecase, ruc *biz.ReservationUsecase) *CatalogService {
	return &CatalogService{puc: puc, cuc: cuc, ruc: ruc}
}

var validSorts = map[string]bool{"": true, "price_asc": true, "price_desc": true, "created_desc": true}

func (s *CatalogService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	if req.PageSize > 100 {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "page_size max is 100")
	}
	if !validSorts[req.Sort] {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid sort value")
	}

	filter := biz.ListProductsFilter{
		Q:        req.Q,
		Sort:     req.Sort,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	if req.CategoryId != "" {
		id, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid category_id")
		}
		filter.CategoryID = &id
	}
	if req.MinPrice > 0 {
		filter.MinPrice = &req.MinPrice
	}
	if req.MaxPrice > 0 {
		filter.MaxPrice = &req.MaxPrice
	}

	res, err := s.puc.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	prods := make([]*v1.Product, len(res.Products))
	for i, p := range res.Products {
		prods[i] = bizToProduct(p)
	}
	return &v1.ListProductsResponse{
		Products: prods,
		Page:     res.Page,
		PageSize: res.PageSize,
		Total:    res.Total,
	}, nil
}

func (s *CatalogService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.Product, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	p, err := s.puc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return bizToProduct(p), nil
}

func (s *CatalogService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.Product, error) {
	if req.Name == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "name is required")
	}
	p := &biz.Product{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		ImageURL:    req.ImageUrl,
		Theme:       req.Theme,
		Stock:       int(req.Stock),
	}
	if req.CategoryId != "" {
		id, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid category_id")
		}
		p.CategoryID = &id
	}
	created, err := s.puc.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return bizToProduct(created), nil
}

func (s *CatalogService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.Product, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	if req.Name == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "name is required")
	}
	p := &biz.Product{
		ID:          id,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		ImageURL:    req.ImageUrl,
		Theme:       req.Theme,
		Stock:       int(req.Stock),
	}
	if req.CategoryId != "" {
		catID, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid category_id")
		}
		p.CategoryID = &catID
	}
	updated, err := s.puc.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return bizToProduct(updated), nil
}

func (s *CatalogService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	if err := s.puc.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *CatalogService) ListCategories(ctx context.Context, req *v1.ListCategoriesRequest) (*v1.ListCategoriesResponse, error) {
	if req.PageSize > 100 {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "page_size max is 100")
	}
	res, err := s.cuc.List(ctx, biz.ListCategoriesFilter{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, err
	}
	cats := make([]*v1.Category, len(res.Categories))
	for i, c := range res.Categories {
		cats[i] = bizToCategory(c)
	}
	return &v1.ListCategoriesResponse{
		Categories: cats,
		Page:       res.Page,
		PageSize:   res.PageSize,
		Total:      res.Total,
	}, nil
}

func (s *CatalogService) GetCategory(ctx context.Context, req *v1.GetCategoryRequest) (*v1.Category, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	c, err := s.cuc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return bizToCategory(c), nil
}

func (s *CatalogService) CreateCategory(ctx context.Context, req *v1.CreateCategoryRequest) (*v1.Category, error) {
	if req.Name == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "name is required")
	}
	c, err := s.cuc.Create(ctx, &biz.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return bizToCategory(c), nil
}

func (s *CatalogService) UpdateCategory(ctx context.Context, req *v1.UpdateCategoryRequest) (*v1.Category, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	if req.Name == "" {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "name is required")
	}
	c, err := s.cuc.Update(ctx, &biz.Category{
		ID:          id,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return bizToCategory(c), nil
}

func (s *CatalogService) DeleteCategory(ctx context.Context, req *v1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid id")
	}
	if err := s.cuc.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func bizToProduct(p *biz.Product) *v1.Product {
	pb := &v1.Product{
		Id:          p.ID.String(),
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		Currency:    p.Currency,
		ImageUrl:    p.ImageURL,
		Theme:       p.Theme,
		Stock:       int32(p.Stock),
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.CategoryID != nil {
		pb.CategoryId = p.CategoryID.String()
	}
	return pb
}

func bizToCategory(c *biz.Category) *v1.Category {
	return &v1.Category{
		Id:          c.ID.String(),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *CatalogService) ReserveStock(ctx context.Context, req *v1.ReserveStockRequest) (*v1.ReserveStockResponse, error) {
	cartID, err := uuid.Parse(req.CartId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid cart_id")
	}
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid product_id")
	}
	if req.Quantity <= 0 {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "quantity must be positive")
	}
	r, err := s.ruc.Reserve(ctx, cartID, productID, int(req.Quantity))
	if err != nil {
		return nil, err
	}
	return &v1.ReserveStockResponse{Reservation: bizToReservation(r)}, nil
}

func (s *CatalogService) ReleaseStock(ctx context.Context, req *v1.ReleaseStockRequest) (*emptypb.Empty, error) {
	cartID, err := uuid.Parse(req.CartId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid cart_id")
	}
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid product_id")
	}
	if err := s.ruc.Release(ctx, cartID, productID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *CatalogService) AdjustReservation(ctx context.Context, req *v1.AdjustReservationRequest) (*v1.ReserveStockResponse, error) {
	cartID, err := uuid.Parse(req.CartId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid cart_id")
	}
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid product_id")
	}
	if req.Quantity <= 0 {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "quantity must be positive")
	}
	r, err := s.ruc.Adjust(ctx, cartID, productID, int(req.Quantity))
	if err != nil {
		return nil, err
	}
	return &v1.ReserveStockResponse{Reservation: bizToReservation(r)}, nil
}

func (s *CatalogService) CommitReservation(ctx context.Context, req *v1.CommitReservationRequest) (*emptypb.Empty, error) {
	cartID, err := uuid.Parse(req.CartId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid cart_id")
	}
	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid product_id")
	}
	if err := s.ruc.Commit(ctx, cartID, productID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *CatalogService) ReleaseAllReservations(ctx context.Context, req *v1.ReleaseAllReservationsRequest) (*emptypb.Empty, error) {
	cartID, err := uuid.Parse(req.CartId)
	if err != nil {
		return nil, errors.BadRequest(v1.ErrorReason_INVALID_ARGUMENT.String(), "invalid cart_id")
	}
	if err := s.ruc.ReleaseAll(ctx, cartID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func bizToReservation(r *biz.Reservation) *v1.Reservation {
	return &v1.Reservation{
		Id:        r.ID.String(),
		CartId:    r.CartID.String(),
		ProductId: r.ProductID.String(),
		Quantity:  int32(r.Quantity),
		Status:    string(r.Status),
		ExpiresAt: r.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
