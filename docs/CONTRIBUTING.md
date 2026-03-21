# Contributing to NanoFuse

Thank you for your interest in contributing to NanoFuse! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Docker (for building images)
- Linux with KVM support (for integration testing)
- golangci-lint (for linting)

### Development Setup

1. **Fork and clone the repository**:

```bash
git clone https://github.com/YOUR_USERNAME/nanofuse.git
cd nanofuse
```

2. **Install dependencies**:

```bash
make deps
```

3. **Build the project**:

```bash
make build
```

4. **Run tests**:

```bash
make test
```

## Development Workflow

### Making Changes

1. **Create a feature branch**:

```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes** following the code style guidelines

3. **Run tests and linters**:

```bash
make test
make lint
```

4. **Commit your changes**:

```bash
git add .
git commit -m "feat: Add your feature description"
```

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test additions or changes
- `chore:` Maintenance tasks
- `ci:` CI/CD changes

5. **Push to your fork**:

```bash
git push origin feature/your-feature-name
```

6. **Open a Pull Request** on GitHub

### Code Review Process

1. All PRs must pass CI checks (build, test, lint, security scan)
2. At least one maintainer review is required
3. Code must follow the project's style guidelines
4. Tests must be included for new features
5. Documentation must be updated if needed

## Code Style

### Go Code

- Follow standard Go formatting (`gofmt`, `goimports`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused (max 50 lines preferred)
- Use early returns to reduce nesting

### Error Handling

Always handle errors explicitly:

```go
// Good
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Bad
result, _ := doSomething()  // Don't ignore errors
```

### Testing

- Write unit tests for all new code
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example test structure:

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "result", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FeatureName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FeatureName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FeatureName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Project Structure

```
nanofuse/
├── cmd/                    # Command-line applications
│   ├── nanofuse/          # CLI tool
│   └── nanofused/         # API daemon
├── internal/              # Private application code
│   ├── api/              # API server implementation
│   ├── firecracker/      # Firecracker integration
│   ├── image/            # Image management
│   └── vm/               # VM lifecycle management
├── images/                # Docker images
│   └── base/             # Base microVM image
├── docs/                  # Documentation
├── test/                  # Integration tests
└── systemd/              # Systemd unit files
```

## CI/CD Pipeline

Our CI/CD pipeline runs on every push and PR:

### Jobs

1. **build-go**: Builds binaries and runs tests
2. **build-image**: Builds Docker image
3. **security-scan**: Runs vulnerability scans
4. **lint**: Runs code quality checks
5. **release**: Creates releases (tags only)

### What Gets Published

- **On PR**: Nothing (build and test only)
- **On merge to main**: Docker image to GHCR with `latest` and `sha-*` tags
- **On tag `v*`**: GitHub release with binaries + Docker image with version tag

## Making a Release

Releases are automated via GitHub Actions when you push a version tag:

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Create and push a tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

This will:
1. Trigger the CI/CD pipeline
2. Build all artifacts
3. Create a GitHub release
4. Upload binaries to the release
5. Push Docker image with version tag

## Documentation

When adding features, update:

- **README.md**: User-facing features and usage
- **docs/API_CONTRACT.md**: API changes
- **docs/CLI_SPEC.md**: CLI command changes
- Code comments and examples

## Getting Help

- 📚 Check [documentation](docs/)
- 🐛 Search [existing issues](https://github.com/jpoley/nanofuse/issues)
- 💬 Start a [discussion](https://github.com/jpoley/nanofuse/discussions)
- 📧 Contact maintainers

## Code of Conduct

This project follows a Code of Conduct. By participating, you agree to uphold this code:

- Be respectful and inclusive
- Welcome newcomers
- Focus on what's best for the project
- Show empathy towards others

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Don't hesitate to ask questions in issues or discussions. We're here to help!
