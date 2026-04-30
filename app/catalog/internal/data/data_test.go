package data_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"gomall/app/catalog/internal/data/ent"
)

var testClient *ent.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("catalog_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres: %v\n", err)
		os.Exit(1)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection string: %v\n", err)
		_ = testcontainers.TerminateContainer(pg)
		os.Exit(1)
	}

	testClient, err = ent.Open(dialect.Postgres, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ent open: %v\n", err)
		_ = testcontainers.TerminateContainer(pg)
		os.Exit(1)
	}

	if err := testClient.Schema.Create(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "schema create: %v\n", err)
		testClient.Close()
		_ = testcontainers.TerminateContainer(pg)
		os.Exit(1)
	}

	code := m.Run()
	testClient.Close()
	_ = testcontainers.TerminateContainer(pg)
	os.Exit(code)
}

func truncate(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	testClient.Product.Delete().ExecX(ctx)
	testClient.Category.Delete().ExecX(ctx)
}
