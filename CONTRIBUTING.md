# Contributing to New Relic Exporter

Thank you for your interest in contributing to the New Relic Exporter! This document provides guidelines and information for contributors.

## Project Status

**This project is actively maintained and welcomes contributions!**

We appreciate contributions of all kinds, including:
- Bug reports and fixes
- Feature requests and implementations
- Documentation improvements
- Test coverage improvements
- Code quality enhancements
- Performance optimizations

## How to Contribute

### Reporting Issues

If you find a bug or have a feature request:

1. **Check existing issues** to avoid duplicates
2. **Open a new issue** with a clear title and description
3. **Include details**:
   - For bugs: Steps to reproduce, expected vs actual behavior, environment details
   - For features: Use case, proposed solution, any alternatives considered
4. **Add labels** if you have permission (bug, enhancement, documentation, etc.)

### Submitting Pull Requests

1. **Fork the repository** and create a new branch from `main`
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes**:
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed
   - Keep commits focused and atomic

3. **Test your changes**:
   ```bash
   # Run tests
   go test -v ./...

   # Check formatting
   gofmt -l .

   # Run go vet
   go vet ./...

   # Build to ensure no errors
   go build .
   ```

4. **Commit your changes**:
   - Use clear, descriptive commit messages
   - Reference related issues (e.g., "Fixes #123")
   - Follow conventional commits format if possible:
     ```
     feat: add support for mobile metrics
     fix: resolve timeout issue with large metric sets
     docs: improve configuration examples
     test: add tests for config validation
     ```

5. **Push to your fork** and submit a pull request:
   ```bash
   git push origin feature/my-new-feature
   ```

6. **In your pull request**:
   - Provide a clear description of the changes
   - Link to related issues
   - Include screenshots/examples if applicable
   - Explain any breaking changes

### Development Setup

#### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)
- Docker (for testing containerized builds)

#### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/newrelic_exporter.git
cd newrelic_exporter

# Install dependencies (once Go modules are set up)
go mod download

# Build the project
make
# or
go build -o newrelic_exporter .

# Run tests
go test -v ./...

# Run the exporter
cp newrelic_exporter.yml.example newrelic_exporter.yml
# Edit the config file with your settings
./newrelic_exporter
```

### Code Style

- Follow standard Go conventions and idiomatic Go code
- Run `gofmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and reasonably sized

### Testing

- Add tests for new functionality
- Maintain or improve test coverage
- Include both unit tests and integration tests where applicable
- Test edge cases and error conditions

Example test:
```go
func TestMyNewFeature(t *testing.T) {
    // Setup
    input := "test input"

    // Execute
    result := MyNewFunction(input)

    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Add/update code comments for maintainers
- Update configuration examples if adding new options
- Consider adding examples for new features

## Code Review Process

1. **All contributions** go through code review
2. **Maintainers** will review your PR and may request changes
3. **CI/CD** must pass (tests, linting, builds)
4. **Approval** from at least one maintainer is required
5. **Merge** will be performed by a maintainer

## Release Process

Releases are managed by project maintainers:

1. Version tags follow semantic versioning (v1.0.0, v1.1.0, etc.)
2. Releases are automated via CI/CD
3. Docker images are automatically published to Docker Hub
4. GitHub releases include compiled binaries

## Areas for Contribution

Looking for ways to contribute? Here are some areas that need attention:

### High Priority
- [ ] Expand test coverage (especially integration tests)
- [ ] Add support for additional New Relic services (mobile, browser, etc.)
- [ ] Performance optimization for large metric sets
- [ ] Improved error handling and logging

### Medium Priority
- [ ] Add Grafana dashboard examples
- [ ] Support for New Relic Insights API
- [ ] Metrics for exporter health/status
- [ ] Configuration validation improvements

### Documentation
- [ ] Add more real-world configuration examples
- [ ] Create troubleshooting guide
- [ ] Add architecture documentation
- [ ] Video tutorials or walkthroughs

### Nice to Have
- [ ] Web UI for configuration
- [ ] Metric discovery/browsing tool
- [ ] Support for custom metric transformations
- [ ] Rate limiting for API calls

## Getting Help

If you need help with your contribution:

- **GitHub Issues**: Ask questions in issue comments
- **GitHub Discussions**: Start a discussion for broader topics
- **Documentation**: Check the README and configuration examples
- **Code**: Look at existing code for patterns and examples

## Recognition

Contributors will be:
- Credited in pull request history
- Mentioned in release notes for significant contributions
- Added to a CONTRIBUTORS file (if one is created)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

## Thank You!

Every contribution, no matter how small, helps make this project better. We appreciate your time and effort!

## Questions?

If you have questions about contributing, feel free to:
- Open an issue with the `question` label
- Start a discussion on GitHub Discussions
- Reach out to the maintainers

Happy contributing! 🎉
