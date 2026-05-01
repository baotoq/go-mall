package outbox_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"gomall/pkg/outbox"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgC, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("outbox_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testcontainers postgres: %v\n", err)
		os.Exit(1)
	}

	dsn, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection string: %v\n", err)
		_ = testcontainers.TerminateContainer(pgC)
		os.Exit(1)
	}

	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql open: %v\n", err)
		_ = testcontainers.TerminateContainer(pgC)
		os.Exit(1)
	}

	// Apply schema once for all tests.
	if err := applyTestSchema(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "schema: %v\n", err)
		_ = testDB.Close()
		_ = testcontainers.TerminateContainer(pgC)
		os.Exit(1)
	}

	code := m.Run()

	_ = testDB.Close()
	_ = testcontainers.TerminateContainer(pgC)
	os.Exit(code)
}

func applyTestSchema(ctx context.Context) error {
	cfg := outbox.DefaultConfig()
	client, err := outbox.NewWithPublisher(testDB, &fakePub{}, cfg, kratoslog.DefaultLogger)
	if err != nil {
		return err
	}
	return client.Migrate(ctx)
}

func resetTables(t *testing.T) {
	t.Helper()
	_, err := testDB.ExecContext(context.Background(), `TRUNCATE outbox_messages, inbox_states`)
	require.NoError(t, err)
}

func defaultCfg() outbox.Config {
	cfg := outbox.DefaultConfig()
	cfg.EnableRelay = false
	cfg.ConsumerID = "test-consumer"
	return cfg
}

func noopLogger() kratoslog.Logger {
	return kratoslog.DefaultLogger
}

// fakePub implements the unexported publisher interface via outbox_test package.
type fakePub struct {
	mu      sync.Mutex
	calls   []string
	errFn   func(call int) error
	callIdx int
}

func (f *fakePub) PublishEvent(_ context.Context, _, topic string, _ interface{}, _ ...dapr.PublishEventOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, topic)
	idx := f.callIdx
	f.callIdx++
	if f.errFn != nil {
		return f.errFn(idx)
	}
	return nil
}

func (f *fakePub) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// ---- Tests ----

func TestMigrate_Idempotent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cfg := defaultCfg()

	// Act — apply schema a second time (first applied in TestMain)
	c, err := outbox.NewWithPublisher(testDB, &fakePub{}, cfg, noopLogger())
	require.NoError(t, err)
	err = c.Migrate(ctx)

	// Assert
	assert.NoError(t, err)
}

func TestPublish_TransactionalAtomicity(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	c, err := outbox.NewWithPublisher(testDB, &fakePub{}, cfg, noopLogger())
	require.NoError(t, err)

	// Act — publish inside a tx, then rollback
	tx, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = c.Publish(ctx, tx, "test.topic", map[string]string{"k": "v"})
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())

	// Assert — no row in outbox
	var count int
	require.NoError(t, testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_messages`).Scan(&count))
	assert.Equal(t, 0, count)
}

func TestPublish_CommitLandsRow(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	c, err := outbox.NewWithPublisher(testDB, &fakePub{}, cfg, noopLogger())
	require.NoError(t, err)

	// Act — publish inside a tx, then commit
	tx, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	msgID, err := c.Publish(ctx, tx, "order.created", map[string]string{"order_id": "123"})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	// Assert — 1 row with correct topic and null processed_at
	var topic string
	var processedAt sql.NullTime
	row := testDB.QueryRowContext(ctx, `SELECT topic, processed_at FROM outbox_messages WHERE id = $1`, msgID)
	require.NoError(t, row.Scan(&topic, &processedAt))
	assert.Equal(t, "order.created", topic)
	assert.False(t, processedAt.Valid)
}

func TestRelay_DeliversBatch(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	fp := &fakePub{}
	c, err := outbox.NewWithPublisher(testDB, fp, cfg, noopLogger())
	require.NoError(t, err)

	// Insert 3 rows manually
	for i := 0; i < 3; i++ {
		tx, _ := testDB.BeginTx(ctx, nil)
		_, err = c.Publish(ctx, tx, "test.deliver", map[string]string{"i": fmt.Sprintf("%d", i)})
		require.NoError(t, err)
		require.NoError(t, tx.Commit())
	}

	// Act
	require.NoError(t, c.ProcessBatch(ctx))

	// Assert — all 3 have processed_at set
	var count int
	require.NoError(t, testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_messages WHERE processed_at IS NOT NULL`).Scan(&count))
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, fp.callCount())
}

func TestRelay_RetryOnError(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	callCount := 0
	fp := &fakePub{
		errFn: func(call int) error {
			callCount++
			if call == 0 {
				return errors.New("transient error")
			}
			return nil
		},
	}
	c, err := outbox.NewWithPublisher(testDB, fp, cfg, noopLogger())
	require.NoError(t, err)

	tx, _ := testDB.BeginTx(ctx, nil)
	msgID, _ := c.Publish(ctx, tx, "test.retry", map[string]string{})
	_ = tx.Commit()

	// Act — first batch: publish fails → retry_count increments
	require.NoError(t, c.ProcessBatch(ctx))

	var retryCount int
	var processedAt sql.NullTime
	row := testDB.QueryRowContext(ctx, `SELECT retry_count, processed_at FROM outbox_messages WHERE id = $1`, msgID)
	require.NoError(t, row.Scan(&retryCount, &processedAt))
	assert.Equal(t, 1, retryCount)
	assert.False(t, processedAt.Valid)

	// Reset next_attempt_at so second batch picks it up immediately
	_, err = testDB.ExecContext(ctx, `UPDATE outbox_messages SET next_attempt_at = NOW() WHERE id = $1`, msgID)
	require.NoError(t, err)

	// Act — second batch: publish succeeds → processed_at set
	require.NoError(t, c.ProcessBatch(ctx))

	row = testDB.QueryRowContext(ctx, `SELECT processed_at FROM outbox_messages WHERE id = $1`, msgID)
	require.NoError(t, row.Scan(&processedAt))
	assert.True(t, processedAt.Valid)
}

func TestRelay_DeadLetter(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	cfg.MaxAttempts = 2
	fp := &fakePub{errFn: func(_ int) error { return errors.New("always fails") }}
	c, err := outbox.NewWithPublisher(testDB, fp, cfg, noopLogger())
	require.NoError(t, err)

	tx, _ := testDB.BeginTx(ctx, nil)
	msgID, _ := c.Publish(ctx, tx, "test.dead", map[string]string{})
	_ = tx.Commit()

	// Act — need MaxAttempts processBatch calls to dead-letter
	// After each failed attempt, reset next_attempt_at to NOW() so it's picked up again
	for i := 0; i < cfg.MaxAttempts; i++ {
		_, _ = testDB.ExecContext(ctx, `UPDATE outbox_messages SET next_attempt_at = NOW() WHERE id = $1`, msgID)
		require.NoError(t, c.ProcessBatch(ctx))
	}

	// Assert — next_attempt_at is far in the future (dead-lettered)
	var nextAttempt time.Time
	row := testDB.QueryRowContext(ctx, `SELECT next_attempt_at FROM outbox_messages WHERE id = $1`, msgID)
	require.NoError(t, row.Scan(&nextAttempt))
	assert.True(t, nextAttempt.After(time.Now().Add(50*365*24*time.Hour)), "expected dead letter in far future, got %v", nextAttempt)
}

func TestRelay_SkipLockedConcurrent(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	cfg.BatchSize = 10

	fp := &fakePub{
		errFn: func(_ int) error {
			return nil
		},
	}

	c, err := outbox.NewWithPublisher(testDB, fp, cfg, noopLogger())
	require.NoError(t, err)

	// Insert 5 rows
	for i := 0; i < 5; i++ {
		tx, _ := testDB.BeginTx(ctx, nil)
		_, _ = c.Publish(ctx, tx, fmt.Sprintf("test.concurrent.%d", i), map[string]string{})
		_ = tx.Commit()
	}

	// Act — run 2 concurrent processBatch calls
	var wg sync.WaitGroup
	wg.Add(2)
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			defer wg.Done()
			errs[i] = c.ProcessBatch(ctx)
		}()
	}
	wg.Wait()

	for _, e := range errs {
		assert.NoError(t, e)
	}

	// Assert — each row published exactly once (total call count == 5)
	fp.mu.Lock()
	total := len(fp.calls)
	fp.mu.Unlock()
	assert.Equal(t, 5, total)

	// All processed
	var count int
	require.NoError(t, testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_messages WHERE processed_at IS NOT NULL`).Scan(&count))
	assert.Equal(t, 5, count)
}

func TestSubscribe_HappyPath(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	c, err := outbox.NewWithPublisher(testDB, &fakePub{}, cfg, noopLogger())
	require.NoError(t, err)

	handlerCalled := 0
	handler := func(_ context.Context, msg outbox.Message) error {
		handlerCalled++
		return nil
	}

	// Act
	fn := c.Subscribe("test.topic", handler)
	retry, err := fn(ctx, &daprTopicEvent{ID: "msg-001", Topic: "test.topic", RawData: []byte(`"hello"`)})

	// Assert
	require.NoError(t, err)
	assert.False(t, retry)
	assert.Equal(t, 1, handlerCalled)

	// processed_at should be set
	var processedAt sql.NullTime
	row := testDB.QueryRowContext(ctx, `SELECT processed_at FROM inbox_states WHERE message_id = 'msg-001'`)
	require.NoError(t, row.Scan(&processedAt))
	assert.True(t, processedAt.Valid)
}

func TestSubscribe_TrueDuplicate(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	c, err := outbox.NewWithPublisher(testDB, &fakePub{}, cfg, noopLogger())
	require.NoError(t, err)

	handlerCalled := 0
	handler := func(_ context.Context, msg outbox.Message) error {
		handlerCalled++
		return nil
	}

	// Act — call twice with same event ID
	fn := c.Subscribe("test.dedup", handler)
	retry1, err1 := fn(ctx, &daprTopicEvent{ID: "dup-msg-001", Topic: "test.dedup", RawData: []byte(`{}`)})
	retry2, err2 := fn(ctx, &daprTopicEvent{ID: "dup-msg-001", Topic: "test.dedup", RawData: []byte(`{}`)})

	// Assert — handler called once; second returns (false, nil)
	assert.NoError(t, err1)
	assert.False(t, retry1)
	assert.NoError(t, err2)
	assert.False(t, retry2)
	assert.Equal(t, 1, handlerCalled)
}

func TestSweep_DeletesOldRows(t *testing.T) {
	// Arrange
	resetTables(t)
	ctx := context.Background()
	cfg := defaultCfg()
	cfg.RetentionPeriod = 24 * time.Hour
	c, err := outbox.NewWithPublisher(testDB, &fakePub{}, cfg, noopLogger())
	require.NoError(t, err)

	// Insert a processed row with processed_at = NOW()-48h
	_, err = testDB.ExecContext(ctx, `
		INSERT INTO outbox_messages (id, topic, payload, headers, processed_at, next_attempt_at)
		VALUES ('old-msg-001', 'test.sweep', '{}', '{}', NOW() - INTERVAL '48 hours', NOW() - INTERVAL '48 hours')
	`)
	require.NoError(t, err)

	// Act
	c.RunSweep(ctx)

	// Assert — row deleted
	var count int
	require.NoError(t, testDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_messages WHERE id = 'old-msg-001'`).Scan(&count))
	assert.Equal(t, 0, count)
}

// daprTopicEvent aliases the Dapr SDK TopicEvent for use in test helpers.
type daprTopicEvent = common.TopicEvent
