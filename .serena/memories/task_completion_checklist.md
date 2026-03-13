# Task Completion Checklist

When completing a coding task in open-jarvis:

## Go (Backend)
1. `go build ./...` — ensure code compiles
2. `go vet ./...` — check for issues
3. `go test ./...` — run tests (once tests exist)
4. `go mod tidy` — keep dependencies clean

## TypeScript (Frontend)
1. `npm run build` — ensure Next.js builds without errors
2. `npm run lint` — run linter (once configured)

## General
- Commit with clear, descriptive messages
- Keep services decoupled (Go backend, Next.js frontend are separate)
