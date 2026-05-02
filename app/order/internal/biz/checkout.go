package biz

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/workflow"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrCheckoutMissingKey is returned when Schedule is called without an
// idempotency_key. Service-layer validators should catch this earlier; the
// biz-layer guard exists for direct callers (tests, future internal flows).
var ErrCheckoutMissingKey = errors.New("checkout: idempotency_key required")

// ErrInvalidSagaConfig is returned when the SagaConfig in use is missing
// required values (e.g. MaxPaymentAttempts <= 0) that would cause the saga
// to behave incorrectly.
var ErrInvalidSagaConfig = errors.New("checkout: invalid saga config")

// ErrCheckoutDuplicateKey is returned when an idempotency key is reused with a
// different user_id, or when a key is presented after the workflow has been
// purged (use a new key).
var ErrCheckoutDuplicateKey = errors.New("checkout duplicate key")

// IdempotencyKeyRepo persists checkout idempotency keys so that duplicate
// Schedule calls with the same key + same user return the previously scheduled
// checkout. The data-layer implementation lands in Wave 2b alongside the ent
// schema for idempotency_keys.
type IdempotencyKeyRepo interface {
	// Get returns the stored checkout record for the given key, if any.
	// hit=false means the key has never been seen.
	Get(ctx context.Context, key string) (StoredCheckout, bool, error)
	// Put stores a new idempotency-key entry.
	Put(ctx context.Context, key string, sc StoredCheckout) error
}

// StoredCheckout is the value persisted under an idempotency key.
type StoredCheckout struct {
	CheckoutID string
	UserID     string
	OrderID    string
}

// CheckoutUsecase schedules and queries OrderSaga workflow instances.
type CheckoutUsecase struct {
	wfc      *workflow.Client
	idemRepo IdempotencyKeyRepo
	log      *log.Helper
	cfg      SagaConfig
}

// NewCheckoutUsecase constructs a CheckoutUsecase.
func NewCheckoutUsecase(
	wfc *workflow.Client,
	idemRepo IdempotencyKeyRepo,
	cfg SagaConfig,
	logger log.Logger,
) *CheckoutUsecase {
	return &CheckoutUsecase{
		wfc:      wfc,
		idemRepo: idemRepo,
		log:      log.NewHelper(logger),
		cfg:      cfg,
	}
}

// Schedule starts a new OrderSaga for the given checkout input.
//
// Delta D duplicate detection (plan v0.4 §3.6):
//  1. Check idempotency-key store; return stored response on hit with matching user_id.
//  2. Call wfc.ScheduleWorkflow with WithInstanceID(idempotencyKey).
//  3. On codes.AlreadyExists (gRPC) or "already exists" string → fetch existing.
//  4. Store new entry on success.
func (uc *CheckoutUsecase) Schedule(ctx context.Context, in CheckoutInput) (checkoutID, orderID string, err error) {
	if in.IdempotencyKey == "" {
		return "", "", ErrCheckoutMissingKey
	}

	// 1. Check idempotency-key store.
	if uc.idemRepo != nil {
		stored, hit, gerr := uc.idemRepo.Get(ctx, in.IdempotencyKey)
		if gerr != nil {
			uc.log.WithContext(ctx).Errorf("idem key lookup: %v", gerr)
			// non-fatal: proceed to schedule
		} else if hit {
			if stored.UserID != in.UserID {
				// Different user presented the same key — reject.
				return "", "", ErrCheckoutDuplicateKey
			}
			// Same user: return the stored checkout_id (idempotent response).
			if stored.OrderID == "" {
				// OrderID was not yet known when the entry was stored — try
				// to recover it from the workflow's own output.
				return uc.fetchExisting(ctx, in.IdempotencyKey)
			}
			return stored.CheckoutID, stored.OrderID, nil
		}
	}

	// Guard: SagaConfig must define a positive MaxPaymentAttempts; otherwise
	// the saga payment loop would never run and the workflow would compensate
	// immediately. Surface as a config error rather than silently failing.
	if uc.cfg.MaxPaymentAttempts <= 0 {
		return "", "", ErrInvalidSagaConfig
	}

	// 2. Schedule the workflow.
	id, schedErr := uc.wfc.ScheduleWorkflow(ctx, "OrderSaga",
		workflow.WithInstanceID(in.IdempotencyKey),
		workflow.WithInput(in),
	)
	if schedErr != nil {
		// 3. Detect duplicate-instance error (Delta D).
		if st, ok := status.FromError(unwrapAll(schedErr)); ok && st.Code() == codes.AlreadyExists {
			return uc.fetchExisting(ctx, in.IdempotencyKey)
		}
		if strings.Contains(schedErr.Error(), "already exists") {
			return uc.fetchExisting(ctx, in.IdempotencyKey)
		}
		return "", "", schedErr
	}

	// 4. Persist idempotency entry.
	if uc.idemRepo != nil {
		if putErr := uc.idemRepo.Put(ctx, in.IdempotencyKey, StoredCheckout{
			CheckoutID: id,
			UserID:     in.UserID,
		}); putErr != nil {
			uc.log.WithContext(ctx).Errorf("idem key put: %v", putErr)
			// non-fatal: the workflow is already scheduled
		}
	}
	return id, "", nil
}

// fetchExisting retrieves the checkout_id and order_id for an already-scheduled
// workflow instance. When the stored entry has an empty OrderID (e.g. the
// activity that produces it had not yet completed at Put time), this falls
// back to the workflow's own metadata to recover the order_id from the saga's
// CheckoutResult output, then refreshes the idempotency-key store best-effort
// so subsequent Get calls return it directly.
func (uc *CheckoutUsecase) fetchExisting(ctx context.Context, idemKey string) (checkoutID, orderID string, err error) {
	var stored StoredCheckout
	hit := false
	if uc.idemRepo != nil {
		var gerr error
		stored, hit, gerr = uc.idemRepo.Get(ctx, idemKey)
		if gerr == nil && hit && stored.OrderID != "" {
			return stored.CheckoutID, stored.OrderID, nil
		}
	}

	// Try to recover OrderID from the workflow's own output.
	checkoutID = idemKey
	if hit {
		checkoutID = stored.CheckoutID
	}
	if uc.wfc != nil {
		if res, statusErr := uc.Status(ctx, checkoutID); statusErr == nil && res.OrderID != "" {
			orderID = res.OrderID
			if uc.idemRepo != nil && hit {
				stored.OrderID = orderID
				_ = uc.idemRepo.Put(ctx, idemKey, stored)
			}
			return checkoutID, orderID, nil
		}
	}
	return checkoutID, "", nil
}

// Status fetches the current state of a checkout workflow.
func (uc *CheckoutUsecase) Status(ctx context.Context, checkoutID string) (CheckoutResult, error) {
	meta, err := uc.wfc.FetchWorkflowMetadata(ctx, checkoutID, workflow.WithFetchPayloads(true))
	if err != nil {
		return CheckoutResult{}, err
	}
	result := CheckoutResult{
		State: workflowStateString(meta),
	}
	// Deserialise the serialised output into CheckoutResult when available.
	// WorkflowMetadata is protos.OrchestrationMetadata; output field is Output.
	if meta != nil && meta.Output != nil {
		raw := meta.Output.GetValue()
		if raw != "" {
			var cr CheckoutResult
			if jsonErr := json.Unmarshal([]byte(raw), &cr); jsonErr == nil {
				result = cr
			} else {
				uc.log.WithContext(ctx).Warnf("checkout Status: decode workflow output: %v", jsonErr)
			}
		}
	}
	return result, nil
}

// workflowStateString maps a Dapr workflow runtime status to a clean
// CheckoutResult state string.
func workflowStateString(meta *workflow.WorkflowMetadata) string {
	if meta == nil {
		return "UNKNOWN"
	}
	switch meta.RuntimeStatus {
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_RUNNING:
		return "RUNNING"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_COMPLETED:
		return "COMPLETED"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_FAILED:
		return "FAILED"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_CANCELED:
		return "CANCELED"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_TERMINATED:
		return "TERMINATED"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_PENDING:
		return "PENDING"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_SUSPENDED:
		return "SUSPENDED"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_CONTINUED_AS_NEW:
		return "CONTINUED_AS_NEW"
	case protos.OrchestrationStatus_ORCHESTRATION_STATUS_STALLED:
		return "STALLED"
	default:
		return "UNKNOWN"
	}
}

// unwrapAll walks the full error chain (single and multi-error) and returns
// the deepest error. Used to surface wrapped gRPC status codes.
func unwrapAll(err error) error {
	type multiUnwrapper interface{ Unwrap() []error }
	for {
		if next := errors.Unwrap(err); next != nil {
			err = next
			continue
		}
		if mu, ok := err.(multiUnwrapper); ok {
			found := false
			for _, e := range mu.Unwrap() {
				if e != nil {
					err = e
					found = true
					break
				}
			}
			if found {
				continue
			}
		}
		return err
	}
}
