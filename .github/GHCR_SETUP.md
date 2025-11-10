# GitHub Container Registry (GHCR) Setup for Automated Image Publishing

This repository is configured to automatically build and push Docker images to GitHub Container Registry (ghcr.io) using GitHub Actions.

## Prerequisites

1. Repository admin access (for configuring permissions)
2. GitHub Actions enabled in your repository

## Setup Instructions

### 1. Enable GitHub Container Registry

GitHub Container Registry is automatically available for all GitHub repositories. No additional setup is required!

### 2. Configure Package Permissions

After the first successful push, you may need to configure package visibility:

1. Go to your GitHub profile or organization
2. Navigate to **Packages**
3. Find the `newrelic-exporter` package
4. Click on **Package settings**
5. Under **Danger Zone**, you can:
   - Change package visibility (Public/Private)
   - Link the package to your repository
   - Manage access permissions

### 3. Verify the Workflow

The workflow automatically uses `GITHUB_TOKEN` (no secrets required!) and will:

- **On push to `main` or `master` branch:**
  - Build the Docker image
  - Push it to GHCR with the `latest` tag
  - Tag it with the branch name and git SHA

- **On git tags (e.g., `v1.0.0`):**
  - Build the Docker image
  - Push it with semantic version tags:
    - `1.0.0` (exact version)
    - `1.0` (major.minor)
    - `1` (major)
    - `latest` (if pushed to default branch)

- **On pull requests:**
  - Build the Docker image (but do not push)
  - Validate that the Docker build succeeds

### 4. Testing the Setup

To test the setup:

1. Make a commit to the `main` or `master` branch
2. Go to the **Actions** tab in GitHub
3. Watch the "Docker Build and Push" or "CI/CD Pipeline" workflow run
4. Once complete, check your repository's **Packages** section for the new image

### 5. Creating Releases

To create a versioned release:

```bash
# Tag the commit
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag
git push origin v1.0.0
```

This will trigger the workflow and create multiple version tags on GHCR.

## Workflow Features

- **Multi-architecture builds:** Builds for both `linux/amd64` and `linux/arm64`
- **Build caching:** Uses GitHub Actions cache to speed up builds
- **Automatic tagging:** Smart tagging based on branches, PRs, and git tags
- **Security:** Uses `GITHUB_TOKEN` (automatically provided, no secrets to manage!)
- **Integrated:** Images appear in your repository's Packages tab

## Using the Published Images

Pull the image from GHCR:

```bash
# Pull latest
docker pull ghcr.io/mrf/newrelic-exporter:latest

# Pull specific version
docker pull ghcr.io/mrf/newrelic-exporter:1.0.0
```

For private packages, authenticate first:

```bash
# Create a Personal Access Token with read:packages scope
# Then login:
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

## Updating the Docker Image Name

If you need to change the GHCR repository name, edit the workflows:

**`.github/workflows/docker-publish.yml`:**
```yaml
env:
  IMAGE_NAME: ghcr.io/your-username/your-repository-name
```

**`.github/workflows/ci-cd.yml`:**
```yaml
env:
  DOCKER_IMAGE: ghcr.io/your-username/your-repository-name
```

## Troubleshooting

### Error: "denied: permission_denied"

- Verify that the workflow has `packages: write` permission in the job definition
- Check that GitHub Actions is enabled for your repository
- Ensure you haven't disabled package creation in organization settings

### Build fails

- Check the Actions tab for detailed error logs
- Verify the Dockerfile is valid: `docker build -t test .`
- Ensure all dependencies are available

### Image not appearing in Packages

- Verify you're not building from a pull request (PRs don't push)
- Check that you're pushing to the correct branch (main/master)
- Verify the workflow completed successfully
- Check the repository's **Packages** tab (may take a moment to appear)

### Package is not linked to repository

- Go to Package settings
- Scroll to "Connect repository"
- Select your repository from the dropdown

## Benefits of GHCR vs Docker Hub

- **No secrets required:** Uses `GITHUB_TOKEN` automatically
- **Integrated:** Images appear directly in your repository
- **Unlimited private images:** For private repositories
- **Same infrastructure:** Hosted alongside your code
- **Fine-grained permissions:** Per-package access control
- **Free for public repositories:** No rate limits for public images

## Additional Resources

- [GitHub Container Registry Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
