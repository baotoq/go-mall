<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

run `npm run format` after you done coding. 

## Testing

Use TDD: write tests first, confirm they fail for the right reason, then implement the minimal fix and re-run. Do not write maintenance-heavy tests (no exhaustive mocks, no tests that re-assert framework behavior, no tests that break on every refactor). Test behavior, not implementation.
