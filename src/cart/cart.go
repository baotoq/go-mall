package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"cart/ent"
	catalogclient "cart/internal/clients/catalog"
	paymentclient "cart/internal/clients/payment"
	"cart/internal/config"
	cartevent "cart/internal/event"
	"cart/internal/handler"
	"cart/internal/svc"

	dapr "github.com/dapr/go-sdk/client"
	_ "github.com/mattn/go-sqlite3"
	sharedevent "shared/event"
	"shared/auth"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/cart-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	db := initDb()
	defer db.Close()

	daprClient := initDaprClient()
	defer daprClient.Close()

	ctx := svc.NewServiceContext(
		c,
		db,
		sharedevent.NewOutboxDispatcher[sharedevent.Event](
			sharedevent.NewDaprDispatcher[sharedevent.Event](daprClient),
			cartevent.NewEntStore(db),
		),
		catalogclient.New("", c.CatalogAppID),
		paymentclient.New("", c.PaymentAppID),
	)

	if token, err := ctx.ServiceClient.GetToken(context.Background()); err == nil {
		ctx.CatalogClient.SetAuthToken(token)
		ctx.PaymentClient.SetAuthToken(token)
	}

	handler.RegisterHandlers(server, ctx)
	server.Use(pathAuthMiddleware(ctx.Validator))

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func pathAuthMiddleware(validator auth.TokenValidator) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method

			if method == http.MethodGet && path == "/api/v1/cart/items" {
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(path, "/api/v1/cart/items") ||
				strings.HasPrefix(path, "/api/v1/cart/checkout") {
				auth.RequireAuth(validator)(next)(w, r)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

func initDb() *ent.Client {
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")

	if err != nil {
		logx.Must(err)
		return nil
	}
	if err := db.Schema.Create(context.Background()); err != nil {
		logx.Must(err)
		return nil
	}

	return db
}

func initDaprClient() dapr.Client {
	daprClient, err := dapr.NewClient()
	if err != nil {
		logx.Must(err)
		return nil
	}
	return daprClient
}
