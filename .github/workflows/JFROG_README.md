# JFrog Publishing Workflow

This document describes the GitHub Actions workflow for building and publishing Karpenter container images and Helm charts to JFrog Artifactory.

## Overview

The `jfrog-publish.yaml` workflow automatically builds and publishes:
- Container images for the Karpenter controller
- Helm charts (karpenter and karpenter-crd)

## Triggers

The workflow is triggered on:
- Push to `main` branch
- Push to `release-v*` branches
- Push of version tags (e.g., `v0.1.0`, `v0.1.0-rc.1`, etc.)
- Manual trigger via `workflow_dispatch`

## Required Secrets

The following GitHub secrets must be configured in your repository:

| Secret Name | Description | Example |
|------------|-------------|---------|
| `JFROG_REGISTRY` | JFrog Artifactory registry URL | `mycompany.jfrog.io` |
| `JFROG_REPO` | Docker repository name in JFrog | `karpenter-docker` |
| `JFROG_HELM_REPO` | Helm chart repository path in JFrog | `helm-local` |
| `JFROG_USERNAME` | JFrog username for authentication | `github-actions` |
| `JFROG_PASSWORD` | JFrog password or API token | `<your-api-token>` |

### Setting Up Secrets

1. Go to your repository on GitHub
2. Navigate to Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Add each of the required secrets listed above

## JFrog Setup

### Prerequisites

1. **JFrog Artifactory Instance**: You need access to a JFrog Artifactory instance
2. **Docker Repository**: Create a Docker repository in JFrog (e.g., `karpenter-docker`)
3. **Helm Repository**: Create a Helm repository in JFrog (e.g., `helm-local`)
4. **Service Account**: Create a service account with permissions to push images and charts

### Creating a Service Account

1. Log in to JFrog Artifactory
2. Go to Administration → Identity and Access → Users
3. Click "New User"
4. Create a user (e.g., `github-actions`)
5. Set a password or generate an API token
6. Assign appropriate permissions:
   - Deploy/Cache permissions for Docker repository
   - Deploy/Cache permissions for Helm repository

### Repository Configuration

#### Docker Repository
- Type: Docker
- Repository Key: `karpenter-docker` (or your chosen name)
- Package Type: Docker

#### Helm Repository
- Type: Generic or Helm
- Repository Key: `helm-local` (or your chosen name)
- Package Type: Helm

## Workflow Steps

1. **Checkout**: Fetches the repository code with full history
2. **Install Dependencies**: Sets up Go, ko, and other build tools
3. **Docker Buildx**: Configures Docker for multi-platform builds
4. **Login to JFrog**: Authenticates with JFrog Artifactory
5. **Determine Version**: Extracts version from git tag or commit SHA
6. **Build and Push Image**: Uses `ko` to build and push the controller image
7. **Build and Push Helm Charts**: Packages and uploads Helm charts
8. **Summary**: Displays build results in workflow summary

## Version Tagging

- **Tags**: When pushing a tag (e.g., `v0.1.0`), that version is used
- **Branches**: When pushing to branches, the short commit SHA is used as the version

## Image Naming

Images are pushed to: `<JFROG_REGISTRY>/<JFROG_REPO>/controller:<version>`

Example: `mycompany.jfrog.io/karpenter-docker/controller:v0.1.0`

## Helm Chart Naming

Helm charts are uploaded to: `<JFROG_REGISTRY>/<JFROG_HELM_REPO>/karpenter-<version>.tgz`

Example: `mycompany.jfrog.io/helm-local/karpenter-0.1.0.tgz`

## Testing the Workflow

1. Set up all required secrets in your repository
2. Push a commit to the main branch or create a tag
3. Navigate to Actions tab in GitHub
4. Monitor the "Build and Push to JFrog" workflow
5. Verify the image and charts are available in JFrog

## Troubleshooting

### Authentication Errors

If you see authentication errors:
- Verify `JFROG_USERNAME` and `JFROG_PASSWORD` are correct
- Ensure the service account has necessary permissions
- Check if the API token hasn't expired

### Image Push Failures

If image push fails:
- Verify `JFROG_REGISTRY` and `JFROG_REPO` are correct
- Ensure the Docker repository exists in JFrog
- Check repository permissions

### Helm Push Failures

If Helm chart push fails:
- Verify `JFROG_HELM_REPO` path is correct
- Ensure the Helm repository exists in JFrog
- Check that the service account has deploy permissions

## Security Considerations

- Use API tokens instead of passwords when possible
- Rotate credentials regularly
- Use repository-specific service accounts with minimal permissions
- Enable audit logging in JFrog to track deployments

## Additional Resources

- [JFrog Artifactory Documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation)
- [ko Documentation](https://ko.build/)
- [GitHub Actions Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
