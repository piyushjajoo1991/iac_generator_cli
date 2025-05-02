# IaC Generator CLI Testing Framework

This testing framework provides comprehensive testing for the IaC Manifest Generator CLI. The tests cover all aspects of the system, from individual unit tests to full end-to-end integration tests.

## Test Structure

The test suite is organized into the following packages:

- **utils**: Utility functions and helpers for testing
- **fixtures**: Sample infrastructure descriptions and expected outputs
- **nlp**: Natural language processing and parsing tests
- **infra**: Infrastructure model validation tests
- **template**: Template rendering accuracy tests
- **pipeline**: End-to-end pipeline integration tests
- **adapter**: File generation and structure tests for IaC types
- **cmd**: CLI command execution tests

## Test Types

1. **Unit Tests**: Test individual components in isolation
   - NLP parser and pattern matching
   - Infrastructure model creation and validation
   - HCL and YAML writing

2. **Integration Tests**: Test how components work together
   - Pipeline stages integration
   - Template rendering with infrastructure models
   - Generator output validation

3. **End-to-End Tests**: Test full system functionality
   - CLI command execution
   - Full pipeline processing
   - File generation in different IaC formats

## Running Tests

### Run All Tests

To run all tests, use:

```sh
go test ./test/...
```

### Run Specific Test Packages

To run tests for specific packages:

```sh
# Run NLP parser tests
go test ./test/nlp

# Run template rendering tests
go test ./test/template

# Run pipeline integration tests
go test ./test/pipeline
```

### Run with Verbose Output

For more detailed output:

```sh
go test -v ./test/...
```

### Skip Slow Tests

To skip slow tests (like CLI execution tests):

```sh
go test -short ./test/...
```

## Test Fixtures

Test fixtures are provided in the `test/fixtures` package. These include:

- Sample infrastructure descriptions
- Expected parsed entities
- Reference infrastructure models
- Expected output files for comparison

## Mock Utilities

The testing framework includes mock implementations for:

- File system operations
- Pipeline components (NLP processor, model builder, etc.)
- Output writers

These mocks allow for isolated testing of components without real file system or external dependencies.

## File Comparison

The framework includes utilities for comparing:

- Directory structures
- File contents
- Generated manifests against expected references

These help ensure that the generated outputs are correct.

## Test Environment

A temporary test environment is created for each test, providing:

- Isolated output directories
- Fixture directories
- Clean state for each test run

This environment is automatically cleaned up after each test.

## Adding New Tests

To add new tests:

1. Identify the appropriate package for the test
2. Create a new test file or add to an existing one
3. Use the provided utilities and fixtures
4. Follow the existing patterns for setting up test environments

For table-driven tests, use the `fixtures.TestDescription` structure to organize test cases.

## Running CLI Tests

CLI tests require a built binary. Before running these tests:

1. Build the CLI binary:
   ```sh
   go build -o iacgen main.go
   ```

2. Run the CLI tests:
   ```sh
   go test ./test/cmd
   ```

## Test Coverage

To run tests with coverage reporting:

```sh
go test -cover ./test/...
```

For a detailed coverage report:

```sh
go test -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out
```