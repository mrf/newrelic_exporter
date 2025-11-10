# Docker Hub Setup for Automated Image Publishing

This repository is configured to automatically build and push Docker images to Docker Hub using GitHub Actions.

## Prerequisites

1. A Docker Hub account
2. Repository admin access to configure GitHub secrets

## Setup Instructions

### 1. Create a Docker Hub Access Token

1. Log in to [Docker Hub](https://hub.docker.com/)
2. Click on your username in the top right corner
3. Select **Account Settings** → **Security**
4. Click **New Access Token**
5. Give it a description (e.g., "GitHub Actions - newrelic_exporter")
6. Set the access permissions to **Read, Write, Delete** (or at minimum **Read, Write**)
7. Click **Generate**
8. **Important:** Copy the token immediately - you won't be able to see it again!

### 2. Add GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add the following secrets:

   **DOCKERHUB_USERNAME**
   - Value: Your Docker Hub username

   **DOCKERHUB_TOKEN**
   - Value: The access token you generated in step 1

### 3. Verify the Workflow

Once the secrets are configured, the workflow will automatically:

- **On push to `main` or `master` branch:**
  - Build the Docker image
  - Push it to Docker Hub with the `latest` tag
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
3. Watch the "Docker Build and Push" workflow run
4. Once complete, check your Docker Hub repository for the new image

### 5. Creating Releases

To create a versioned release:

```bash
# Tag the commit
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag
git push origin v1.0.0
```

This will trigger the workflow and create multiple version tags on Docker Hub.

## Workflow Features

- **Multi-architecture builds:** Builds for both `linux/amd64` and `linux/arm64`
- **Build caching:** Uses GitHub Actions cache to speed up builds
- **Automatic tagging:** Smart tagging based on branches, PRs, and git tags
- **Security:** Uses Docker Hub access tokens (not passwords)

## Updating the Docker Image Name

If you need to change the Docker Hub repository name, edit `.github/workflows/docker-publish.yml`:

```yaml
env:
  IMAGE_NAME: your-dockerhub-username/your-repository-name
```

## Troubleshooting

### Error: "unauthorized: authentication required"

- Verify that `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets are set correctly
- Check that the access token has the correct permissions
- Ensure the token hasn't expired

### Build fails

- Check the Actions tab for detailed error logs
- Verify the Dockerfile is valid: `docker build -t test .`
- Ensure all dependencies are available

### Image not appearing on Docker Hub

- Verify you're not building from a pull request (PRs don't push)
- Check that you're pushing to the correct branch (main/master)
- Verify the Docker Hub repository exists or that you have permissions to create it

## Manual Docker Build and Push

If you need to manually build and push:

```bash
# Build the image
docker build -t mrf/newrelic-exporter:latest .

# Log in to Docker Hub
docker login

# Push the image
docker push mrf/newrelic-exporter:latest
```

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Hub Access Tokens](https://docs.docker.com/docker-hub/access-tokens/)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
