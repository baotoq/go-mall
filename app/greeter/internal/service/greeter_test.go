package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "gomall/api/greeter/helloworld/v1"
	"gomall/app/greeter/internal/biz"
	"gomall/app/greeter/internal/service"
)

type nopGreeterRepo struct {
	saveErr error
	findErr error
}

func (r *nopGreeterRepo) Save(_ context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	g.ID = 1
	return g, nil
}
func (r *nopGreeterRepo) Update(_ context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	return g, nil
}
func (r *nopGreeterRepo) FindByID(_ context.Context, id int64) (*biz.Greeter, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return &biz.Greeter{ID: int(id), Hello: "found"}, nil
}
func (r *nopGreeterRepo) ListByHello(_ context.Context, _ string) ([]*biz.Greeter, error) {
	return nil, nil
}
func (r *nopGreeterRepo) ListAll(_ context.Context) ([]*biz.Greeter, error) { return nil, nil }

func newGreeterSvc(repo *nopGreeterRepo) *service.GreeterService {
	return service.NewGreeterService(biz.NewGreeterUsecase(repo))
}

func TestGreeterService_SayHello(t *testing.T) {
	got, err := newGreeterSvc(&nopGreeterRepo{}).SayHello(context.Background(), &v1.HelloRequest{Name: "kratos"})

	require.NoError(t, err)
	assert.Equal(t, "Hello kratos", got.Message)
}

func TestGreeterService_SayHello_propagatesError(t *testing.T) {
	_, err := newGreeterSvc(&nopGreeterRepo{saveErr: errors.New("x")}).SayHello(
		context.Background(), &v1.HelloRequest{Name: "kratos"})

	assert.Error(t, err)
}

func TestGreeterService_CreateGreeter(t *testing.T) {
	got, err := newGreeterSvc(&nopGreeterRepo{}).CreateGreeter(context.Background(), &v1.CreateGreeterRequest{Hello: "hi"})

	require.NoError(t, err)
	assert.Equal(t, "hi", got.Hello)
	assert.Equal(t, int64(1), got.Id)
}

func TestGreeterService_GetGreeter(t *testing.T) {
	got, err := newGreeterSvc(&nopGreeterRepo{}).GetGreeter(context.Background(), &v1.GetGreeterRequest{Id: 9})

	require.NoError(t, err)
	assert.Equal(t, int64(9), got.Id)
	assert.Equal(t, "found", got.Hello)
}

func TestGreeterService_GetGreeter_propagatesError(t *testing.T) {
	_, err := newGreeterSvc(&nopGreeterRepo{findErr: errors.New("x")}).GetGreeter(
		context.Background(), &v1.GetGreeterRequest{Id: 1})

	assert.Error(t, err)
}
