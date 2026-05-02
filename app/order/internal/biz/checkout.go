package biz

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		return "", "", ErrOrderEmptyItems // placeholder; service layer validates UUID format
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
			return stored.CheckoutID, stored.OrderID, nil
		}
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

// fetchExisting retrieves the checkout_id and (empty) order_id for an already-
// scheduled workflow instance.
func (uc *CheckoutUsecase) fetchExisting(ctx context.Context, idemKey string) (checkoutID, orderID string, err error) {
	if uc.idemRepo != nil {
		stored, hit, gerr := uc.idemRepo.Get(ctx, idemKey)
		if gerr == nil && hit {
			return stored.CheckoutID, stored.OrderID, nil
		}
	}
	// Fall back to the workflow instance ID itself.
	return idemKey, "", nil
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
			}
		}
	}
	return result, nil
}

// workflowStateString maps a Dapr workflow runtime status to a CheckoutResult
// state string.
func workflowStateString(meta *workflow.WorkflowMetadata) string {
	if meta == nil {
		return "UNKNOWN"
	}
	return meta.String()
}

// unwrapAll walks the error chain to find the deepest wrapped error, which is
// what status.FromError inspects for gRPC status details.
func unwrapAll(err error) error {
	for {
		unwrapped := unwrapOne(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

type unwrapper interface{ Unwrap() error }

func unwrapOne(err error) error {
	if u, ok := err.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}

// ErrCheckoutDuplicateKey is returned when an idempotency key is reused with a
// different user_id, or when a key is presented after the workflow has been
// purged (use a new key).
var ErrCheckoutDuplicateKey = errCheckoutDuplicateKey{}

type errCheckoutDuplicateKey struct{}

func (e errCheckoutDuplicateKey) Error() string {
	return "checkout duplicate key"
}
