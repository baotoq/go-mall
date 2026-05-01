# Code Review — go-mall

Date: 2026-05-01
Reviewers: 3 worker agents (architecture / bugs / security) + 1 lead (consolidation & verification)
Scope: full repo (`app/`, `api/`, `deploy/`)
Method: each worker scanned independently; lead deduped, opened cited files, and dropped findings that did not hold up against the code. Only verified issues are listed below.

Severity legend: **Critical** = production correctness/security risk; **High** = will hit in practice; **Medium** = real issue, lower blast radius.

---

## Critical

### C1. Cart `AddItem` / `FindOrCreateBySession` — TOCTOU race against unique constraints

`app/cart/internal/data/cart.go:24-42` and `app/cart/internal/data/cart.go:44-96`.

`FindOrCreateBySession` reads the cart, and on `IsNotFound` calls `Cart.Create()`. The schema has `field.String("session_id").Unique()` (`app/cart/internal/data/ent/schema/cart.go:17`). Two concurrent first-add requests for the same session both miss, both `Save`, and the second fails with a unique-constraint error returned to the user.

`AddItem` repeats the pattern at item granularity: `CartItem.Query()...Only(ctx)` followed by `Create()` or `UpdateOne()` (lines 62-90). The schema declares `index.Fields("cart_id", "product_id").Unique()` (`app/cart/internal/data/ent/schema/cart_item.go:38`). Concurrent adds of the same product land in the create branch and the second insert violates the unique index — instead of incrementing quantity.

**Fix:** use `ent`'s `OnConflict` upsert for `CartItem`, and either upsert or `Tx` + advisory-lock the cart create. At minimum, retry once on unique-violation.

---

### C2. Plaintext DB password and `sslmode=disable` in committed manifest

`deploy/k8s/base/secrets/secret.yaml:7`

```
DATABASE_CONNECTION_STRING: "host=postgres port=5432 user=greeter password=greeter dbname=greeter sslmode=disable"
```

This is a `Secret` checked into the repo with the production-shaped DSN. Two problems:

1. The password lives in git history forever; rotating it requires editing the repo.
2. `sslmode=disable` means pod→Postgres traffic is plaintext. The CLAUDE.md flow is "secrets are injected at runtime by Dapr secretstore" (`CLAUDE.md:64`) — the in-tree manifest contradicts that contract.

**Fix:** remove the literal DSN from the manifest; let Dapr's `secretstore` deliver it as documented. Set `sslmode=require` (or `verify-full`) in whatever DSN replaces it.

---

### C3. Payment endpoints have no authentication

`app/payment/internal/server/http.go:15-36` (only `recovery.Recovery()` middleware) and `app/payment/internal/server/grpc.go` (same).

Combined with `RefundPayment(id)` (`api/payment/v1/payment.proto:22-27`) and the biz layer that takes only `id` (`app/payment/internal/biz/payment.go:63-72`), any caller who can reach the service can refund any payment by guessing/enumerating IDs. Same for `GetPayment`, `ListPayments`, and `CreatePayment`.

**Fix:** add a kratos auth middleware (JWT or selector-based) and check `user_id` ownership inside `Refund`/`GetByID` before mutating.

---

## High

### H1. Order `MarkPaid` is two non-atomic writes

`app/order/internal/biz/order.go:113-128`

```go
if _, err := uc.repo.SetPaymentID(ctx, id, paymentID); err != nil { return nil, err }
return uc.repo.UpdateStatus(ctx, id, "PAID")
```

If the second call fails (DB blip, deploy mid-flight, context cancel), the row has `payment_id` set but `status="PENDING"`. Reconciliation queries that look for `status="PAID"` will miss the transaction; retries will re-call `SetPaymentID` with the same value but the status update may already have succeeded under contention, leaving an unclear final state.

**Fix:** combine into a single `UPDATE orders SET payment_id=?, status='PAID' WHERE id=? AND status NOT IN ('PAID','CANCELLED')` — atomic, idempotent, and guards the existing status precondition at lines 118-123 against TOCTOU.

---

### H2. cart/order/payment ship only `Dockerfile.dev.debug`

```
app/cart/Dockerfile.dev.debug
app/order/Dockerfile.dev.debug
app/payment/Dockerfile.dev.debug
```

Only `app/greeter/Dockerfile` and `app/catalog/Dockerfile` exist; the other three services have no production image build. `Dockerfile.dev.debug` embeds Delve and exposes `2345` (`app/payment/Dockerfile.dev.debug:14,19`). If deploy tooling defaults to "the only Dockerfile present," prod could ship a debug image with a remote debugger.

**Fix:** copy `app/greeter/Dockerfile` for the three missing services and ensure `deploy/` references the non-debug variant explicitly.

---

### H3. Containers run as root, no resource limits

Pod specs (`deploy/k8s/base/app/payment.yaml:30-55`, plus `cart.yaml`, `order.yaml`, `greeter.yaml`, `catalog.yaml`) declare no `securityContext` and no `resources`. Dockerfiles (`app/greeter/Dockerfile:16-32`) have no `USER` directive, so the container default UID 0 stands.

A compromised app process therefore runs as root in the container, and a misbehaving process has no memory/CPU ceiling — a single OOM-looping pod can evict neighbours.

**Fix:** add `USER nonroot:nonroot` (or numeric UID) to each Dockerfile; in each Deployment add

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
resources:
  requests: { cpu: 100m, memory: 128Mi }
  limits:   { cpu: 500m, memory: 512Mi }
```

---

### H4. Payment proto has no validation on monetary fields

`api/payment/v1/payment.proto:50-56`

```proto
message CreatePaymentRequest {
  string order_id = 1;
  string user_id = 2;
  int64 amount_cents = 3;
  string currency = 4;
  string provider = 5;
}
```

No `buf.validate` constraints. `PaymentUsecase.Create` (`app/payment/internal/biz/payment.go:46-49`) just sets status and persists, so negative `amount_cents`, `0`, empty `currency`, and empty `order_id` all succeed. A negative-amount payment that later gets refunded is a money bug.

**Fix:** add `(buf.validate.field).int64.gt = 0` on `amount_cents`, `len_bytes` constraint on `currency` (3), and `string.uuid`/non-empty on `order_id`/`user_id`. The proto pipeline already vendors the validate import (`third_party/`).

---

## Medium

### M1. Inconsistent proto path layout

`api/greeter/helloworld/v1/` vs `api/cart/v1/`, `api/catalog/v1/`, `api/order/v1/`, `api/payment/v1/`.

CLAUDE.md describes the convention as `api/<app>/<domain>/v<N>/<name>.proto` (`CLAUDE.md:93`). Greeter follows it; the four newer services drop the `<domain>` segment. Either flatten greeter or reintroduce a domain segment in the others — pick one and document it. Cosmetic, but it bites every new service that copy-pastes from a different template.

### M2. RBAC over-broad on secrets

`deploy/k8s/base/app/rbac.yaml:1-21` grants `get`/`list` on **all** secrets in the namespace to the `default` ServiceAccount. With Dapr already brokering secret access, the application pods should not also have direct API-server read on every secret.

**Fix:** drop the `Role` entirely (rely on Dapr) or scope `resourceNames` to the specific secret(s) used.

### M3. Schema migration uses `context.Background()` with no timeout

`app/cart/internal/data/data.go:25`, plus identical lines in `app/{order,payment,catalog,greeter}/internal/data/data.go`.

```go
if err := client.Schema.Create(context.Background()); err != nil { ... }
```

If Postgres is reachable but slow (e.g., network partition during rollout), the pod hangs in `NewData` indefinitely instead of failing the readiness probe and letting Kubernetes restart it.

**Fix:** wrap with `context.WithTimeout(ctx, 30*time.Second)` and surface the error; the pod will crash-loop with a clear cause instead of silently stalling.

---

## Not flagged (false positives / out-of-scope)

For transparency, the following worker findings were dropped after the lead pass:

- *cart/order/payment missing `NewDaprClient` in `ProviderSet`* — those services don't publish events; main.go retrieves secrets via a local Dapr client (`app/cart/cmd/server/main.go:71-95`). Not a bug.
- *Double-cancel race in main.go retry loop* — re-read of the loop shows `cancel()` is paired correctly on both branches.
- *Nil deref on `c.Edges.Items`* — query at `cart.go:25-28` uses `WithItems()`; ent guarantees a non-nil edge slice on hit and the create branch initialises it explicitly at `cart.go:39`.
- *Dapr secretstore unscoped* — overlaps with M2.
- *Empty default DSN in ConfigMap* — intentional; DSN must come from Dapr.

---

## Suggested fix order

1. **C2** (rotate the leaked password before anything else)
2. **C3** + **H4** (payment auth and validation — money path)
3. **C1** + **H1** (correctness under concurrency)
4. **H2** + **H3** (image / pod hardening before next prod deploy)
5. **M1**–**M3** (cleanup pass)
