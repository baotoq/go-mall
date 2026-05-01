package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	dapr "github.com/dapr/go-sdk/client"
)

type claimedRow struct {
	id         string
	topic      string
	payload    []byte
	headers    []byte
	retryCount int
}

func (c *Client) runRelay(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := c.processBatch(ctx); err != nil {
				c.log.Errorf("outbox relay batch error: %v", err)
			}
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) processBatch(ctx context.Context) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, claimBatchSQL, c.cfg.BatchSize)
	if err != nil {
		return err
	}
	var batch []claimedRow
	for rows.Next() {
		var r claimedRow
		if err := rows.Scan(&r.id, &r.topic, &r.payload, &r.headers, &r.retryCount); err != nil {
			_ = rows.Close()
			return err
		}
		batch = append(batch, r)
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, r := range batch {
		headers := make(map[string]string)
		if len(r.headers) > 0 {
			_ = json.Unmarshal(r.headers, &headers)
		}
		// Propagate message ID so the consumer inbox can deduplicate.
		headers["cloudevent.id"] = r.id

		pubCtx, cancel := context.WithTimeout(ctx, c.cfg.PublishTimeout)
		pubErr := c.pub.PublishEvent(pubCtx, c.cfg.PubsubName, r.topic, r.payload,
			dapr.PublishEventWithContentType("application/json"),
			dapr.PublishEventWithMetadata(headers),
		)
		cancel()

		if pubErr == nil {
			if _, err := tx.ExecContext(ctx, markDeliveredSQL, r.id); err != nil {
				return fmt.Errorf("outbox: mark delivered %s: %w", r.id, err)
			}
		} else if r.retryCount+1 >= c.cfg.MaxRetries {
			if _, err := tx.ExecContext(ctx, markDeadSQL, r.id, r.retryCount+1, truncateErrMsg(pubErr)); err != nil {
				return fmt.Errorf("outbox: mark dead %s: %w", r.id, err)
			}
		} else {
			backoff := computeBackoff(c.cfg.BackoffBase, c.cfg.BackoffMax, r.retryCount)
			if _, err := tx.ExecContext(ctx, markRetrySQL, r.id, r.retryCount+1, backoff.Seconds(), truncateErrMsg(pubErr)); err != nil {
				return fmt.Errorf("outbox: mark retry %s: %w", r.id, err)
			}
		}
	}

	return tx.Commit()
}

func (c *Client) runSweeper(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.SweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			retentionSecs := c.cfg.RetentionPeriod.Seconds()
			if _, err := c.db.ExecContext(ctx, sweepOutboxSQL, retentionSecs); err != nil {
				c.log.Errorf("outbox sweep error: %v", err)
			}
			if _, err := c.db.ExecContext(ctx, sweepInboxSQL, retentionSecs); err != nil {
				c.log.Errorf("inbox sweep error: %v", err)
			}
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func computeBackoff(base, max time.Duration, retryCount int) time.Duration {
	if retryCount > 30 {
		retryCount = 30
	}
	exp := math.Pow(2, float64(retryCount))
	backoff := time.Duration(float64(base) * exp)
	if backoff > max || backoff < 0 {
		backoff = max
	}
	if jitterMax := base / 4; jitterMax > 0 {
		backoff += time.Duration(rand.Int63n(int64(jitterMax)))
	}
	return backoff
}

func truncateErrMsg(err error) string {
	const maxLen = 1024
	s := err.Error()
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
