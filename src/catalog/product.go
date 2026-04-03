// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"

	"catalog/ent"
	"catalog/internal/config"
	"catalog/internal/event"
	"catalog/internal/handler"
	"catalog/internal/svc"

	dapr "github.com/dapr/go-sdk/client"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/catalog-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	db := initDb()
	defer db.Close()

	dapr := initDaprClient()
	defer dapr.Close()

	ctx := svc.NewServiceContext(c, db, event.NewOutboxDispatcher(event.NewDaprDispatcher[event.Event](dapr), db))
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
