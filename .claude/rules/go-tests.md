---
paths:
  - "**/*_test.go"
---

# Go Test Conventions

Applies to all Go tests in this repo. CLAUDE.md sets the TDD workflow; this rule covers test *shape*.

## Package

Use the **external test package** (`package <pkg>_test`), not `package <pkg>`. Tests exercise the public API only.

```go
package biz_test
```

## Imports

- `github.com/stretchr/testify/assert` — non-fatal assertions
- `github.com/stretchr/testify/require` — fatal assertions; use when later lines would panic if the check fails (nil deref, index out of range)

Pick one per check site. Don't double-up (`require.NoError` then `assert.NotNil` on the same value is fine; `assert.NoError` then deref is a bug).

## AAA structure

Every test (and every `t.Run` subtest) must have these three comments, in order, with a blank line between sections:

```go
func TestThing_scenario_expected(t *testing.T) {
    // Arrange
    repo := newStubOrderRepo()
    uc := biz.NewOrderUsecase(repo, ...)

    // Act
    got, err := uc.Create(ctx, input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "PENDING", got.Status)
}
```

If a section is genuinely empty (e.g. no Arrange for a pure constructor test), still write the comment — it documents intent.

## Naming

`TestSubject_Scenario_expected` — three parts, underscore-separated.

```go
TestCheckout_DuplicateSameUser_returnsCachedCheckout
TestOrder_CancelAfterPaid_returnsErrOrderCannotCancel
```

The subject is the function or method under test, not the file. Scenario describes inputs/state. Expected describes the observable outcome.

## Doubles

Use **hand-rolled stubs**, not a mock framework. Stubs live in the same `_test.go` file (or a shared `testing_helpers_test.go` in the package) and have names like `stubOrderRepo`, `stubIdempotencyKeyRepo`. Match the repo interface exactly; return canned data or record calls on a struct field.

Don't introduce `gomock`, `mockery`, or `testify/mock` — none are used in this repo and they add maintenance cost (regen on interface change).

## Table-driven tests

Use when the same Act+Assert applies to many inputs. Keep the case struct small:

```go
cases := []struct {
    name     string
    status   string
    wantErr  error
}{
    {"pending cancels", "PENDING", nil},
    {"paid rejects", "PAID", biz.ErrOrderCannotCancel},
}
for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {
        // Arrange / Act / Assert (still required inside subtest)
    })
}
```

Skip the table when each case needs different setup — separate `Test*` funcs are clearer.

## What NOT to test

Per CLAUDE.md "Test behavior, not implementation". **Don't add maintenance-tax tests.** A test that passes today but breaks tomorrow on every unrelated refactor is a liability, not coverage.

Concretely, don't:

- Re-assert framework behavior (kratos transport, ent query mechanics, proto marshalling, Wire DI)
- Test private functions through reflection or by lowering visibility
- Pin the *shape* of an internal call graph (e.g. "repo.Get was called exactly twice with these args") when only the outcome matters
- Snapshot/golden-file structured data that changes on every legitimate edit
- Write exhaustive mocks of every method on every dependency — only stub what the test path actually invokes
- Mock the database for integration coverage of repo implementations — use a real Postgres (testcontainers) or skip the test
- Duplicate one assertion across many near-identical tests; collapse into a table instead

Before adding a test, ask: *what behavior change would make this test fail?* If the answer is "any refactor in this area," delete the test.

## Errors

- Sentinel errors: `assert.ErrorIs(t, err, biz.ErrOrderNotFound)` — not string compare
- gRPC status: `status.Code(err)` for the code, `status.Convert(err).Message()` for the message
- Wrapped kratos errors: assert on `errors.FromError(err).Reason` against `v1.ErrorReason_XXX.String()`

## Context

Tests that take a `context.Context` should use `context.Background()` unless testing cancellation/deadline behavior. Don't pass `nil`.

## Parallelism

Add `t.Parallel()` at the top of any test that doesn't share mutable state (most stub-based tests qualify). Inside table-driven `t.Run`, capture the loop var (`tc := tc`) before calling `t.Parallel()` — Go ≥1.22 makes this unnecessary, and this repo is on a recent Go, so omit the capture.
