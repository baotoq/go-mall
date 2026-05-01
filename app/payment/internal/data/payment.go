package data

import (
	"context"

	"gomall/app/payment/internal/biz"
	"gomall/app/payment/internal/data/ent"
	entpayment "gomall/app/payment/internal/data/ent/payment"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type paymentRepo struct {
	data *Data
	log  *log.Helper
}

func NewPaymentRepo(data *Data, logger log.Logger) biz.PaymentRepo {
	return &paymentRepo{data: data, log: log.NewHelper(logger)}
}

func (r *paymentRepo) Create(ctx context.Context, p *biz.Payment) (*biz.Payment, error) {
	q := r.data.db.Payment.Create().
		SetOrderID(p.OrderID).
		SetUserID(p.UserID).
		SetAmountCents(p.AmountCents).
		SetProvider(p.Provider).
		SetStatus(p.Status)
	if p.Currency != "" {
		q = q.SetCurrency(p.Currency)
	}
	result, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}
	return entToPayment(result), nil
}

func (r *paymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*biz.Payment, error) {
	p, err := r.data.db.Payment.Query().
		Where(entpayment.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrPaymentNotFound
		}
		return nil, err
	}
	return entToPayment(p), nil
}

func (r *paymentRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*biz.Payment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := r.data.db.Payment.Query().
		Where(entpayment.UserID(userID)).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	payments, err := r.data.db.Payment.Query().
		Where(entpayment.UserID(userID)).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.Payment, 0, len(payments))
	for _, p := range payments {
		result = append(result, entToPayment(p))
	}
	return result, total, nil
}

func (r *paymentRepo) ListByOrder(ctx context.Context, orderID string) ([]*biz.Payment, error) {
	payments, err := r.data.db.Payment.Query().
		Where(entpayment.OrderID(orderID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.Payment, 0, len(payments))
	for _, p := range payments {
		result = append(result, entToPayment(p))
	}
	return result, nil
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*biz.Payment, error) {
	updated, err := r.data.db.Payment.UpdateOneID(id).SetStatus(status).Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrPaymentNotFound
		}
		return nil, err
	}
	return entToPayment(updated), nil
}

func entToPayment(p *ent.Payment) *biz.Payment {
	return &biz.Payment{
		ID:          p.ID,
		OrderID:     p.OrderID,
		UserID:      p.UserID,
		AmountCents: p.AmountCents,
		Currency:    p.Currency,
		Status:      p.Status,
		Provider:    p.Provider,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
