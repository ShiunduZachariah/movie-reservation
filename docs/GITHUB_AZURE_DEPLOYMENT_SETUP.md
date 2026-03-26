# GitHub Actions Azure Deployment Setup

## Purpose

This guide explains how to configure this repository so GitHub Actions can:

- deploy Azure infrastructure with Bicep
- build the backend on every push
- build and publish the backend image to GitHub Container Registry
- deploy the backend image to Azure Container Apps
- build and publish the Azure Functions package

This setup uses GitHub Container Registry (`ghcr.io`) for the API image instead of Azure Container Registry.

## What Was Added

The repo now includes:

- [Dockerfile](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/Dockerfile)
- [deploy-infra.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.github/workflows/deploy-infra.yml)
- [deploy-api.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.github/workflows/deploy-api.yml)
- [deploy-functions.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.github/workflows/deploy-functions.yml)

## Azure Authentication Model

These workflows use GitHub OIDC with `azure/login@v2`.

Create a Microsoft Entra application or user-assigned identity with a federated credential for this GitHub repository, then add these repository secrets:

- `AZURE_CLIENT_ID`
- `AZURE_TENANT_ID`
- `AZURE_SUBSCRIPTION_ID`

Official references:

- https://learn.microsoft.com/en-us/azure/developer/github/connect-from-azure-openid-connect
- https://github.com/Azure/login

## GitHub Secrets

Create these repository secrets in GitHub:

### Azure auth

- `AZURE_CLIENT_ID`
- `AZURE_TENANT_ID`
- `AZURE_SUBSCRIPTION_ID`

### Infrastructure

- `AZURE_POSTGRES_ADMIN_PASSWORD`

### Container Apps image pull from GHCR

- `GHCR_USERNAME`
- `GHCR_PAT`

Use a GitHub personal access token with at least:

- `read:packages`

If you also want to use the same PAT outside Actions for manual pushes or tests, `write:packages` is also useful.

### Function app secrets

- `AZURE_WEBJOBS_STORAGE`
- `AZURE_FUNCTION_DATABASE_URL`
- `AZURE_RESEND_API_KEY`
- `AZURE_RESEND_FROM`

## GitHub Repository Variables

Create these repository variables in GitHub:

- `AZURE_RESOURCE_GROUP`
- `AZURE_LOCATION`
- `AZURE_PREFIX`
- `AZURE_POSTGRES_ADMIN_LOGIN`
- `AZURE_STORAGE_CONTAINER_NAME`
- `AZURE_TICKET_EMAIL_QUEUE_NAME`
- `AZURE_CONTAINER_APP_NAME`
- `AZURE_FUNCTION_APP_NAME`

Recommended example values:

- `AZURE_RESOURCE_GROUP=rg-movie-reservation-prod`
- `AZURE_LOCATION=eastus`
- `AZURE_PREFIX=moviereservationprod`
- `AZURE_POSTGRES_ADMIN_LOGIN=postgresadmin`
- `AZURE_STORAGE_CONTAINER_NAME=movie-posters`
- `AZURE_TICKET_EMAIL_QUEUE_NAME=ticket-emails`
- `AZURE_CONTAINER_APP_NAME=movie-reservation-api`
- `AZURE_FUNCTION_APP_NAME=movie-reservation-functions`

## GitHub Container Registry Setup

The API workflow pushes images to:

- `ghcr.io/<github-owner>/movie-reservation-api:<git-sha>`

The workflow uses:

- `github.actor`
- `secrets.GITHUB_TOKEN`

to push the image from GitHub Actions.

Important:

Azure Container Apps still needs credentials to pull the image if the package is private. That is why `GHCR_USERNAME` and `GHCR_PAT` are used in the deployment step.

Official GHCR and Container Apps guidance:

- https://learn.microsoft.com/en-us/azure/container-apps/github-actions

## Azure Resources You Still Need

The workflows assume these Azure resources already exist or will be created by the infra workflow:

- resource group
- PostgreSQL Flexible Server
- storage account
- blob container
- storage queue
- key vault
- log analytics workspace

These resources are still expected to exist before the API and Functions deployment workflows can fully succeed:

- Azure Container Apps environment
- Azure Container App for the API
- Azure Function App for the functions package

The current Bicep does not yet create the Container Apps environment, Container App, or Function App host.

## Suggested Setup Order

1. Create the GitHub secrets and variables.
2. Create the OIDC federated credential in Azure for this repo.
3. Run `Deploy Azure Infrastructure`.
4. Create the Azure Container Apps environment and API app.
5. Create the Azure Function App.
6. Run `Build and Deploy API`.
7. Run `Build and Deploy Functions`.

## Container App Registry Access

The API workflow runs:

```bash
az containerapp registry set --server ghcr.io ...
```

This configures the Container App to pull the private GHCR image.

If you later make the image public, this step can be simplified, but Microsoft still recommends explicitly configuring the registry server for non-ACR registries.

## Function App Notes

The functions deployment now packages a Linux Go custom handler binary and deploys:

- `host.json`
- `send-ticket-email/function.json`
- `expire-reservations/function.json`
- the email template
- the compiled `handler` executable

Important:

- the Function App must be configured to run custom handlers
- `FUNCTIONS_WORKER_RUNTIME=Custom` is set by the workflow
- the queue trigger uses `TICKET_EMAIL_QUEUE_NAME`

Official references:

- https://learn.microsoft.com/en-us/azure/azure-functions/functions-custom-handlers
- https://learn.microsoft.com/en-us/azure/azure-functions/functions-how-to-github-actions

## What The Workflows Do

### `deploy-infra.yml`

- logs into Azure with OIDC
- creates the resource group if needed
- deploys `azure/main.bicep`
- passes the PostgreSQL password and storage names as parameters
- skips Azure Container Registry creation for now

### `deploy-api.yml`

- builds the backend Docker image
- pushes it to GHCR with a SHA tag
- configures Container Apps to pull from GHCR
- deploys the new image to the existing Container App

### `deploy-functions.yml`

- builds the custom handler binary for Linux
- assembles a deployable Function App package
- sets required Function App settings
- deploys the package using `Azure/functions-action`

## Remaining Azure Gaps

These workflows are now in place, but the repo still needs full host resource automation if you want one-click provisioning:

- Bicep module for Container Apps environment
- Bicep module for API Container App
- Bicep module for Azure Function App
- Key Vault reference wiring into runtime app settings

## Verification Checklist

After setup, verify:

- GitHub Actions can assume Azure identity through OIDC
- the infra workflow creates or updates Azure resources
- the API workflow pushes a new image to GHCR
- Azure Container Apps picks up the new revision
- the functions workflow publishes successfully
- `send-ticket-email` receives queue-triggered invocations
- `expire-reservations` runs on schedule
