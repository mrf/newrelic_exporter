# Continuous Integration and Deployment

This project uses GitHub Actions for automated CI/CD with standard GitHub-hosted runners.

## Overview

The CI/CD pipeline automatically:
- ✅ Runs tests on every push and pull request
- ✅ Checks code formatting and quality
- ✅ Builds binaries for multiple platforms
- ✅ Creates and pushes Docker images
- ✅ Generates releases with compiled binaries
- ✅ Performs security scans

## Workflows

### Main Workflow: `ci-cd.yml`

**Triggers:**
- Push to `main` or `master` branches
- Pull requests to `main` or `master`
- Git tags matching `v*.*.*` (e.g., v1.0.0)

**Jobs:**

#### 1. Test Job
Runs on: `ubuntu-latest`

- Checks out code
- Sets up Go 1.21
- Downloads and verifies dependencies
- Runs `go vet` for static analysis
- Checks code formatting with `gofmt`
- Runs all tests with race detection and coverage
- Uploads coverage to Codecov
- Stores test artifacts

#### 2. Build Job
Runs on: `ubuntu-latest` (after tests pass)

Builds binaries for multiple platforms:
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64
- **Windows**: amd64

Creates compressed archives:
- `.tar.gz` for Linux and macOS
- `.zip` for Windows

Uploads all binaries as artifacts.

#### 3. Docker Job
Runs on: `ubuntu-latest` (after tests pass, only on push)

- Builds multi-architecture Docker images (amd64, arm64)
- Pushes to Docker Hub with appropriate tags:
  - `latest` for main/master branch
  - Version tags for git tags (e.g., `1.0.0`, `1.0`, `1`)
  - SHA tags for tracking
  - Branch name tags

Uses build cache for faster builds.

#### 4. Release Job
Runs on: `ubuntu-latest` (only for version tags)

- Downloads all build artifacts
- Creates GitHub Release
- Attaches compiled binaries for all platforms
- Auto-generates release notes

#### 5. Security Job
Runs on: `ubuntu-latest`

- Scans code for vulnerabilities using Trivy
- Uploads results to GitHub Security tab
- Runs on every push and PR

## GitHub-Hosted Runners

All jobs use standard GitHub-hosted runners:

**ubuntu-latest:**
- 2-core CPU
- 7 GB RAM
- 14 GB SSD storage
- Ubuntu 22.04 LTS

**Benefits:**
- ✅ No setup required
- ✅ Always up-to-date
- ✅ Free for public repositories
- ✅ Reliable and fast
- ✅ Managed by GitHub

## Configuration

### Required Secrets

Set these in **Settings → Secrets and variables → Actions**:

| Secret | Description | Required For |
|--------|-------------|--------------|
| `DOCKERHUB_USERNAME` | Docker Hub username | Docker publishing |
| `DOCKERHUB_TOKEN` | Docker Hub access token | Docker publishing |
| `GITHUB_TOKEN` | Automatically provided | Releases |

### Docker Hub Setup

1. Log in to [Docker Hub](https://hub.docker.com/)
2. Go to **Account Settings → Security**
3. Click **New Access Token**
4. Name: "GitHub Actions - newrelic_exporter"
5. Permissions: **Read & Write**
6. Copy the token
7. Add to GitHub repository secrets as `DOCKERHUB_TOKEN`

### Codecov (Optional)

Coverage reporting works automatically for public repositories. For private repositories:

1. Sign up at [Codecov](https://codecov.io/)
2. Add your repository
3. Get the upload token
4. Add as `CODECOV_TOKEN` secret (optional, workflow continues on error)

## Usage

### Running Tests Locally

Before pushing, run tests locally:

```bash
# Run all tests
go test -v ./...

# Run with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# Check formatting
gofmt -l .

# Run static analysis
go vet ./...

# Or use make
make check
```

### Creating a Release

1. **Create and push a tag:**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Run all tests
   - Build binaries for all platforms
   - Create multi-arch Docker images
   - Push images with version tags
   - Create GitHub Release with binaries

3. **Find your release at:**
   - GitHub: `https://github.com/mrf/newrelic_exporter/releases`
   - Docker Hub: `https://hub.docker.com/r/mrf/newrelic-exporter`

### Pull Request Workflow

When you open a PR:

1. **Automatic checks run:**
   - Code formatting
   - Static analysis
   - All tests with race detection
   - Security scan

2. **View results:**
   - Check the "Checks" tab in your PR
   - View detailed logs for any failures
   - Coverage report is commented (if configured)

3. **Required status checks:**
   - Test job must pass
   - All checks must be green before merge

### Branch Workflow

**On push to main/master:**
- Runs all tests
- Builds binaries
- Builds and pushes Docker image with `latest` tag
- Runs security scan

**On push to feature branch:**
- Runs all tests
- Runs security scan
- Does not build Docker images

## Caching

The pipeline uses caching to speed up builds:

### Go Module Cache
- Location: `~/go/pkg/mod`
- Key: Based on `go.sum` checksum
- Speeds up dependency downloads

### Docker Build Cache
- Type: GitHub Actions cache
- Reuses layers across builds
- Significantly faster builds

### Typical Build Times

| Job | First Run | Cached Run |
|-----|-----------|------------|
| Test | ~2-3 min | ~30-60 sec |
| Build | ~1-2 min | ~30-45 sec |
| Docker | ~5-8 min | ~2-3 min |

## Monitoring

### GitHub Actions UI

View workflow runs:
1. Go to repository on GitHub
2. Click **Actions** tab
3. Select a workflow run
4. View logs, artifacts, and status

### Status Badges

Add to README:

```markdown
![CI/CD](https://github.com/mrf/newrelic_exporter/actions/workflows/ci-cd.yml/badge.svg)
```

### Notifications

Configure in **Settings → Notifications**:
- Email on workflow failure
- Slack/Discord webhooks (via GitHub Apps)

## Artifacts

Build artifacts are available for 90 days:

- **Test results**: Coverage reports
- **Binaries**: Compiled executables for all platforms
- **Docker images**: Published to Docker Hub

Download artifacts from workflow run page.

## Troubleshooting

### Build Failing

**Check the logs:**
1. Go to Actions tab
2. Click on the failed run
3. Click on the failed job
4. Expand the failed step

**Common issues:**

| Error | Solution |
|-------|----------|
| Tests fail | Run `go test ./...` locally |
| Formatting check fails | Run `gofmt -w .` |
| `go vet` fails | Fix static analysis issues |
| Docker build fails | Test with `docker build .` locally |
| Module issues | Run `go mod tidy` |

### Docker Push Fails

**Check:**
- `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` are set correctly
- Docker Hub token has write permissions
- Token hasn't expired

**Regenerate token:**
1. Log in to Docker Hub
2. Go to Account Settings → Security
3. Delete old token
4. Create new token
5. Update GitHub secret

### Release Not Created

**Check:**
- Tag format is `v*.*.*` (e.g., v1.0.0)
- All jobs completed successfully
- You have write permissions
- Tag was pushed: `git push origin v1.0.0`

### Slow Builds

**Optimization tips:**
- Ensure caching is working (check cache hit/miss in logs)
- Reduce dependencies if possible
- Use `go mod vendor` for fully offline builds

## Advanced Configuration

### Custom Runners

To use self-hosted runners:

```yaml
jobs:
  test:
    runs-on: self-hosted  # or [self-hosted, linux, x64]
```

### Matrix Strategy

Current matrix builds for:
- Linux: amd64, arm64
- macOS: amd64, arm64
- Windows: amd64

To add more platforms:

```yaml
strategy:
  matrix:
    include:
      - goos: linux
        goarch: 386
```

### Conditional Jobs

Jobs can run conditionally:

```yaml
if: github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
```

## Security

### Secrets Management

- Never commit secrets to code
- Use GitHub Secrets for sensitive data
- Rotate tokens periodically
- Use least-privilege access

### Dependency Security

- Trivy scanner runs on every build
- Results uploaded to Security tab
- Dependabot alerts enabled (configure in Settings)

### Supply Chain Security

- Builds use pinned action versions (@v4)
- Go module checksums verified
- Docker images use specific base image versions

## Costs

**For public repositories:**
- ✅ Unlimited minutes on standard runners
- ✅ Free storage for artifacts (90 days)
- ✅ Free Docker Hub pulls

**For private repositories:**
- 2,000 minutes/month free (then $0.008/min)
- Consider usage limits in Settings

## Comparison: GitHub Actions vs CircleCI

| Feature | GitHub Actions | CircleCI |
|---------|----------------|----------|
| Setup | Built-in | External service |
| Authentication | Automatic | API token required |
| Free tier | 2,000 min/month (private) | 6,000 min/month |
| Configuration | YAML in `.github/` | YAML in `.circleci/` |
| Runners | GitHub-hosted or self-hosted | CircleCI cloud or self-hosted |
| Integration | Native GitHub | Via API |
| Secrets | GitHub Secrets | CircleCI Environment Variables |
| Artifacts | 90 days retention | 30 days retention |

**Why we chose GitHub Actions:**
- ✅ Native GitHub integration
- ✅ No external service dependency
- ✅ Simpler setup and maintenance
- ✅ Better for open source projects
- ✅ Standard runners work great

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Go Setup Action](https://github.com/actions/setup-go)
- [GitHub Hosted Runners](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners)

## Support

For issues with CI/CD:
- Check workflow logs in Actions tab
- Review this documentation
- Open an issue on GitHub
- Check [GitHub Actions Community](https://github.community/c/code-to-cloud/github-actions/41)
