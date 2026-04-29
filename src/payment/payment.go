// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"payment/ent"
	"payment/internal/config"
	paymentevent "payment/internal/event"
	"payment/internal/handler"
	"payment/internal/svc"
	"shared/auth"
	"shared/health"

	dapr "github.com/dapr/go-sdk/client"
	sharedevent "shared/event"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	_ "github.com/lib/pq"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	db := initDb(c.DB.DSN)
	if db != nil {
		defer db.Close()
	}

	daprClient := initDaprClient()
	if daprClient != nil {
		defer daprClient.Close()
	}

	ctx := svc.NewServiceContext(c, db, sharedevent.NewOutboxDispatcher[sharedevent.Event](
		sharedevent.NewDaprDispatcher[sharedevent.Event](daprClient),
		paymentevent.NewEntStore(db),
	))
	handler.RegisterHandlers(server, ctx)
	health.Register(server, health.DaprProbe{Client: daprClient})

	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
				next(w, r)
				return
			}
			auth.RequireServiceAuth(ctx.Validator)(next)(w, r)
		}
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func initDb(dsn string) *ent.Client {
	db, err := ent.Open("postgres", dsn)
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
