# Feature Specs

One file per upcoming feature from `docs/ROADMAP.md`. Each file is a self-contained user-story-driven spec: stakeholder needs, acceptance criteria, implementation surface, risks, test plan, out-of-scope.

## How to use

1. Pick a feature from the table below.
2. Open its file. The frontmatter holds metadata (status, priority, owner, dependencies). The body holds the user story breakdown and engineering detail.
3. As work begins, update `status` (`not-started` → `in-progress` → `in-review` → `done`) and assign `owner`.
4. Anything not covered by the template, add as a section under `## Notes` at the end.

New features: copy `TEMPLATE.md`, rename to `P{phase}.{n}-{slug}.md`, add a row to the table below.

## Index

### Phase 0 — Correctness foundations
| ID | Feature | Status | Priority |
|----|---------|--------|----------|
| [P0.1](P0.1-auth-cart-order.md) | Auth on cart and order | not-started | must |
| [P0.2](P0.2-order-price-validation.md) | Order → Catalog price validation | not-started | must |
| [P0.3](P0.3-stock-reservation.md) | Stock reservation via `order.created` | not-started | must |
| [P0.4](P0.4-order-saga.md) | Order saga: react to stock + payment events | not-started | must |
| [P0.5](P0.5-dapr-pubsub-wiring.md) | Wire Dapr pubsub end-to-end | not-started | must |

### Phase 1 — Domain expansion
| ID | Feature | Status | Priority |
|----|---------|--------|----------|
| [P1.1](P1.1-identity.md) | `identity` service (users, addresses) | not-started | must |
| [P1.2](P1.2-payment-gateway.md) | Payment gateway adapter (Stripe) | not-started | must |
| [P1.3](P1.3-shipping.md) | `shipping` service | not-started | should |
| [P1.4](P1.4-tax-pricing.md) | `tax` / pricing engine | not-started | should |
| [P1.5](P1.5-notifications.md) | `notification` service | not-started | should |
| [P1.6](P1.6-inventory-split.md) | `inventory` service split | not-started | could |

### Phase 2 — Commerce growth
| ID | Feature | Status | Priority |
|----|---------|--------|----------|
| [P2.1](P2.1-search.md) | Search service | not-started | should |
| [P2.2](P2.2-promotions.md) | Promotions / coupons | not-started | should |
| [P2.3](P2.3-reviews.md) | Reviews & ratings | not-started | should |
| [P2.4](P2.4-wishlist.md) | Wishlist / saved-for-later | not-started | could |
| [P2.5](P2.5-recommendations.md) | Recommendations | not-started | could |
| [P2.6](P2.6-returns-rma.md) | Returns / RMA | not-started | should |

### Phase 3 — Operational maturity
| ID | Feature | Status | Priority |
|----|---------|--------|----------|
| [P3.1](P3.1-observability.md) | Observability stack | not-started | must |
| [P3.2](P3.2-api-gateway.md) | API gateway / BFF | not-started | should |
| [P3.3](P3.3-admin-app.md) | Admin app | not-started | should |
| [P3.4](P3.4-reporting.md) | Reporting / analytics read-model | not-started | could |
| [P3.5](P3.5-audit-log.md) | Audit log | not-started | should |

### Phase 4 — Platform & scale
| ID | Feature | Status | Priority |
|----|---------|--------|----------|
| [P4.1](P4.1-multi-currency.md) | Multi-currency / locale | not-started | could |
| [P4.2](P4.2-feature-flags.md) | Feature flags | not-started | should |
| [P4.3](P4.3-multi-tenant.md) | Tenant isolation | not-started | wont (yet) |
| [P4.4](P4.4-disaster-recovery.md) | Disaster recovery | not-started | should |
| [P4.5](P4.5-performance.md) | Performance pass | not-started | should |

## Conventions

- **Priority:** `must` (P0/blocker), `should` (next quarter), `could` (nice-to-have), `wont` (explicitly deferred).
- **Status:** `not-started` → `in-progress` → `in-review` → `done` → `archived`.
- **Estimate:** rough t-shirt size — `XS` (<1d), `S` (1-3d), `M` (1-2w), `L` (2-4w), `XL` (1m+).
- **IDs are stable.** When a feature splits, append a letter (`P1.2a`, `P1.2b`); when it's deleted, mark status `archived` and keep the file.
