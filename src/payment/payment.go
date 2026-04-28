// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"context"
	"flag"
	"fmt"

	"payment/ent"
	"payment/internal/config"
	paymentevent "payment/internal/event"
	"payment/internal/handler"
	"payment/internal/svc"

	dapr "github.com/dapr/go-sdk/client"
	sharedevent "shared/event"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	_ "github.com/mattn/go-sqlite3"
)

var configFile = flag.String("f", "etc/payment-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	db := initDb()
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

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
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
