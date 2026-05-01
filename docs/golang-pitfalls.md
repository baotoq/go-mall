# Go Pitfalls — A Deep Reference

A curated catalog of the traps Go developers fall into, organized by topic. Each entry has a short problem statement, a minimal example, and the fix. Examples assume Go 1.22+ unless noted.

---

## 1. The `for` Loop Variable Capture (the famous one)

**Problem (pre-Go 1.22):** loop variables were shared across iterations. Closures and goroutines captured the *variable*, not the value, so they almost always saw the last value.

```go
// BUG (pre-1.22): prints "3, 3, 3" — not "0, 1, 2"
for i := 0; i < 3; i++ {
    go func() { fmt.Println(i) }()
}
```

**Fix:**
- **Go 1.22+** (with `go 1.22` in `go.mod`): each iteration gets its own `i`. Just works.
- **Pre-1.22:** shadow the variable: `i := i` inside the loop, or pass it as an argument: `go func(i int) {...}(i)`.

**Still relevant in 2026 because:** modules pinned to `go 1.21` or earlier keep the old semantics. Audit the `go` directive in `go.mod` before assuming you're safe. See [Fixing For Loops in Go 1.22](https://go.dev/blog/loopvar-preview).

---

## 2. Nil Interface ≠ Nil Pointer

**Problem:** an interface value has *two* slots — type and value. It is `nil` only when **both** are nil. A typed nil pointer wrapped in an interface is not equal to nil.

```go
func returnsError() error {
    var p *MyError = nil
    return p // type=*MyError, value=nil
}

err := returnsError()
fmt.Println(err == nil) // false! 😱
```

**Fix:**
- Return the literal `nil` from functions whose return type is an interface, not a typed nil pointer.
- If you must wrap: `if p == nil { return nil }` before returning.
- Lint with `nilness` / `nilerr` / `errcheck`.

See [yourbasic: Why nil error is not equal to nil](https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/).

---

## 3. Slice Sharing & Append Aliasing

**Problem:** slicing does **not** copy. Two slices can share a backing array. `append` may write into the backing array of the original if capacity allows, silently corrupting it.

```go
s := []int{1, 2, 3, 4, 5}
a := s[:2]            // len=2, cap=5 — shares backing array
b := append(a, 99)    // writes 99 at s[2], clobbering original!
fmt.Println(s)        // [1 2 99 4 5]
```

**Fix:**
- Use **full slice expressions** to cap capacity: `a := s[:2:2]` forces append to allocate.
- Defensive copy with `slices.Clone(...)` (Go 1.21+) when handing slices to other code.
- Treat any slice returned from a parser, decoder, or `bytes.Buffer.Bytes()` as **borrowed** — copy if you keep it.

See [Go slice gotchas](https://rednafi.com/go/slice-gotchas/).

---

## 4. Concurrent Map Access Panics

**Problem:** Go's built-in map is **not** safe for concurrent reads + writes. A concurrent write triggers `fatal error: concurrent map writes` — uncatchable, kills the process.

**Fix:**
- Wrap with `sync.RWMutex` (most common).
- Use `sync.Map` only when access pattern is *write-once-read-many* or disjoint key sets per goroutine. It's slower than RWMutex for general workloads.
- `sync.Map` stores `any` — boxing has cost. Don't reach for it reflexively.

See [Data Race Patterns in Go (Uber)](https://www.uber.com/us/en/blog/data-race-patterns-in-go/).

---

## 5. Goroutine Leaks

The four canonical leak patterns:

### 5a. Forgotten Receiver
```go
ch := make(chan int)        // unbuffered
go func() { ch <- compute() }()
// caller hits an early return — sender blocks forever
```

### 5b. Forgotten Sender
```go
ch := make(chan int)
go func() { v := <-ch; ... }()
// nobody ever sends, channel never closed
```

### 5c. `select` with no exit
```go
for {
    select {
    case v := <-ch:
        ...
    } // no <-ctx.Done(), no default — leaks if ch never delivers
}
```

### 5d. WaitGroup misuse
`wg.Add` after `wg.Wait` returns, or forgetting `wg.Done` on an error path.

**Fix:**
- Every long-lived goroutine must accept a `context.Context` and select on `ctx.Done()`.
- Use buffered channels of size 1 for "fire and forget" results so the sender can never block.
- In tests, use `goleak` (Uber) or Go 1.25+ `testing/synctest` to assert no leaks.
- Go 1.26 ships an experimental `goroutineleak` pprof profile (`GOEXPERIMENT=goroutineleakprofile`).

See [Goroutine Leaks: The Forgotten Sender (Ardan Labs)](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html) and [Detecting goroutine leaks with synctest/pprof](https://antonz.org/detecting-goroutine-leaks/).

---

## 6. Send on Closed Channel Panics

**Problem:** sending on a closed channel **always** panics. Closing a closed channel also panics. Closing a `nil` channel panics.

**Rules of thumb (Go 101's "channel closing principle"):**
1. Don't close a channel from the receiver side.
2. Don't close a channel that has multiple senders.
3. The single sender owns the close.

**Multi-sender safe shutdown:** use a separate `done`/`stop` channel that signals senders to stop. The senders exit, and only after they're done does someone close the data channel (or just let it be garbage-collected).

See [How to Gracefully Close Channels (Go 101)](https://go101.org/article/channel-closing.html).

---

## 7. `defer` in a Loop

**Problem:** `defer` fires at **function** return, not loop-iteration end. In a long loop, each iteration stacks up another deferred call, holding file handles / DB connections / mutexes open.

```go
for _, name := range files {
    f, _ := os.Open(name)
    defer f.Close() // ❌ all files stay open until function exits
    // ... process f ...
}
```

**Fix:** extract the body into a helper function so each iteration has its own scope:
```go
for _, name := range files {
    func() {
        f, _ := os.Open(name)
        defer f.Close()
        // ...
    }()
}
```
Or just call `f.Close()` directly at the end of each iteration when control flow allows.

---

## 8. HTTP Response Body Pitfalls

**Problem 1 — order:** deferring `resp.Body.Close()` before checking the error panics on transport errors when `resp` is nil.
```go
resp, err := http.Get(url)
defer resp.Body.Close() // ❌ panics if err != nil
if err != nil { return err }
```

**Fix:** check `err` first, then defer close.

**Problem 2 — must always close:** even if you don't read the body. Otherwise the underlying TCP connection is leaked into `CLOSE_WAIT` and you exhaust file descriptors.

**Problem 3 — connection reuse:** if you don't fully **drain** the body, the connection can't be reused from the keep-alive pool. Use `io.Copy(io.Discard, resp.Body)` before closing.

**Problem 4 — default client has no timeout:** `http.DefaultClient` will wait forever. Build your own:
```go
client := &http.Client{Timeout: 10 * time.Second}
```
For finer control, use `context.WithTimeout` and `http.NewRequestWithContext`.

See [Don't use Go's default HTTP client (in production)](https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779).

---

## 9. `context` Misuse

**Common pitfalls:**
- Forgetting to `defer cancel()` after `WithTimeout`/`WithCancel`/`WithDeadline` — the context (and any timer) leaks.
- Storing `context.Context` in a struct. Don't — pass it as the first parameter to functions.
- Stuffing request-scoped values via `context.WithValue` for things that should be plain function parameters.
- Using `context.Background()` deep in handler code instead of the request's context.
- Ignoring `<-ctx.Done()` in long-running loops; the context expires and your work continues uselessly.

See [pkg.go.dev: context](https://pkg.go.dev/context).

---

## 10. Error Wrapping & `errors.Is` / `errors.As`

**Pitfalls:**
- Comparing wrapped errors with `==` instead of `errors.Is`.
- Using a type assertion (`err.(*MyError)`) instead of `errors.As`.
- Passing a value (not a pointer) to `errors.As` → **panics**.
- Wrapping internal errors and exposing them at the API boundary, leaking implementation detail.
- Using `%v` instead of `%w` in `fmt.Errorf` — produces a string, drops the chain.
- Wrapping the **same** error multiple times across layers, producing a noisy error message.

```go
// ✅ wrap with %w
return fmt.Errorf("loading config: %w", err)

// ✅ check with errors.Is
if errors.Is(err, os.ErrNotExist) { ... }

// ✅ extract typed error with errors.As — note the &
var pathErr *os.PathError
if errors.As(err, &pathErr) { ... }
```

See [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors).

---

## 11. Pointer Receivers vs. Value Receivers

**Pitfalls:**
- Mixing receiver kinds on the same type — confusing and a common source of "type X does not implement interface Y" surprises.
- A method with a `*T` receiver is **not** in `T`'s method set. So `T` (the value) does **not** satisfy an interface that requires that method. Only `*T` does.
- Map values and interface values are not addressable, so you can't call pointer-receiver methods on them directly.

**Rules of thumb:**
- If any method needs a pointer receiver, **all** of them should use pointer receivers (consistency).
- Use pointer receivers for large structs (avoid copies) or when mutating state.
- Use value receivers for small, immutable, comparable types (e.g., `time.Time`).

---

## 12. `time.Time` Equality and the Monotonic Clock

**Problem:** `time.Time` carries both a wall-clock and monotonic reading. `t1 == t2` compares both — plus the `Location`. Round-tripping through JSON drops the monotonic part, so `before == after` becomes false.

**Rules:**
- Never use `==` on `time.Time` — use `t1.Equal(t2)`.
- Don't use `time.Time` as a map key without first calling `t.Round(0)` to strip the monotonic reading and `.UTC()` to normalize location.
- Use the monotonic clock for *durations* (`time.Since`, `time.Until`) — it's immune to NTP/DST jumps.
- The monotonic clock may stop while the machine sleeps; `Sub` may underreport elapsed time.

See [pkg.go.dev: time](https://pkg.go.dev/time).

---

## 13. JSON Encoding/Decoding Surprises

- **Numbers become `float64`** when decoded into `any`/`map[string]any`. Integers larger than 2^53 lose precision. Use `decoder.UseNumber()` and `json.Number`, or unmarshal into a typed struct.
- **`omitempty` doesn't trigger on zero structs or zero `time.Time`** — they're not "empty." Go 1.24 added `omitzero` which does.
- **Unexported fields are silently ignored.** Marshaling skips them; unmarshaling doesn't populate them.
- **Unknown fields are silently dropped** by default. Use `decoder.DisallowUnknownFields()` to surface them.
- **Slices vs. arrays:** unmarshaling into `[5]int` requires exactly 5 elements; into `[]int` it's flexible.
- **Pointer vs. value:** `*int` distinguishes "missing" from "zero"; plain `int` cannot.
- Go 1.25 introduces `encoding/json/v2` (experimental) — addresses many of these. See [pkg.go.dev: encoding/json/v2](https://pkg.go.dev/encoding/json/v2).

See [Surprises and Gotchas When Working with JSON](https://www.alexedwards.net/blog/json-surprises-and-gotchas).

---

## 14. Struct Field Alignment

**Problem:** the compiler inserts padding to satisfy each field's alignment. Bad ordering inflates struct size and thrashes cache lines.

```go
// 24 bytes (8 wasted on padding)
type Bad struct {
    a bool   // 1
    _ [7]byte
    b int64  // 8
    c bool   // 1
    _ [7]byte
}

// 16 bytes — no waste
type Good struct {
    b int64
    a bool
    c bool
}
```

**Fix:** order fields **largest → smallest**. Run `go vet -vettool=$(which fieldalignment)` (from `golang.org/x/tools/go/analysis/passes/fieldalignment`).

When it matters: hot structs allocated in the millions, network packet structs, atomic fields (which require 8-byte alignment on 32-bit platforms — put them first).

---

## 15. Package-Level `init` and Globals

**Pitfalls:**
- File **lexical** order influences init order. Renaming a file can quietly change behavior.
- `init()` runs at import time, including in tests, and can poison test state.
- Variable initialization runs in dependency-respecting declaration order. Cycles cause compile errors; near-cycles produce surprises.
- Globals + `init` make code hard to test because there's no seam to inject test fakes.

**Fix:** prefer explicit constructors / DI; reserve `init` for `database/sql` driver registration and similar registry patterns.

---

## 16. `iota` Surprises

- `iota` resets per `const` block, not per file.
- Skipping an entry shifts all subsequent values.
- Using `iota` for values that get persisted (DB, wire) is fragile — reordering the block silently rewrites your storage schema. Prefer explicit integer constants for persisted enums.

---

## 17. Variable Shadowing with `:=`

```go
var err error
if cond {
    x, err := f() // ❌ shadows outer err
    _ = x
}
// outer err is still nil
```

**Fix:** use `=` when at least one variable on the LHS already exists, or run `go vet -shadow` (now in `go vet` via the analyzers in `golang.org/x/tools`).

---

## 18. Generics Pitfalls

- **No methods with type parameters** (as of Go 1.25). Common workaround: lift the method into a free function, or parameterize the receiver type instead.
- **Type sets don't expose common fields.** Even if every type in your union is a struct with field `.X`, you can't access `.X` through the constraint. Define a method-based interface instead.
- **Inference can fail** in subtle places (return-only type parameters, methods on generic types). When it does, the compile error is rough — be ready to write `Func[T](...)` explicitly.
- **`comparable` ≠ ordered.** Use `cmp.Ordered` (Go 1.21+) for `<`, `>`.
- **Don't reach for generics first.** Often `interface{}` + a small switch, or just code duplication, is clearer. See [100 Go Mistakes #9](https://100go.co/9-generics/).

---

## 19. `range` on a Map Has Random Order

Iteration order is intentionally randomized per Go spec. Code that depends on order — or even *appears* to depend on it under light testing — will surface bugs in production. Sort keys explicitly when order matters.

---

## 20. `string([]byte)` and `string(int)`

- `string(rune)` converts to UTF-8 — fine.
- `string(int)` converts an integer **to its Unicode code point**, not its decimal representation. `string(65)` is `"A"`, not `"65"`. `go vet` warns; use `strconv.Itoa`.

---

## 21. `bytes.Buffer` and `strings.Builder` Pitfalls

- `bytes.Buffer.Bytes()` returns a slice **aliased to the buffer's internal storage**. Subsequent writes can mutate it. Copy if you keep it.
- `strings.Builder` must not be copied after first use — copying breaks the underlying pointer aliasing it relies on for the zero-allocation `String()`.

---

## 22. `os/exec` and `Stdin`/`Stdout`

- Forgetting to call `cmd.Wait()` after `cmd.Start()` leaks the process and pipes.
- Reading from `cmd.StdoutPipe()` after `Wait()` returns no data — the pipe is already closed. Read first, *then* `Wait`.
- `exec.Command` does not run a shell; if you need globbing/pipes/redirection, invoke `sh -c "..."` explicitly (and beware injection).

---

## 23. Reflection and `unsafe`

- `reflect.DeepEqual` on `time.Time`, channels, `*sync.Mutex`, etc. behaves in surprising ways. For tests, prefer `cmp.Diff` (`github.com/google/go-cmp`) which lets you customize comparisons.
- `unsafe.Pointer` rules in `unsafe` package docs are **rules**, not suggestions. The garbage collector and the race detector can both break code that violates them.
- Reflection is ~50× slower than direct access; don't use it on hot paths.

---

## 24. `panic` / `recover` Misconceptions

- `recover` only works **inside a deferred function**, and only catches a panic in the same goroutine.
- A panic in a goroutine that nobody recovers from kills the entire process. Always wrap goroutines spawned from libraries with a top-level recover if a single bad worker shouldn't crash the host.
- `panic` is for unrecoverable programmer errors. Don't use it for control flow or expected error paths.

---

## 25. Build & Module Pitfalls

- `go.mod`'s `go` directive controls language semantics (e.g., loop var fix). Bumping it changes behavior — read the release notes.
- `replace` directives are not honored by transitive consumers of your module. They only work in the *main* module.
- `go install` of a package outside a module installs the latest version, not the version in your `go.sum`. Pin with `@version`.
- `vendor/` and `GOFLAGS=-mod=...` interact in subtle ways with CI. Pick a strategy and document it.

---

## Tools to Catch These

- `go vet` — built-in, free.
- `staticcheck` — sharper checks (`honnef.co/go/tools`).
- `golangci-lint` — meta-runner; enable `errcheck`, `govet`, `staticcheck`, `gosec`, `revive`, `bodyclose`, `contextcheck`, `nilerr`, `ineffassign`, `gocritic`, `unparam`.
- `fieldalignment` — struct ordering.
- `goleak` (Uber) / `testing/synctest` (Go 1.25+) — goroutine leak detection in tests.
- Race detector: `go test -race`, `go run -race`. Run in CI.

---

## Sources

- [Goroutine Leaks - The Forgotten Sender (Ardan Labs)](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html)
- [Goroutine Leaks in Go: The 4 Patterns and the New Profile in Go 1.26](https://dev.to/gabrielanhaia/goroutine-leaks-in-go-the-4-patterns-and-the-new-profile-in-go-126-5e73)
- [Detecting goroutine leaks with synctest/pprof](https://antonz.org/detecting-goroutine-leaks/)
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [Help: Nil is not nil — yourbasic](https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/)
- [Explaining nil interface{} gotcha in Go](https://blog.kowalczyk.info/a-d9qh/explaining-nil-interface-gotcha-in-go.html)
- [Fixing For Loops in Go 1.22](https://go.dev/blog/loopvar-preview)
- [Go Wiki: LoopvarExperiment](https://go.dev/wiki/LoopvarExperiment)
- [Go slice gotchas — Redowan's Reflections](https://rednafi.com/go/slice-gotchas/)
- [Go's append is not always thread safe](https://medium.com/@cep21/gos-append-is-not-always-thread-safe-a3034db7975)
- [Data Race Patterns in Go (Uber)](https://www.uber.com/us/en/blog/data-race-patterns-in-go/)
- [Defer Functions in Golang: Common Mistakes and Best Practices](https://rezakhademix.medium.com/defer-functions-in-golang-common-mistakes-and-best-practices-96eacdb551f0)
- [How to spot and fix memory leaks in Go (Datadog)](https://www.datadoghq.com/blog/go-memory-leaks/)
- [pkg.go.dev: context](https://pkg.go.dev/context)
- [How To Use Contexts in Go (DigitalOcean)](https://www.digitalocean.com/community/tutorials/how-to-use-contexts-in-go)
- [Concurrent map access in Go](https://medium.com/@luanrubensf/concurrent-map-access-in-go-a6a733c5ffd1)
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Don't just check errors, handle them gracefully — Dave Cheney](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
- [pkg.go.dev: errors](https://pkg.go.dev/errors)
- [Padding is hard — Dave Cheney](https://dave.cheney.net/2015/10/09/padding-is-hard)
- [Memory Layouts — Go 101](https://go101.org/article/memory-layout.html)
- [Struct Field Alignment — Go Optimization Guide](https://goperf.dev/01-common-patterns/fields-alignment/)
- [Monotonic and Wall Clock Time in the Go time package](https://victoriametrics.com/blog/go-time-monotonic-wall-clock/)
- [pkg.go.dev: time](https://pkg.go.dev/time)
- [How to Gracefully Close Channels — Go 101](https://go101.org/article/channel-closing.html)
- [Go by Example: Closing Channels](https://gobyexample.com/closing-channels)
- [pkg.go.dev: encoding/json](https://pkg.go.dev/encoding/json)
- [pkg.go.dev: encoding/json/v2](https://pkg.go.dev/encoding/json/v2)
- [Surprises and Gotchas When Working with JSON](https://www.alexedwards.net/blog/json-surprises-and-gotchas)
- [Understanding init in Go (DigitalOcean)](https://www.digitalocean.com/community/tutorials/understanding-init-in-go)
- [Package Initialization Order in Go Modules](https://medium.com/@AlexanderObregon/package-initialization-order-in-go-modules-8624a8732fa1)
- [Choosing a value or pointer receiver — Go Tour](https://go.dev/tour/methods/8)
- [Pointer Receivers and Interface Compliance in Go](https://themsaid.com/pointer-receivers-interface-compliance-go)
- [TIL: Go Response Body MUST be closed (Manish R Jain)](https://manishrjain.com/must-close-golang-http-response)
- [Don't use Go's default HTTP client (in production)](https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779)
- [How to Use Generics with Type Constraints in Go](https://oneuptime.com/blog/post/2026-01-23-go-generics-constraints/view)
- [100 Go Mistakes #9 — Being confused about when to use generics](https://100go.co/9-generics/)
