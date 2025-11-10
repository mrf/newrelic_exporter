# Go Modules Migration

This project has been migrated from the legacy GOPATH-based dependency management to Go modules (go.mod).

## What Changed

### Before
- Dependencies were managed using `go get` and vendoring
- Required setting GOPATH
- Used `Makefile.COMMON` to download Go version and manage dependencies
- No explicit dependency versioning

### After
- Dependencies are managed via `go.mod` and `go.sum`
- Works with any Go 1.21+ installation
- Reproducible builds with locked dependency versions
- Simpler project setup and contribution process

## Benefits

1. **Reproducible Builds**: Dependencies are locked to specific versions in `go.sum`
2. **Easier Contribution**: Contributors don't need to set up GOPATH
3. **Version Control**: Explicit dependency versions in `go.mod`
4. **Better Tooling**: Modern Go tools work better with modules
5. **Simpler CI/CD**: No need for special GOPATH setup in pipelines

## Module Information

**Module Path**: `github.com/mrf/newrelic_exporter`

**Go Version**: 1.21+

## Dependencies

### Direct Dependencies

- **github.com/antonholmquist/jason** v1.0.0 - JSON parsing
- **github.com/prometheus/client_golang** v1.11.1 - Prometheus client library
- **github.com/prometheus/log** v0.0.0-20151026012452-9a3136781e1f - Logging
- **github.com/tomnomnom/linkheader** v0.0.0-20180220141516-dd9dcf9c3b8b - HTTP Link header parsing
- **gopkg.in/yaml.v2** v2.4.0 - YAML parsing

## Usage

### Building

```bash
# Simple build (using go modules)
go build

# Or using make
make
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Using make
make test
```

### Adding Dependencies

```bash
# Add a new dependency
go get github.com/some/package@version

# Update dependencies
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy
```

### Updating Dependencies

```bash
# Update a specific dependency
go get github.com/some/package@latest

# Update all dependencies
go get -u ./...
go mod tidy
```

### Vendoring (Optional)

If you want to vendor dependencies:

```bash
# Create vendor directory
go mod vendor

# Build using vendor
go build -mod=vendor
```

## For Contributors

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/mrf/newrelic_exporter.git
   cd newrelic_exporter
   ```

2. Dependencies are automatically downloaded:
   ```bash
   go build
   ```

That's it! No GOPATH setup needed.

### Adding Features

1. Make your changes
2. If you add new dependencies, they'll be automatically added to `go.mod`:
   ```bash
   go get github.com/new/dependency
   ```
3. Run tests:
   ```bash
   go test ./...
   ```
4. Tidy up:
   ```bash
   go mod tidy
   ```
5. Commit both code changes and `go.mod`/`go.sum` changes

## Troubleshooting

### "Cannot find package" errors

Run:
```bash
go mod download
go mod tidy
```

### Dependency version conflicts

Check `go.mod` and update conflicting dependencies:
```bash
go get github.com/conflicting/package@latest
go mod tidy
```

### Building fails with "replace" directives

The project no longer uses replace directives. If you see errors about them:
```bash
# Clean module cache
go clean -modcache

# Re-download
go mod download
```

## Migration Notes

### Compatibility

- **Minimum Go version**: 1.21
- **Recommended**: Latest stable Go version

### Breaking Changes

None - the API and functionality remain unchanged. Only the build process has been modernized.

### Legacy Build System

The old `Makefile.COMMON` is retained for reference but is no longer used. The simpler `Makefile` now uses standard `go` commands.

## CI/CD Integration

### GitHub Actions

```yaml
- name: Set up Go
  uses: actions/setup-go@v4
  with:
    go-version: '1.21'

- name: Download dependencies
  run: go mod download

- name: Build
  run: go build

- name: Test
  run: go test -v ./...
```

### CircleCI

```yaml
- run:
    name: Download dependencies
    command: go mod download

- run:
    name: Build
    command: go build

- run:
    name: Test
    command: go test -v ./...
```

## Additional Resources

- [Official Go Modules Documentation](https://go.dev/doc/modules/managing-dependencies)
- [Go Modules Wiki](https://github.com/golang/go/wiki/Modules)
- [Go Modules Tutorial](https://go.dev/blog/using-go-modules)

## Questions?

If you encounter issues with Go modules:
- Check that you're using Go 1.21 or later: `go version`
- Try clearing the module cache: `go clean -modcache`
- Re-download dependencies: `go mod download`
- Ask in GitHub Issues or Discussions
