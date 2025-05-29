.PHONY: test test-verbose examples clean

# Run all tests
test:
	go test ./... -race

# Run all tests with verbose output
test-verbose:
	go test ./... -v -race

# Run all examples
examples:
	go run ./examples/code/main.go
	go run ./examples/react/main.go
	go run ./examples/react_in_chain/main.go
	go run ./examples/weather/main.go

# Clean build artifacts
clean:
	rm -rf ./bin

# Help message
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run all tests with verbose output"
	@echo "  examples      - Run all examples"
	@echo "  clean         - Clean build artifacts"
