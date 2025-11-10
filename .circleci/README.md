# CircleCI Integration

This repository includes a comprehensive CircleCI configuration for automated building, testing, and deployment.

## Features

The CircleCI pipeline includes:

1. **Build and Test**
   - Compiles the Go application
   - Runs all tests with race detection
   - Generates code coverage reports
   - Runs `go vet` for static analysis
   - Checks code formatting with `go fmt`
   - Stores test results and coverage reports as artifacts

2. **Docker Build**
   - Builds Docker images
   - Uses Docker layer caching for faster builds
   - Tags images appropriately

3. **Docker Push** (main/master and tags only)
   - Pushes images to Docker Hub
   - Tags with commit SHA, version numbers, and 'latest'
   - Supports semantic versioning (v1.0.0 → 1.0.0, 1.0, 1)

4. **GitHub Releases** (tags only)
   - Automatically creates GitHub releases for version tags
   - Attaches compiled binaries to releases

## Setup Instructions

### 1. Enable CircleCI for Your Repository

1. Go to [CircleCI](https://circleci.com/)
2. Sign in with your GitHub account
3. Navigate to **Projects** in the sidebar
4. Find your repository and click **Set Up Project**
5. CircleCI will automatically detect the `.circleci/config.yml` file
6. Click **Start Building**

### 2. Configure Environment Variables

You need to configure the following environment variables in CircleCI:

#### Option A: Using Project Environment Variables

1. In CircleCI, go to your project settings
2. Navigate to **Environment Variables**
3. Add the following variables:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Your Docker Hub access token
   - `GITHUB_TOKEN`: GitHub personal access token (for releases)

#### Option B: Using Context (Recommended)

Contexts allow you to share environment variables across multiple projects.

1. In CircleCI, go to **Organization Settings**
2. Navigate to **Contexts**
3. Click **Create Context** and name it `docker-hub`
4. Add the following environment variables to the context:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Your Docker Hub access token
   - `GITHUB_TOKEN`: GitHub personal access token

The config is already set up to use the `docker-hub` context for Docker push jobs.

### 3. Create Required Tokens

#### Docker Hub Access Token

1. Log in to [Docker Hub](https://hub.docker.com/)
2. Go to **Account Settings** → **Security**
3. Click **New Access Token**
4. Give it a name (e.g., "CircleCI")
5. Set permissions to **Read & Write**
6. Copy the token and save it as `DOCKERHUB_TOKEN` in CircleCI

#### GitHub Personal Access Token

1. Go to GitHub **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Click **Generate new token** → **Generate new token (classic)**
3. Give it a name (e.g., "CircleCI Releases")
4. Select scopes:
   - `repo` (Full control of private repositories)
5. Click **Generate token**
6. Copy the token and save it as `GITHUB_TOKEN` in CircleCI

## Workflows

### Pull Request Workflow

When you open a pull request:
- Runs build and test jobs
- Does NOT push Docker images or create releases
- Provides test results and coverage reports

### Main/Master Branch Workflow

When you push to `main` or `master`:
- Runs build and test jobs
- Builds Docker image
- Pushes Docker image to Docker Hub with:
  - `latest` tag
  - Commit SHA tag (e.g., `abc1234`)

### Release Workflow

When you create a git tag (e.g., `v1.0.0`):
- Runs build and test jobs
- Builds Docker image
- Pushes Docker image to Docker Hub with:
  - Full version tag (e.g., `1.0.0`)
  - Major.Minor tag (e.g., `1.0`)
  - Major tag (e.g., `1`)
  - Commit SHA tag
- Creates a GitHub release with the compiled binary

To create a release:

```bash
# Tag the commit
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag
git push origin v1.0.0
```

## CircleCI Configuration Details

### Jobs

- **build-and-test**: Compiles code, runs tests, performs static analysis
- **build-docker**: Builds Docker image (for branches)
- **push-docker**: Builds and pushes Docker image to Docker Hub
- **create-release**: Creates GitHub release with artifacts

### Executors

- **golang-executor**: Uses `cimg/go:1.21` for Go builds
- **docker/docker**: Uses Docker-in-Docker for container builds

### Caching

The pipeline uses caching to speed up builds:
- Go module cache (for dependencies)
- Docker layer cache (for Docker builds)

## Troubleshooting

### Build Fails on Test Job

- Check the test output in CircleCI
- Run tests locally: `go test -v ./...`
- Ensure all dependencies are properly declared

### Docker Push Fails

- Verify `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` are set correctly
- Check that the Docker Hub token has write permissions
- Ensure you're pushing to the correct branch (main/master)

### Release Creation Fails

- Verify `GITHUB_TOKEN` is set and has `repo` scope
- Check that the tag follows the `v*.*.*` pattern
- Ensure the GitHub CLI (`gh`) can authenticate

### "No go.sum file found" Error

The project needs to be migrated to Go modules first. See the separate issue for Go modules migration.

## Manual Trigger

You can manually trigger a workflow in CircleCI:

1. Go to your project in CircleCI
2. Click on a branch
3. Click **Trigger Pipeline**

## Comparison with GitHub Actions

This repository also includes GitHub Actions for CI/CD. Here's when to use each:

| Feature | CircleCI | GitHub Actions |
|---------|----------|----------------|
| **Better for** | Complex workflows, advanced caching | GitHub-native integrations |
| **Configuration** | `.circleci/config.yml` | `.github/workflows/*.yml` |
| **Free tier** | 6,000 build minutes/month | 2,000 build minutes/month |
| **Contexts** | Centralized secrets management | Repository/org secrets |
| **Best use case** | Organizations with multiple projects | Single GitHub repository |

You can use both systems simultaneously or choose one based on your needs.

## Additional Resources

- [CircleCI Documentation](https://circleci.com/docs/)
- [CircleCI Go Language Guide](https://circleci.com/docs/language-go/)
- [CircleCI Contexts](https://circleci.com/docs/contexts/)
- [Docker Hub Access Tokens](https://docs.docker.com/docker-hub/access-tokens/)
