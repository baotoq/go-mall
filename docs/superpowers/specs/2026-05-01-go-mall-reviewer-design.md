# go-mall-reviewer — design spec

**Date:** 2026-05-01
**Status:** Approved, ready for implementation plan
**Owner:** Bao To

## Purpose

A project-tuned subagent that reviews Go code in this repository and produces an explanatory, decision-ready report. Fills the gap between the existing `simplify` / `code-simplifier` (which apply changes) and `code-reviewer` (which judges against a plan) by foregrounding **why** an issue matters and showing **current vs. improved** code side by side.

The agent is not a generic Go reviewer. It encodes go-mall's architecture (kratos `service → biz → data`), Wire/ent conventions, and Dapr secret-handling rules.

## Non-goals

- Not a generic, language-agnostic reviewer.
- Not a replacement for `golangci-lint` in CI — the agent runs lint as input, but local CI/pre-commit gates remain authoritative.
- Not a test author — it does not generate new tests.
- Not a refactoring agent — apply mode handles mechanical fixes only; architectural changes are reported and skipped.

## Architecture

A single subagent definition at `.claude/agents/go-mall-reviewer.md`. No supporting scripts, no separate slash command. Invocation is via Claude Code's Task tool, triggered by the agent description.

### Frontmatter

```yaml
---
name: go-mall-reviewer
description: Use when the user asks for code review, audit, or improvement suggestions on Go files in this repo. Reviews readability, performance, best practices, kratos layer boundaries (service→biz→data), Go idioms (naming, errors, concurrency), kratos/Dapr conventions, ent/Wire hygiene, and security. Defaults to suggest-only (produces a report with current vs. improved code); switches to apply mode when the user explicitly says "apply", "fix", or "make the changes". Accepts a file path, package directory, or "--diff" / "--diff <ref>" for git-diff scope.
tools: Read, Glob, Grep, Bash, Skill, Edit, Write
---
```

`Edit` and `Write` are listed but the agent body forbids using them outside apply mode.

### Operating procedure

1. **Resolve scope** from the invocation prompt:
   - Path ending in `.go` → single-file mode.
   - Directory → package mode. Read all `.go` files; skip `*_test.go` unless the user names them; skip `internal/data/ent/` generated code (but keep `ent/schema/`).
   - `--diff` → review files from `git diff --name-only` (unstaged + staged).
   - `--diff <ref>` → review files from `git diff --name-only <ref>...HEAD`.
   - Diff modes review **whole changed files**, not just hunks.
2. **Detect mode.** Default suggest-only. Switch to apply mode only when the prompt contains an explicit imperative: "apply", "fix them", "make the changes", or equivalent. If ambiguous, ask the user once before doing anything.
3. **Run static analysis** in parallel:
   - `go vet ./<scope>/...`
   - `golangci-lint run <paths>` if `.golangci.yml` exists.
   - Capture findings as `(file, line, source, message)` tuples.
4. **LLM review pass** over each in-scope file with these categories:
   - Readability, performance, best practices.
   - Layer boundaries — `biz` must not import `data`; `service` must not import `data`; `cmd/server` is the only place that wires concretes.
   - Go idioms — naming, errors, concurrency. The agent invokes the matching skill (`go-naming`, `go-error-handling`, `go-concurrency`) **at most once per category per run** via the Skill tool, and reuses that guidance across files.
   - Kratos conventions — proto-driven errors via `v1.ErrorReason_XXX.String()`, middleware composition, server setup. Consult `kratos-skills` once.
   - Ent/Wire hygiene — warn if the user appears to be hand-editing `ent/` (other than `ent/schema/`) or `wire_gen.go`. Recommend `make ent` / `make wire`.
   - Security — input validation, raw SQL outside ent, reading `bc.Data.Database.Source` / `bc.Data.Redis.Addr` before Dapr secret injection at startup, accidental secret logging.
5. **Synthesize the report** (suggest mode) or **apply edits then summarize** (apply mode).

### Tools and why each is allowed

| Tool | Purpose |
|------|---------|
| `Read` | Load source files for review. |
| `Glob` | Discover `.go` files in package mode. |
| `Grep` | Spot-check patterns (e.g., `data` imports inside `biz/`). |
| `Bash` | Run `go vet`, `golangci-lint`, `git diff --name-only`. |
| `Skill` | Consult Go and kratos skills once per category per run. |
| `Edit` | Apply mechanical fixes (apply mode only). |
| `Write` | Reserved; only used if a fix needs a new file (rare). |

## Output format

### Suggest mode

Markdown report, single document per run.

```
# Go-mall review — <scope>

**Scope:** <file | package | diff>
**Mode:** suggest-only
**Files reviewed:** N
**Static analysis:** go vet (X findings) · golangci-lint (Y findings)

## Summary
- 🔴 Critical: K issues (security, layer violation)
- 🟡 Important: K issues (errors, idioms)
- 🟢 Nits: K issues (naming, formatting)

## 🔴 Critical

### 1. `path/to/file.go:42` — <category>

<one-paragraph why-it-matters>

**Current:**
\`\`\`go
<code>
\`\`\`

**Improved:**
\`\`\`go
<code>
\`\`\`

**See also:** <skill name and section>

---
```

Severity scheme:

- **🔴 Critical** — security issues, layer-boundary violations, panics, data races, broken builds.
- **🟡 Important** — idiomatic errors, error-handling mistakes, performance traps, kratos convention drift.
- **🟢 Nit** — naming, formatting, redundant comments.

Static-analysis findings are folded in alongside LLM observations; identical issues are deduped (LLM wins on explanation, lint wins on file:line precision).

A lint-clean, idiom-clean file gets one line: `✓ path/to/file.go — no issues`.

### Apply mode

Same report shape, with each issue ending in one of:

- `**Applied:** ✓` — the fix was mechanical (rename, error wrap, missing nil check) and the agent edited the file.
- `**Skipped:** requires manual redesign` — the issue is architectural (layer violation, schema redesign) and the agent will not guess.

A short diff summary at the end of the report lists files modified and total lines changed.

## Edge cases

| Situation | Behavior |
|-----------|----------|
| Path does not exist or is not a `.go` file | One-line error, stop. |
| Empty `--diff` | One-line note "no changed files", stop. |
| `golangci-lint` not installed | Fall back to `go vet` only; note in report header. |
| `go vet` reports a compile error | Report the build error verbatim as the only finding; do not attempt quality review of broken code. |
| Generated file passed explicitly (`*.pb.go`, `wire_gen.go`, `ent/<table>.go`) | Refuse with one-line message pointing at the source (`make api` / `make wire` / schema). |
| Generated files in package-mode discovery | Skip silently. |
| Apply-mode hits non-mechanical fix | Skip with `**Skipped:** requires manual redesign`. Do not attempt. |
| Scope >20 files or >5000 LOC | Review first 20 files; list rest in a "not reviewed" appendix; advise user to scope down. |
| User asks for review of `*_test.go` explicitly | Include them. Otherwise skip in package mode. |

## Validation plan

Manual smoke test, three scenarios. No automated harness — subagent behavior is hard to fixture.

1. **Single-file suggest run.** Invoke against `app/greeter/internal/biz/greeter.go`. Expect a structured report. If any layer-boundary issue exists, expect it surfaced; otherwise expect `✓` plus any lint findings as 🟢 nits.
2. **Diff suggest run.** Introduce a temporary shadowed `err` in a single function, do not commit, invoke with `--diff`. Expect the shadowed error caught as 🟡 with current/improved code. Revert the change.
3. **Apply run.** Same shadowed-err setup, invoke with "apply the fixes". Expect the file edited and the report shows `**Applied:** ✓` for that issue. If a layer-boundary issue is also present in the same diff, expect it `**Skipped:** requires manual redesign`. Revert.

The validation is run by the implementer once after the agent file lands.

## Risks and mitigations

| Risk | Mitigation |
|------|------------|
| Agent description is too broad and Claude delegates to it for every code question. | Description leads with "code review, audit, or improvement suggestions" specifically. Implementation will tune wording if over-triggers in practice. |
| Skill consultation balloons the agent's context. | Hard rule: at most one skill invocation per category per run, cached for the rest of the run. |
| Apply mode silently introduces wrong fix. | Mechanical-only policy with explicit `**Skipped:** requires manual redesign`. Apply mode requires explicit imperative phrasing in the invocation. |
| Lint findings drown out judgment-call observations. | Severity bucketing — most lint output lands in 🟢 nits; LLM observations occupy 🔴 / 🟡 unless they overlap with lint, in which case they merge. |
| Generated files reviewed as if hand-written. | Skip list (`*.pb.go`, `wire_gen.go`, `internal/data/ent/*.go` except `ent/schema/`). Refuse explicit invocations on generated files. |

## Out of scope for this spec

- Slash command wrapper. (Can be added later if invocation ergonomics are awkward.)
- CI integration. (Separate decision; the agent is for interactive review.)
- Multi-language support. (Repo is Go-only at the relevant layer.)
- Automated regression tests for the agent itself.
