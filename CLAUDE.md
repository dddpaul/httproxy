# Development Guidelines for gopkm

## Build & Test Commands
- Build: `go build ./...`
- Run tests: `go test ./...`
- Run specific test: `go test -run TestName ./path/to/package`
- Run tests with coverage: `go test -cover ./...`
- Run linting: `golangci-lint run ./...`
- Format code: `gofmt -s -w .`
- Run code generation: `go generate ./...` (needed before commit if any file has //go:generate directives)

## Important Workflow Notes
- Do not execute any tasks that not being asked to do
- Always run build, tests, linter before committing
- Run tests and linter after making significant changes to verify functionality
- Go version: 1.24+
- Don't add "Generated with Claude Code" or "Co-Authored-By: Claude" to commit messages or PRs
- Do not include "Test plan" sections in PR descriptions
- Do not add comments that describe changes, progress, or historical modifications. Avoid comments like “new function,” “added test,” “now we changed this,” or “previously used X, now using Y.” Comments should only describe the current state and purpose of the code, not its history or evolution.
- Use go:generate for generating mocks, never modify generated files manually. Mocks are generated with moq and stored in the mocks package.
- After important functionality added, update README.md accordingly
- When merging master changes to an active branch, make sure both branches are pulled and up to date first
- Don't add "Test plan" section to PRs

## Code Style Guidelines
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use snake_case for filenames, camelCase for variables, PascalCase for exported names
- Group imports: standard library, then third-party, then local packages
- Error handling: check errors immediately and return them with context
- Use meaningful variable names; avoid single-letter names except in loops
- Validate function parameters at the start before processing
- Return early when possible to avoid deep nesting
- Prefer composition over inheritance
- Function size preferences:
  - Aim for functions around 50-60 lines when possible
  - Don't break down functions too small as it can reduce readability
  - Maintain focus on a single responsibility per function
- Comment style: in-function comments should be lowercase sentences
- Code width: keep lines under 130 characters when possible

# Test Guidelines
- Do not use mocks with testify/mock, use interfaces and structs to simulate testing object
