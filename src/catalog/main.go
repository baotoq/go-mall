package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"catalog/ent"
	"catalog/internal/config"
	catalogevent "catalog/internal/event"
	"catalog/internal/handler"
	"catalog/internal/svc"

	dapr "github.com/dapr/go-sdk/client"
	_ "github.com/mattn/go-sqlite3"
	"shared/auth"
	sharedevent "shared/event"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/catalog-api.yaml", "the config file")

var (
	increaseStockRE = regexp.MustCompile(`^/api/v1/products/.+/increase-stock$`)
	confirmRE       = regexp.MustCompile(`^/api/v1/reservations/.+/confirm$`)
	cancelRE        = regexp.MustCompile(`^/api/v1/reservations/.+/cancel$`)
)

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

	ctx := svc.NewServiceContext(c, db, sharedevent.NewOutboxDispatcher[sharedevent.Event](
		sharedevent.NewDaprDispatcher[sharedevent.Event](dapr),
		catalogevent.NewEntStore(db),
	))
	handler.RegisterHandlers(server, ctx)

	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method

			// Public routes: GET products, GET categories
			if (method == http.MethodGet && strings.HasPrefix(path, "/api/v1/products")) ||
				(method == http.MethodGet && strings.HasPrefix(path, "/api/v1/categories")) {
				next(w, r)
				return
			}

			// Admin routes: require admin role
			if (method == http.MethodPost && path == "/api/v1/products") ||
				(method == http.MethodPut && strings.HasPrefix(path, "/api/v1/products")) ||
				(method == http.MethodDelete && strings.HasPrefix(path, "/api/v1/products")) ||
				(method == http.MethodPost && path == "/api/v1/categories") ||
				(method == http.MethodPost && increaseStockRE.MatchString(path)) {
				authMiddleware := auth.RequireRole(ctx.Validator, "admin")
				authMiddleware(next)(w, r)
				return
			}

			// Service routes: require auth, no role check
			if (method == http.MethodPost && path == "/api/v1/reservations") ||
				(method == http.MethodPost && confirmRE.MatchString(path)) ||
				(method == http.MethodPost && cancelRE.MatchString(path)) {
				authMiddleware := auth.RequireAuth(ctx.Validator)
				authMiddleware(next)(w, r)
				return
			}

			// Default: require auth
			authMiddleware := auth.RequireAuth(ctx.Validator)
			authMiddleware(next)(w, r)
		}
	})

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
	if err := seedIfEmpty(context.Background(), db); err != nil {
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
