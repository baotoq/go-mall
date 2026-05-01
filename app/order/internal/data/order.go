package data

import (
	"context"

	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/data/ent"
	entorder "gomall/app/order/internal/data/ent/order"
	"gomall/app/order/internal/data/ent/schema"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type orderRepo struct {
	data *Data
	log  *log.Helper
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{data: data, log: log.NewHelper(logger)}
}

func (r *orderRepo) Create(ctx context.Context, o *biz.Order) (*biz.Order, error) {
	q := r.data.db.Order.Create().
		SetUserID(o.UserID).
		SetSessionID(o.SessionID).
		SetItems(bizItemsToSchema(o.Items)).
		SetTotalCents(o.TotalCents).
		SetCurrency(o.Currency).
		SetStatus(o.Status)
	if o.PaymentID != "" {
		q = q.SetPaymentID(o.PaymentID)
	}
	result, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}
	return entToOrder(result), nil
}

func (r *orderRepo) CreateWithEvent(ctx context.Context, o *biz.Order, emit func(context.Context, biz.TxExecer, *biz.Order) error) (*biz.Order, error) {
	sqlTx, err := r.data.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sqlTx.Rollback() }()

	drv := entsql.NewDriver(dialect.Postgres, entsql.Conn{ExecQuerier: sqlTx})
	txClient := ent.NewClient(ent.Driver(drv))
	defer txClient.Close()

	q := txClient.Order.Create().
		SetUserID(o.UserID).
		SetSessionID(o.SessionID).
		SetItems(bizItemsToSchema(o.Items)).
		SetTotalCents(o.TotalCents).
		SetCurrency(o.Currency).
		SetStatus(o.Status)
	if o.PaymentID != "" {
		q = q.SetPaymentID(o.PaymentID)
	}
	result, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}
	created := entToOrder(result)

	if err := emit(ctx, sqlTx, created); err != nil {
		return nil, err
	}

	if err := sqlTx.Commit(); err != nil {
		return nil, err
	}
	return created, nil
}

func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*biz.Order, error) {
	o, err := r.data.db.Order.Query().
		Where(entorder.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrOrderNotFound
		}
		return nil, err
	}
	return entToOrder(o), nil
}

func (r *orderRepo) ListByUser(ctx context.Context, userID, status string, page, pageSize int) ([]*biz.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	q := r.data.db.Order.Query().Where(entorder.UserID(userID))
	if status != "" {
		q = q.Where(entorder.Status(status))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	orders, err := q.Offset(offset).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, entToOrder(o))
	}
	return result, total, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*biz.Order, error) {
	updated, err := r.data.db.Order.UpdateOneID(id).SetStatus(status).Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrOrderNotFound
		}
		return nil, err
	}
	return entToOrder(updated), nil
}

func (r *orderRepo) MarkPaid(ctx context.Context, id uuid.UUID, paymentID string) (*biz.Order, error) {
	updated, err := r.data.db.Order.UpdateOneID(id).
		SetPaymentID(paymentID).
		SetStatus("PAID").
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrOrderNotFound
		}
		return nil, err
	}
	return entToOrder(updated), nil
}

func bizItemsToSchema(items []biz.OrderItem) []schema.OrderItem {
	out := make([]schema.OrderItem, len(items))
	for i, it := range items {
		out[i] = schema.OrderItem{
			ProductID:     it.ProductID,
			Name:          it.Name,
			PriceCents:    it.PriceCents,
			Currency:      it.Currency,
			ImageURL:      it.ImageURL,
			Quantity:      it.Quantity,
			SubtotalCents: it.SubtotalCents,
		}
	}
	return out
}

func schemaItemsToBiz(items []schema.OrderItem) []biz.OrderItem {
	out := make([]biz.OrderItem, len(items))
	for i, it := range items {
		out[i] = biz.OrderItem{
			ProductID:     it.ProductID,
			Name:          it.Name,
			PriceCents:    it.PriceCents,
			Currency:      it.Currency,
			ImageURL:      it.ImageURL,
			Quantity:      it.Quantity,
			SubtotalCents: it.SubtotalCents,
		}
	}
	return out
}

func entToOrder(o *ent.Order) *biz.Order {
	return &biz.Order{
		ID:        o.ID,
		UserID:    o.UserID,
		SessionID: o.SessionID,
		Items:     schemaItemsToBiz(o.Items),
		TotalCents: o.TotalCents,
		Currency:  o.Currency,
		Status:    o.Status,
		PaymentID: o.PaymentID,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}
