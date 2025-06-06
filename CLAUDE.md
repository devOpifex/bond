# CLAUDE.md for Bond Codebase

## Build & Test Commands
- `make test` - Run all tests with race detection
- `make test-verbose` - Run all tests with verbose output
- `make examples` - Run example programs
- `make clean` - Clean build artifacts
- Single test: `go test ./path/to/package -run TestName -v`

## Code Style Guidelines
- **Imports**: Standard library first, blank line, then external imports
- **Formatting**: Standard Go formatting (gofmt/goimports)
- **Types**: Interfaces in models/interfaces.go, types in models/types.go
- **Naming**: CamelCase for exports, lowercase package names.
Keep function and method names short.
- **Error Handling**: Explicit error checking with if err != nil pattern
- **Documentation**: Package-level docs, godoc format for functions
- **Project Structure**: Organized by functionality (providers, tools, reasoning)
- **Testing**: Table-driven tests in *_test.go files, mocks for dependencies
- Use `any` instead of `interface{}` where possible
- Use early returns and avoid `else` blocks

Use `make` commands whenever possible for consistency.
