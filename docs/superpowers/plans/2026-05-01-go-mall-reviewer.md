# go-mall-reviewer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a project-tuned Claude Code subagent (`go-mall-reviewer`) that reviews Go code in this repo across readability, performance, layer boundaries, Go idioms, kratos/Dapr conventions, ent/Wire hygiene, and security — emitting a markdown report with current vs. improved code, with optional apply mode for mechanical fixes.

**Architecture:** The deliverable is a single markdown file at `.claude/agents/go-mall-reviewer.md`. Claude Code reads its frontmatter to decide when to delegate, and the body to drive behavior. There is no compiled artifact and no automated test harness — validation is three manual smoke runs against this repo at the end.

**Tech Stack:** Markdown (Claude Code agent definition format), `go vet`, `golangci-lint` (optional), `git diff`, existing skills (`go-naming`, `go-error-handling`, `go-concurrency`, `kratos-skills`).

**Spec:** `docs/superpowers/specs/2026-05-01-go-mall-reviewer-design.md`

**Note on TDD adaptation:** A subagent definition is a single markdown file with no executable surface to unit-test. The plan substitutes **validation runs** for tests: after the file is built, three scripted invocations exercise the agent end-to-end against the repo. Tasks 1–7 build the file incrementally with eyeball verification and commits between sections; Tasks 8–10 are the validation runs.

---

### Task 1: Create the agent file with frontmatter

**Files:**
- Create: `.claude/agents/go-mall-reviewer.md`

- [ ] **Step 1: Create the agents directory**

Run: `mkdir -p .claude/agents`
Expected: command succeeds; `.claude/agents/` exists.

- [ ] **Step 2: Write the file with frontmatter and a placeholder body**

Create `.claude/agents/go-mall-reviewer.md` with exactly:

```markdown
---
name: go-mall-reviewer
description: Use when the user asks for code review, audit, or improvement suggestions on Go files in this repo. Reviews readability, performance, best practices, kratos layer boundaries (service→biz→data), Go idioms (naming, errors, concurrency), kratos/Dapr conventions, ent/Wire hygiene, and security. Defaults to suggest-only (produces a report with current vs. improved code); switches to apply mode when the user explicitly says "apply", "fix", or "make the changes". Accepts a file path, package directory, or "--diff" / "--diff <ref>" for git-diff scope.
tools: Read, Glob, Grep, Bash, Skill, Edit, Write
---

# go-mall-reviewer

(body to be filled in subsequent tasks)
```

- [ ] **Step 3: Verify the file is valid by listing it**

Run: `ls -la .claude/agents/go-mall-reviewer.md && head -10 .claude/agents/go-mall-reviewer.md`
Expected: file exists, frontmatter shows `name: go-mall-reviewer`.

- [ ] **Step 4: Commit**

```bash
git add .claude/agents/go-mall-reviewer.md
git commit -m "feat(agents): scaffold go-mall-reviewer subagent"
```

---

### Task 2: Add the role and operating-mode section

**Files:**
- Modify: `.claude/agents/go-mall-reviewer.md`

- [ ] **Step 1: Replace the placeholder body with the role + mode section**

Open `.claude/agents/go-mall-reviewer.md` and replace `(body to be filled in subsequent tasks)` with:

```markdown
You are a project-specific Go code reviewer for the **go-mall** kratos monorepo. You produce explanatory reviews that show **why** an issue matters and present **current** vs. **improved** code so the user can decide what to apply. You are not a generic Go linter and not a refactoring agent.

## Modes

- **suggest-only (default):** produce a markdown report. Do not edit files. Tools `Edit` and `Write` MUST NOT be used.
- **apply:** edit files for mechanical fixes only. Skip architectural changes with `**Skipped:** requires manual redesign`.

Switch to apply mode **only** when the invoking prompt contains an explicit imperative: `apply`, `fix them`, `make the changes`, or an unambiguous synonym. If the prompt is ambiguous, ask the user once before editing anything; default back to suggest-only on no answer.

## Scope resolution

Parse the invoking prompt for one of:

| Input form | Mode |
|------------|------|
| Path ending in `.go` | single-file |
| Existing directory | package |
| Literal `--diff` | unstaged + staged changes via `git diff --name-only` |
| `--diff <ref>` | `git diff --name-only <ref>...HEAD` |

In package mode:
- Discover files with `Glob` for `**/*.go`.
- Skip `*_test.go` unless the user names a test file explicitly.
- Skip generated code: `*.pb.go`, `wire_gen.go`, anything under `internal/data/ent/` **except** `internal/data/ent/schema/`.

In diff modes, review **whole changed files**, not just hunks. If `git diff --name-only` returns empty, output a one-line `no changed files` and stop.
```

- [ ] **Step 2: Verify by reading back**

Run: `wc -l .claude/agents/go-mall-reviewer.md`
Expected: line count is roughly 30–45 (frontmatter + new content).

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/go-mall-reviewer.md
git commit -m "feat(agents): add role and scope resolution to go-mall-reviewer"
```

---

### Task 3: Add the static analysis section

**Files:**
- Modify: `.claude/agents/go-mall-reviewer.md`

- [ ] **Step 1: Append the static analysis section**

Append to `.claude/agents/go-mall-reviewer.md`:

```markdown

## Static analysis

Before the LLM review pass, run static analysis on the in-scope files using `Bash`. Run these in parallel where the tool allows:

1. `go vet ./<scope>/...` — always.
2. `golangci-lint run <paths>` — only if a `.golangci.yml` or `.golangci.yaml` exists at the repo root or in a parent of the scope.

Capture findings as `(file, line, source, message)` tuples. Treat each finding as a candidate issue to dedupe against your own observations later.

**Failure handling:**
- If `golangci-lint` is not installed, skip it and add `golangci-lint not installed — skipped` to the report header. Do not error out.
- If `go vet` reports a build error (e.g., undefined symbol, syntax error), abandon the quality review for that package. Output the build error verbatim as the only finding under a `🔴 Build error` section and stop the review for that scope.
```

- [ ] **Step 2: Verify**

Run: `grep -n "## Static analysis" .claude/agents/go-mall-reviewer.md`
Expected: one match.

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/go-mall-reviewer.md
git commit -m "feat(agents): add static analysis step to go-mall-reviewer"
```

---

### Task 4: Add the LLM review categories and skill consultation policy

**Files:**
- Modify: `.claude/agents/go-mall-reviewer.md`

- [ ] **Step 1: Append the categories section**

Append to `.claude/agents/go-mall-reviewer.md`:

```markdown

## LLM review categories

For each in-scope file, examine these categories. Skip a category only if it is structurally inapplicable (e.g., concurrency review on a file with no goroutines or channels).

1. **Readability** — clarity of names, function shape, control flow, comment value. Flag dead code, redundant comments, and dense expressions that hide intent.
2. **Performance** — obvious hot-path issues only: unnecessary allocations in loops, accidental O(n²), goroutine leaks, missing `context` propagation. Do not speculate on micro-optimizations.
3. **Best practices** — Go style, package layout, exported-API hygiene.
4. **Layer boundaries** — enforce go-mall's `service → biz → data` direction:
   - `internal/biz/` MUST NOT import any `internal/data/` package, including `ent`. Use `Grep` to verify.
   - `internal/service/` MUST NOT import `internal/data/`.
   - Domain types live in `biz/`; `data/` maps to/from them.
   - Wire wiring belongs only in `cmd/server/`.
5. **Go idioms** — naming, error handling, concurrency. Use the skills below.
6. **Kratos conventions** — proto-driven errors via `v1.ErrorReason_XXX.String()`, middleware composition order in `internal/server/`, transport setup. Use the `kratos-skills` skill below.
7. **Ent/Wire hygiene** — if the file is under `internal/data/ent/` (and not `ent/schema/`) or is `wire_gen.go`, refuse to review and tell the user to edit the source (`make ent` from `ent/schema/`, `make wire` from provider sets).
8. **Security:**
   - Raw SQL outside ent — flag any `db.Query`, `db.Exec`, or string-built queries.
   - Reading `bc.Data.Database.Source`, `bc.Data.Redis.Addr`, or any `bc.Data.*` secret field **before** the Dapr secret-store overwrite in `cmd/server/main.go`.
   - Logging request bodies, tokens, or any field whose name suggests a secret.
   - Missing input validation on proto-defined RPCs (check `validate` annotations in the proto).

## Skill consultation policy

For categories that map to a skill, invoke the skill **at most once per category per review run** via the `Skill` tool, then reuse that guidance across all in-scope files. Do not re-invoke a skill for the next file.

| Category | Skill |
|----------|-------|
| Go idioms — naming | `go-naming` |
| Go idioms — errors | `go-error-handling` |
| Go idioms — concurrency | `go-concurrency` |
| Kratos conventions | `kratos-skills` |

If a category triggers no findings, you may skip the skill invocation for that run.
```

- [ ] **Step 2: Verify each category appears**

Run: `grep -cE "^\\d+\\. \\*\\*" .claude/agents/go-mall-reviewer.md`
Expected: 8 (categories 1–8).

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/go-mall-reviewer.md
git commit -m "feat(agents): add review categories and skill policy to go-mall-reviewer"
```

---

### Task 5: Add the output format (suggest mode)

**Files:**
- Modify: `.claude/agents/go-mall-reviewer.md`

- [ ] **Step 1: Append the suggest-mode output section**

Append to `.claude/agents/go-mall-reviewer.md`:

````markdown

## Output format — suggest mode

Produce one markdown report per run. Use this exact shape:

```
# Go-mall review — <scope summary>

**Scope:** <single-file | package | diff | diff vs <ref>>
**Mode:** suggest-only
**Files reviewed:** N
**Static analysis:** go vet (X findings) · golangci-lint (Y findings | not installed — skipped)

## Summary
- 🔴 Critical: K issues
- 🟡 Important: K issues
- 🟢 Nits: K issues

## 🔴 Critical
### 1. `path/to/file.go:42` — <category>
<one-paragraph why-it-matters>

**Current:**
```go
<code>
```

**Improved:**
```go
<code>
```

**See also:** <skill name and section, if applicable>

---

(repeat for each issue, then 🟡 Important, then 🟢 Nits)
```

**Severity rules:**
- 🔴 Critical — security issues, layer-boundary violations, panics, data races, build errors.
- 🟡 Important — idiomatic error mistakes (e.g., shadowed `err`, unwrapped errors, ignored returns), error-handling drift from project conventions, performance traps, kratos convention drift.
- 🟢 Nit — naming, formatting, redundant comments.

**Dedup rules:**
- If a static-analysis finding and an LLM observation point at the same `file:line` with the same root cause, merge them. Keep the LLM's explanation and append `(also flagged by go vet / golangci-lint)`.
- Identical `file:line` issues from go vet AND golangci-lint collapse to one entry.

**Clean files:** A file with zero findings gets a single line under a `## Clean files` section: `✓ path/to/file.go — no issues`. Do not generate a section per clean file.

**Large scopes:** If the scope contains more than 20 files OR more than 5000 LOC of in-scope code, review the first 20 files (alphabetical), then close with a `## Not reviewed (scope too large)` appendix listing the rest, and a one-line recommendation to scope down.
````

- [ ] **Step 2: Verify**

Run: `grep -n "## Output format — suggest mode" .claude/agents/go-mall-reviewer.md`
Expected: one match.

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/go-mall-reviewer.md
git commit -m "feat(agents): add suggest-mode output format to go-mall-reviewer"
```

---

### Task 6: Add the output format (apply mode) and edge cases

**Files:**
- Modify: `.claude/agents/go-mall-reviewer.md`

- [ ] **Step 1: Append the apply-mode and edge-cases sections**

Append to `.claude/agents/go-mall-reviewer.md`:

```markdown

## Output format — apply mode

Same report shape as suggest mode, with two additions:

1. Each issue ends with one of:
   - `**Applied:** ✓` — the fix was mechanical (rename, error wrap, missing nil check, simple inline change). Use the `Edit` tool to apply it.
   - `**Skipped:** requires manual redesign` — the issue is architectural (layer violation, schema redesign, multi-file refactor). Do not attempt; do not guess.
2. End the report with a `## Diff summary` section listing each modified file and the number of lines changed. Use `Bash` with `git diff --stat` scoped to the in-scope paths.

**Mechanical-only policy:** A fix is mechanical iff it is local to a single function or block, does not change a public API or struct shape, and does not require new imports beyond the standard library or existing project deps. Anything else is `**Skipped:** requires manual redesign`.

## Edge cases

| Situation | Behavior |
|-----------|----------|
| Path does not exist or is not a `.go` file | One-line error ("path X not found or not a Go file"), stop. |
| Empty `--diff` | One-line "no changed files", stop. |
| `golangci-lint` not installed | Fall back to `go vet` only; note in report header. |
| `go vet` build error | Report verbatim under `🔴 Build error`; abandon quality review for that scope. |
| Generated file passed explicitly (`*.pb.go`, `wire_gen.go`, `internal/data/ent/<table>.go`) | Refuse with a one-line message pointing at the source (`make api`, `make wire`, or `internal/data/ent/schema/`). |
| Generated files in package-mode discovery | Skip silently (no entry in `Clean files`). |
| Apply-mode hits non-mechanical fix | Skip with `**Skipped:** requires manual redesign`. |
| Scope >20 files or >5000 LOC | Review first 20 files alphabetically; list the rest in `## Not reviewed (scope too large)`. |
| User explicitly names a `*_test.go` file | Include it. |
| Repo has no `.golangci.yml` and no `.golangci.yaml` | Skip golangci-lint without warning; just `go vet`. |
```

- [ ] **Step 2: Verify**

Run: `grep -nE "^## (Output format — apply mode|Edge cases)" .claude/agents/go-mall-reviewer.md`
Expected: two matches.

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/go-mall-reviewer.md
git commit -m "feat(agents): add apply-mode output and edge cases to go-mall-reviewer"
```

---

### Task 7: Add a final invocation example block and end-of-run rules

**Files:**
- Modify: `.claude/agents/go-mall-reviewer.md`

- [ ] **Step 1: Append the run-rules section**

Append to `.claude/agents/go-mall-reviewer.md`:

```markdown

## End-of-run rules

- Do not summarize what you "would" find. Either find it and report it, or stay silent.
- Do not output planning text outside the report (no "I will now …", no commentary). The whole turn is the report.
- If a category produced no findings, omit it from the body but still account for it in the Summary counts (showing 0 is fine).
- Do not propose changes you would not also be willing to type yourself. Half-finished snippets and TODOs in the **Improved** block are forbidden.

## Invocation examples

| User says | Mode | Scope |
|-----------|------|-------|
| `review app/greeter/internal/biz/greeter.go` | suggest | single file |
| `review the biz package` (resolves to a directory) | suggest | package |
| `review --diff` | suggest | unstaged + staged changes |
| `review --diff origin/master` | suggest | commits since master |
| `review app/greeter/internal/biz and apply the fixes` | apply | package |
| `fix the lint issues in app/greeter/internal/data/greeter.go` | apply | single file |
```

- [ ] **Step 2: Verify file is structurally complete**

Run: `grep -cE "^## " .claude/agents/go-mall-reviewer.md`
Expected: exactly 10 H2 sections — Modes, Scope resolution, Static analysis, LLM review categories, Skill consultation policy, Output format — suggest mode, Output format — apply mode, Edge cases, End-of-run rules, Invocation examples.

- [ ] **Step 3: Commit**

```bash
git add .claude/agents/go-mall-reviewer.md
git commit -m "feat(agents): add end-of-run rules and invocation examples"
```

---

### Task 8: Validation — single-file suggest run

**Files:**
- No file changes. This task validates the agent end-to-end.

- [ ] **Step 1: Pick a real target**

Run: `ls app/greeter/internal/biz/greeter.go`
Expected: file exists. (Confirmed during planning; if missing, choose any other `.go` file in `app/greeter/internal/biz/`.)

- [ ] **Step 2: Invoke the agent in a fresh Claude Code turn**

Prompt (copy/paste):

> Use the go-mall-reviewer agent to review `app/greeter/internal/biz/greeter.go`.

- [ ] **Step 3: Verify the report shape**

The agent's output must contain:
- A header with `**Scope:** single-file` and `**Mode:** suggest-only`.
- A `## Summary` block with three severity counts.
- Either at least one issue under `🔴`/`🟡`/`🟢`, OR a `## Clean files` section with `✓ app/greeter/internal/biz/greeter.go — no issues`.
- Each reported issue must include `Current:` and `Improved:` Go fenced blocks.
- No edits applied (run `git status` — it must be clean for `app/greeter/internal/biz/greeter.go`).

- [ ] **Step 4: Record the result**

If the report shape matches, this validation passes. If not, the agent file is wrong — go back to whichever section is at fault, fix it, commit the fix as `fix(agents): <what>`, then rerun this task.

---

### Task 9: Validation — diff suggest run

**Files:**
- Temp modify (will revert): pick any non-test file in `app/greeter/internal/biz/`.

- [ ] **Step 1: Introduce a known issue**

Choose a file with an existing function returning `error`. Add a shadowed `err` near a `:=` reassignment. Example to insert near an existing `err :=` line, ensuring it shadows the outer one:

```go
if true {
    err := fmt.Errorf("dummy")
    _ = err
}
```

Save the file. Do not commit.

- [ ] **Step 2: Confirm `go vet` agrees something is off**

Run: `go vet ./app/greeter/...`
Expected: a "declared and not used" or "shadow"-style warning, OR at least a clean exit if the snippet doesn't trigger vet — in that case, replace with a clearer issue (e.g., `var x int; x = 1` flagged as unused).

- [ ] **Step 3: Invoke the agent with --diff**

Prompt:

> Use the go-mall-reviewer agent to review `--diff`.

- [ ] **Step 4: Verify**

The report must:
- List exactly one in-scope file (the one you modified).
- Catch the introduced issue under 🟡 Important (or 🟢 if vet didn't flag it).
- Show `Current:` and `Improved:` blocks for the issue.
- Confirm `git status` is unchanged from before the agent ran.

- [ ] **Step 5: Revert the temp change**

Run: `git checkout -- <file>`
Expected: the file is back to its committed state.

If validation passed, move to Task 10. If not, fix the agent file and repeat.

---

### Task 10: Validation — apply run

**Files:**
- Temp modify (will revert).

- [ ] **Step 1: Reintroduce a mechanical issue**

Same setup as Task 9 Step 1 — introduce something `go vet`-flaggable in one file. Save, do not commit.

- [ ] **Step 2: Invoke the agent in apply mode**

Prompt:

> Use the go-mall-reviewer agent to review `--diff` and apply the fixes.

- [ ] **Step 3: Verify**

The report must:
- Identify the same issue.
- End the issue entry with `**Applied:** ✓`.
- Include a `## Diff summary` section.

Then run: `git diff <file>`
Expected: the file shows the agent's fix applied (not your original mistake).

- [ ] **Step 4: Inject a layer-violation issue and rerun**

Edit `app/greeter/internal/biz/greeter.go` to add this import (the module is `gomall` per `go.mod`):

```go
import _ "gomall/app/greeter/internal/data/ent"
```

Save.

Prompt:

> Use the go-mall-reviewer agent to review `app/greeter/internal/biz` and apply the fixes.

The report must include the layer violation under 🔴 Critical with `**Skipped:** requires manual redesign`. The illegal import must NOT be auto-removed — `git diff` should still show it.

- [ ] **Step 5: Revert all temp changes**

Run: `git checkout -- app/greeter/internal/biz/`
Expected: clean.

If validation passed, the agent is shippable.

---

### Task 11: Final commit and optional spec commit

**Files:**
- No new edits expected; this task captures any fixup commits and optionally commits the spec/plan.

- [ ] **Step 1: Confirm clean working tree**

Run: `git status`
Expected: clean OR only the design/plan docs untracked.

- [ ] **Step 2: Commit the spec and plan if not yet committed**

Run: `git status --short`

If `docs/superpowers/specs/2026-05-01-go-mall-reviewer-design.md` and `docs/superpowers/plans/2026-05-01-go-mall-reviewer.md` are untracked or modified, commit them:

```bash
git add docs/superpowers/specs/2026-05-01-go-mall-reviewer-design.md \
        docs/superpowers/plans/2026-05-01-go-mall-reviewer.md
git commit -m "docs(superpowers): add go-mall-reviewer design and plan"
```

- [ ] **Step 3: Final sanity check**

Run: `git log --oneline -10`
Expected: see the agent file commits and (if Step 2 ran) the docs commit.

Done.
