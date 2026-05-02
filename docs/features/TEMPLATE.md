---
id: PX.Y
title: Short feature title
phase: PX
status: not-started        # not-started | in-progress | in-review | done | archived
priority: should           # must | should | could | wont
estimate: TBD              # XS | S | M | L | XL
owner: TBD
depends_on: []             # list of feature IDs, e.g. [P0.1, P0.5]
emits: []                  # event topic names this feature publishes
consumes: []               # event topic names this feature subscribes to
services_touched: []       # e.g. [catalog, cart, order]
---

# PX.Y — Short feature title

## Problem

What's broken or missing today? One paragraph. No solution yet.

## User stories

- As a **{persona}**, I want **{capability}**, so that **{outcome}**.
- As a **{persona}**, I want **{capability}**, so that **{outcome}**.

## Acceptance criteria

Write each as a checkable, testable statement. No vague terms — replace "fast" with a number, "secure" with a specific guarantee.

- [ ] …
- [ ] …
- [ ] …

## Out of scope

What this ticket explicitly does **not** cover. Forces clarity, prevents scope creep.

- …

## Implementation notes

### Files / packages
- `path/to/file.go`
- …

### Pattern reference
Existing service or `pkg/*` to model after.

### Schema changes
ent schema additions, new tables, indexes.

### Events
- **Emits:** `topic.name` — payload schema link or inline struct.
- **Consumes:** `topic.name` — handler location, idempotency key.

### API surface
- New RPCs, HTTP routes, request/response shape.

## Risks & mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| …    | low/med/hi| low/med/hi | … |

## Test plan

- **Unit:** behaviour-level tests in `…_test.go`.
- **Integration:** real DB / real broker via testcontainers; cover the happy path + 1 failure case per AC.
- **E2E:** if user-visible, exercise from `app/web` Playwright suite.
- **Manual smoke:** steps to verify in `tilt up`.

## Open questions

- …

## Notes

Anything else: links to wiki refs (`docs/saga.md`, `docs/ecommerce-microservices-pitfalls.md`), prior art, alternatives considered.
