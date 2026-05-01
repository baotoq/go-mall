package outbox

import (
	"context"
	"database/sql"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS outbox_messages (
    id              TEXT         PRIMARY KEY,
    topic           VARCHAR(255) NOT NULL,
    payload         TEXT         NOT NULL,
    headers         TEXT         NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ  NULL,
    retry_count     INT          NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_error      TEXT         NULL
);
CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed
    ON outbox_messages (next_attempt_at, created_at)
    WHERE processed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outbox_processed
    ON outbox_messages (processed_at)
    WHERE processed_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS inbox_states (
    message_id    VARCHAR(255) PRIMARY KEY,
    topic         VARCHAR(255) NOT NULL,
    consumer_id   VARCHAR(255) NOT NULL,
    received_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    processed_at  TIMESTAMPTZ  NULL,
    error         TEXT         NULL
);
CREATE INDEX IF NOT EXISTS idx_inbox_received_at ON inbox_states (received_at);
`

const insertOutboxSQL = `INSERT INTO outbox_messages (id, topic, payload, headers) VALUES ($1, $2, $3, $4)`

// clock_timestamp() returns wall-clock time, not transaction start time (unlike NOW()).
// This is critical inside the relay's long-running transaction so retry intervals are accurate.

const claimBatchSQL = `
SELECT id, topic, payload, headers, retry_count
FROM outbox_messages
WHERE processed_at IS NULL AND next_attempt_at <= clock_timestamp()
ORDER BY created_at
LIMIT $1
FOR UPDATE SKIP LOCKED`

const markDeliveredSQL = `UPDATE outbox_messages SET processed_at = clock_timestamp(), last_error = NULL WHERE id = $1`

// $3 is backoff in seconds as float64
const markRetrySQL = `UPDATE outbox_messages SET retry_count = $2, next_attempt_at = clock_timestamp() + ($3 * INTERVAL '1 second'), last_error = $4 WHERE id = $1`

const markDeadSQL = `UPDATE outbox_messages SET retry_count = $2, next_attempt_at = clock_timestamp() + INTERVAL '100 years', last_error = $3, processed_at = clock_timestamp() WHERE id = $1`

const sweepOutboxSQL = `DELETE FROM outbox_messages WHERE id IN (SELECT id FROM outbox_messages WHERE processed_at < clock_timestamp() - ($1 * INTERVAL '1 second') LIMIT 1000)`

const sweepInboxSQL = `DELETE FROM inbox_states WHERE message_id IN (SELECT message_id FROM inbox_states WHERE processed_at IS NOT NULL AND processed_at < clock_timestamp() - ($1 * INTERVAL '1 second') LIMIT 1000)`

const inboxClaimSQL = `
INSERT INTO inbox_states (message_id, topic, consumer_id)
VALUES ($1, $2, $3)
ON CONFLICT (message_id) DO UPDATE SET received_at = inbox_states.received_at
-- DO UPDATE (no-op) instead of DO NOTHING so the RETURNING clause always fires.
RETURNING processed_at, (xmax = 0) AS inserted`

const inboxMarkDoneSQL = `UPDATE inbox_states SET processed_at = clock_timestamp(), error = NULL WHERE message_id = $1`

const inboxRollbackSQL = `DELETE FROM inbox_states WHERE message_id = $1 AND processed_at IS NULL`

func applySchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schemaSQL)
	return err
}
