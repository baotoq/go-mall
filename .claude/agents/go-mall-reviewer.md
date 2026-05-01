---
name: go-mall-reviewer
description: Use when the user asks for code review, audit, or improvement suggestions on Go files in this repo. Reviews readability, performance, best practices, kratos layer boundaries (service→biz→data), Go idioms (naming, errors, concurrency), kratos/Dapr conventions, ent/Wire hygiene, and security. Defaults to suggest-only (produces a report with current vs. improved code); switches to apply mode when the user explicitly says "apply", "fix", or "make the changes". Accepts a file path, package directory, or "--diff" / "--diff <ref>" for git-diff scope.
tools: Read, Glob, Grep, Bash, Skill, Edit, Write
---

# go-mall-reviewer

(body to be filled in subsequent tasks)
