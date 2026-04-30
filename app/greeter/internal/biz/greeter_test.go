package biz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/greeter/internal/biz"
)

type stubGreeterRepo struct {
	saved   *biz.Greeter
	saveErr error
	findErr error
}

func (r *stubGreeterRepo) Save(_ context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	g.ID = 42
	r.saved = g
	return g, nil
}
func (r *stubGreeterRepo) Update(_ context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	return g, nil
}
func (r *stubGreeterRepo) FindByID(_ context.Context, id int64) (*biz.Greeter, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return &biz.Greeter{ID: int(id), Hello: "world"}, nil
}
func (r *stubGreeterRepo) ListByHello(_ context.Context, _ string) ([]*biz.Greeter, error) {
	return nil, nil
}
func (r *stubGreeterRepo) ListAll(_ context.Context) ([]*biz.Greeter, error) {
	return nil, nil
}

func TestGreeterUsecase_CreateGreeter(t *testing.T) {
	repo := &stubGreeterRepo{}
	uc := biz.NewGreeterUsecase(repo)

	got, err := uc.CreateGreeter(context.Background(), &biz.Greeter{Hello: "kratos"})

	require.NoError(t, err)
	assert.Equal(t, 42, got.ID)
	assert.Equal(t, "kratos", repo.saved.Hello)
}

func TestGreeterUsecase_CreateGreeter_propagatesError(t *testing.T) {
	repo := &stubGreeterRepo{saveErr: errors.New("boom")}
	uc := biz.NewGreeterUsecase(repo)

	_, err := uc.CreateGreeter(context.Background(), &biz.Greeter{Hello: "x"})

	assert.Error(t, err)
}

func TestGreeterUsecase_GetGreeter(t *testing.T) {
	uc := biz.NewGreeterUsecase(&stubGreeterRepo{})

	got, err := uc.GetGreeter(context.Background(), 7)

	require.NoError(t, err)
	assert.Equal(t, 7, got.ID)
	assert.Equal(t, "world", got.Hello)
}

func TestGreeterUsecase_GetGreeter_propagatesError(t *testing.T) {
	uc := biz.NewGreeterUsecase(&stubGreeterRepo{findErr: errors.New("nope")})

	_, err := uc.GetGreeter(context.Background(), 1)

	assert.Error(t, err)
}
