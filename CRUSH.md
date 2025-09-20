# CRUSH.md

This file provides guidelines for working with this Go codebase.

## Commands

- **Build:** `go build -o ./bin/server ./...`
- **Run:** `go run ./sample-app/src/main.go`
- **Test:** `go test ./...`
- **Run a single test:** `go test -run ^TestName$`
- **Lint:** `go vet ./...` (or `golangci-lint run` if installed)

## Code Style

- **Imports:** Group imports in a single block. Standard library packages first, then third-party packages.
- **Formatting:** Use `gofmt` to format your code before committing.
- **Types:** Use structs for data modeling. Use JSON tags for serialization.
- **Naming Conventions:**
  - Structs and Interfaces: `PascalCase`
  - Functions and variables: `camelCase`
- **Error Handling:**
  - Check for errors explicitly using `if err != nil`.
  - Use `log.Fatal` or `log.Fatalf` for critical errors that should stop the application.
  - Use `http.Error` to send HTTP error responses.
  - Use `log.Printf` for logging non-fatal errors.
- **Dependencies:** Manage dependencies using Go modules (`go.mod`).
