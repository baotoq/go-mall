package outbox

import (
	"context"
	"encoding/json"
	"fmt"
)

// TypedHandler wraps a typed handler function into a Handler.
func TypedHandler[T any](fn func(ctx context.Context, evt T) error) Handler {
	return func(ctx context.Context, msg Message) error {
		var v T
		if err := json.Unmarshal(msg.Payload, &v); err != nil {
			return fmt.Errorf("outbox: typed unmarshal: %w", err)
		}
		return fn(ctx, v)
	}
}
