# Saga Operations Runbook

## Quick triage flow

```
Alert fires
    │
    ├─ saga_orphan_payment_total > 0  ──► Scenario 1: Orphan payment
    │
    ├─ workflow RUNNING > 10 min       ──► Scenario 2: Stuck workflow
    │
    ├─ "failed to schedule workflow"   ──► Scenario 3: Workflowstore unreachable
    │   in order service logs
    │
    ├─ outbox_pending_gauge > 100      ──► Scenario 4: Outbox backed up
    │   for > 5 min
    │
    └─ Postgres disk > 75%             ──► Scenario 5: Disk pressure on workflowstore
```

## Scenarios

### 1. Orphan payment (saga_orphan_payment_total > 0)

**Symptom:** `saga_orphan_payment_total` counter is non-zero. A `payment.completed` or `payment.failed` event arrived for a workflow instance that was already terminated (timed out, cancelled, or purged). The event was written to the DLQ instead of being applied.

**Diagnose:**
```sql
SELECT id, topic, workflow_instance_id, reason, payload_json, created_at
FROM workflow_dead_letter_events
WHERE created_at > now() - interval '1h'
ORDER BY created_at DESC;
```

**Fix:** Decide per row.

- If the payment is COMPLETED but the workflow is gone → the order was likely cancelled before payment landed. Manually issue a refund:
  ```bash
  curl -X POST http://localhost:8001/v1/payments/<payment_id>/refund \
    -H 'Content-Type: application/json' \
    -d '{}'
  ```
- If the payment is FAILED → no action needed on the order side (the order was already cancelled during saga compensation). Clean up the DLQ row:
  ```sql
  DELETE FROM workflow_dead_letter_events WHERE id = '<row_id>';
  ```

---

### 2. Workflow stuck RUNNING > 10 min

**Symptom:** `GetCheckoutStatus` returns `state=RUNNING` for more than 10 minutes with no progression. Expected happy-path duration is under 10 seconds.

**Diagnose:**
```bash
dapr workflow get -i <instance_id>
```

Review the history output for the last event and timestamp to determine where the workflow is blocked.

**Fix options:**

1. **Manually raise the missing event** (preferred — unsticks without losing the saga):
   ```bash
   dapr workflow raise-event -i <instance_id> \
     -e payment-result-1 \
     -d '{"success":true,"payment_id":"<pay_id>"}'
   ```
   Adjust the event name (`payment-result-1`, `-2`, `-3`) to match the current attempt.

2. **Terminate as last resort** (saga compensates on next check):
   ```bash
   dapr workflow terminate -i <instance_id>
   ```
   After termination, check whether a COMPLETED payment already exists for the order. If so, follow Scenario 1 to issue a refund.

---

### 3. Workflowstore unreachable

**Symptom:** Order service logs repeat `"failed to schedule workflow"` in a tight loop. New `Checkout` calls return errors. Existing workflows cannot progress.

**Mitigation — immediate (flip to manual mode):**
```bash
kubectl edit configmap order-config
# Set: saga.enabled: "false"

kubectl rollout restart deployment/order
```

New `Checkout` calls will return `501 Unimplemented`. Callers must use `CreateOrder` (manual mode) until the workflowstore is restored.

**Recovery:**
1. Restore workflowstore connectivity (check `kubectl get pods -n go-mall`, `kubectl describe component workflowstore -n go-mall`).
2. Once healthy, flip `saga.enabled: "true"` and `kubectl rollout restart deployment/order`.
3. In-flight workflows that were interrupted will resume (Dapr Workflow is durable; they pick up from the last committed checkpoint).

---

### 4. Outbox backed up (outbox_pending_gauge > 100 for > 5 min)

**Symptom:** `outbox_pending_gauge` stays above 100 for more than 5 minutes. Payment requests are delayed, which can cause saga per-attempt timeouts to fire before the payment service processes the request — leading to spurious retries or cancellations.

**Diagnose:**
```sql
SELECT topic, count(*)
FROM outbox_messages
WHERE sent_at IS NULL
GROUP BY topic
ORDER BY count DESC;
```

**Fix:**
```bash
kubectl rollout restart deployment/order
```

The outbox relay restarts and resumes processing from the last unpublished row. Rows are not lost — they remain in Postgres until `sent_at` is set.

If the backlog does not clear within 2 minutes after restart, check the Dapr sidecar logs for pubsub errors:
```bash
kubectl logs deployment/order -c daprd -n go-mall --tail=100
```

---

### 5. Disk pressure on workflowstore

**Symptom:** Postgres disk usage on the workflowstore database exceeds 75%.

**Mitigation:**

1. Confirm the PurgeWorkflow cron is running. Check the `completed_workflows` table for recent `purged_at` timestamps:
   ```sql
   SELECT max(purged_at) FROM completed_workflows;
   ```
   If `purged_at` is stale (> 6h ago), the cron may be stuck — restart the order service to reinitialize it.

2. If purge is healthy but disk still grows, lower the workflow history retention window. Edit `deploy/k8s/base/infra/dapr/workflow-config.yaml`:
   ```yaml
   spec:
     workflow:
       stateRetentionPolicy:
         completed: "72h"   # was 168h
         failed: "720h"
         terminated: "72h"  # was 168h
   ```
   Apply and restart:
   ```bash
   kubectl apply -f deploy/k8s/base/infra/dapr/workflow-config.yaml
   kubectl rollout restart deployment/order
   ```

3. At 85% disk → roll back to manual mode (Scenario 3 procedure) to stop new workflow state accumulation immediately.

---

## Capacity model

```
workflow_history_row_size_bytes × peak_checkouts_per_day × retention_days × 3 (headroom)
= required Postgres disk for workflowstore
```

Example: 30 KB × 200,000 × 7 × 3 = ~125 GiB required.

| Threshold | Action |
|-----------|--------|
| 60% | Warn (Slack alert) |
| 75% | Page on-call; investigate PurgeWorkflow health |
| 85% | Rollback to manual mode (flip `saga.enabled=false`) |

Review capacity before any traffic ramp. Recalculate if `peak_checkouts_per_day` increases by more than 2×.

---

## Canary rollout

Order saga is gated by `saga.enabled` in `app/order/configs/config.yaml`.

### Step 1 — Payment service first

Deploy payment with the new `CompletePayment` and `FailPayment` RPCs and the `workflow_instance_id` / `attempt` ent fields. These changes are additive: the new column is nullable, the new RPCs are idle until the order service subscribes to `payment.requested`. No traffic shift required.

Verify:
```bash
make build   # from repo root
make test    # from app/payment/
```

### Step 2 — Order with saga.enabled=false

Deploy order service with saga code present but `saga.enabled=false`. Verify:
```bash
make build   # from repo root
make test    # from app/order/
```

Confirm that the `payment.requested` subscriber on the payment side and the `payment.completed` / `payment.failed` subscribers on the order side start without errors (check pod logs). No customer-facing change — `Checkout` returns `501 Unimplemented`.

### Step 3 — 5% canary

Flip `saga.enabled=true` in a 5% canary cohort (via ingress weight or feature flag in config). Watch for 24 hours:

| Metric | Expected |
|--------|----------|
| `saga_compensations_total` | Low — single digits per hour at low traffic |
| `saga_orphan_payment_total` | Zero — any non-zero value is a hard alert |
| `saga_duration_seconds` p99 | < 10s for happy path |

### Step 4 — Ramp

25% → 50% → 100%, with 24-hour soak between each step. At each ramp, re-check the three metrics above before proceeding. If any metric regresses, hold the ramp and investigate.

---

## Rollback

Single config flip:

```bash
kubectl edit configmap order-config
# Set: saga.enabled: "false"

kubectl rollout restart deployment/order
```

New `Checkout` calls return `501 Unimplemented` until the flag is restored. In-flight workflows continue to drive to completion (best-effort) — the Dapr Workflow runtime does not abort running instances on config change.
