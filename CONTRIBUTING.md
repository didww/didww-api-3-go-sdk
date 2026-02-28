# Contributing to DIDWW API v3 Go SDK

Thank you for your interest in contributing to the DIDWW API v3 Go SDK!

## Development

### Prerequisites

- Go 1.22 or later
- Git

### Running Tests

To run all tests:

```bash
go test -v ./...
```

To run tests with race detection and coverage:

```bash
go test -race -coverprofile=coverage.out -covermode=atomic ./...
```

To view coverage report:

```bash
go tool cover -html=coverage.out
```

### Linting

We use [golangci-lint](https://golangci-lint.run/) for code linting:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Formatting

Ensure your code is properly formatted:

```bash
# Format code
go fmt ./...

# Fix imports
goimports -w .
```

## Continuous Integration

This project uses GitHub Actions for CI/CD with the following workflows:

### CI Workflow (`.github/workflows/ci.yml`)

- **Triggers**: Push/PR to `main`, `master`, `develop` branches
- **Go versions**: 1.22, 1.23
- **Platforms**: Ubuntu, Windows, macOS
- **Features**:
  - Dependency verification
  - Code analysis with `go vet`
  - Race condition detection
  - Test coverage reporting
  - Codecov integration

### Lint Workflow (`.github/workflows/lint.yml`)

- **Triggers**: Push/PR to `main`, `master`, `develop` branches
- **Features**:
  - golangci-lint analysis
  - Code formatting verification
  - Import organization checks

## Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch from `main`
3. **Make** your changes
4. **Add** tests for new functionality
5. **Run** tests locally: `go test ./...`
6. **Run** linter: `golangci-lint run`
7. **Commit** your changes with clear commit messages
8. **Push** to your fork
9. **Create** a pull request

### Pull Request Requirements

- ✅ All tests pass
- ✅ Code coverage is maintained
- ✅ Code passes linting
- ✅ Changes are documented
- ✅ Commit messages are descriptive

## Code Style

This project follows standard Go conventions:

- Use `gofmt` for formatting
- Use `goimports` for import organization  
- Follow effective Go guidelines
- Write clear, self-documenting code
- Include comments for exported types and functions

## Testing

- Write table-driven tests where appropriate
- Use the existing test helpers in [`testhelper_test.go`](testhelper_test.go)
- Mock external dependencies using [`testdata/fixtures`](testdata/fixtures)
- Ensure good test coverage for new code
- Test error conditions and edge cases

## Documentation

- Update README.md for new features
- Add examples to the [`examples/`](examples/) directory
- Document public APIs with clear comments
- Include usage examples in docstrings

## License

By contributing, you agree that your contributions will be licensed under the MIT License.