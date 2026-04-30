package data

import (
	"context"

	"gomall/app/cart/internal/biz"
	"gomall/app/cart/internal/data/ent"
	entcart "gomall/app/cart/internal/data/ent/cart"
	entcartitem "gomall/app/cart/internal/data/ent/cartitem"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type cartRepo struct {
	data *Data
	log  *log.Helper
}

func NewCartRepo(data *Data, logger log.Logger) biz.CartRepo {
	return &cartRepo{data: data, log: log.NewHelper(logger)}
}

func (r *cartRepo) FindOrCreateBySession(ctx context.Context, sessionID string) (*biz.Cart, error) {
	c, err := r.data.db.Cart.Query().
		Where(entcart.SessionID(sessionID)).
		WithItems().
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		c, err = r.data.db.Cart.Create().
			SetSessionID(sessionID).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		c.Edges.Items = []*ent.CartItem{}
	}
	return entToCart(c), nil
}

func (r *cartRepo) AddItem(ctx context.Context, sessionID string, item *biz.CartItem) (*biz.Cart, error) {
	c, err := r.data.db.Cart.Query().
		Where(entcart.SessionID(sessionID)).
		WithItems().
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		c, err = r.data.db.Cart.Create().
			SetSessionID(sessionID).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		c.Edges.Items = []*ent.CartItem{}
	}

	existing, err := r.data.db.CartItem.Query().
		Where(
			entcartitem.CartID(c.ID),
			entcartitem.ProductID(item.ProductID),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if existing != nil {
		_, err = r.data.db.CartItem.UpdateOne(existing).
			SetQuantity(existing.Quantity + item.Quantity).
			Save(ctx)
	} else {
		q := r.data.db.CartItem.Create().
			SetCartID(c.ID).
			SetProductID(item.ProductID).
			SetName(item.Name).
			SetPriceCents(item.PriceCents).
			SetQuantity(item.Quantity)
		if item.Currency != "" {
			q = q.SetCurrency(item.Currency)
		}
		if item.ImageURL != "" {
			q = q.SetImageURL(item.ImageURL)
		}
		_, err = q.Save(ctx)
	}
	if err != nil {
		return nil, err
	}

	return r.FindOrCreateBySession(ctx, sessionID)
}

func (r *cartRepo) UpdateItem(ctx context.Context, sessionID string, productID uuid.UUID, quantity int) (*biz.Cart, error) {
	c, err := r.data.db.Cart.Query().
		Where(entcart.SessionID(sessionID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrCartNotFound
		}
		return nil, err
	}

	item, err := r.data.db.CartItem.Query().
		Where(
			entcartitem.CartID(c.ID),
			entcartitem.ProductID(productID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrItemNotFound
		}
		return nil, err
	}

	if quantity <= 0 {
		err = r.data.db.CartItem.DeleteOne(item).Exec(ctx)
	} else {
		_, err = r.data.db.CartItem.UpdateOne(item).SetQuantity(quantity).Save(ctx)
	}
	if err != nil {
		return nil, err
	}

	return r.FindOrCreateBySession(ctx, sessionID)
}

func (r *cartRepo) RemoveItem(ctx context.Context, sessionID string, productID uuid.UUID) (*biz.Cart, error) {
	c, err := r.data.db.Cart.Query().
		Where(entcart.SessionID(sessionID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrCartNotFound
		}
		return nil, err
	}

	item, err := r.data.db.CartItem.Query().
		Where(
			entcartitem.CartID(c.ID),
			entcartitem.ProductID(productID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrItemNotFound
		}
		return nil, err
	}

	if err := r.data.db.CartItem.DeleteOne(item).Exec(ctx); err != nil {
		return nil, err
	}

	return r.FindOrCreateBySession(ctx, sessionID)
}

func (r *cartRepo) Clear(ctx context.Context, sessionID string) error {
	c, err := r.data.db.Cart.Query().
		Where(entcart.SessionID(sessionID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}

	_, err = r.data.db.CartItem.Delete().
		Where(entcartitem.CartID(c.ID)).
		Exec(ctx)
	return err
}

func entToCart(c *ent.Cart) *biz.Cart {
	cart := &biz.Cart{
		ID:        c.ID,
		SessionID: c.SessionID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	var total int64
	for _, ci := range c.Edges.Items {
		item := entToCartItem(ci)
		cart.Items = append(cart.Items, item)
		total += item.SubtotalCents
	}
	cart.TotalCents = total
	return cart
}

func entToCartItem(ci *ent.CartItem) *biz.CartItem {
	item := &biz.CartItem{
		ID:        ci.ID,
		CartID:    ci.CartID,
		ProductID: ci.ProductID,
		Name:      ci.Name,
		PriceCents: ci.PriceCents,
		Currency:  ci.Currency,
		Quantity:  ci.Quantity,
		CreatedAt: ci.CreatedAt,
		UpdatedAt: ci.UpdatedAt,
	}
	item.SubtotalCents = item.PriceCents * int64(item.Quantity)
	if ci.ImageURL != nil {
		item.ImageURL = *ci.ImageURL
	}
	return item
}
