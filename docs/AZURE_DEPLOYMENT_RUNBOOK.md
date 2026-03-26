# Azure Deployment Runbook

## Purpose

This runbook explains how to deploy the `movie-reservation` service to Azure using the infrastructure and application code currently present in this repository.

It is written against the repo state as of March 26, 2026.

## Scope

This runbook covers:

- base Azure infrastructure deployment with Bicep
- backend API deployment approach
- Azure Functions deployment approach
- secret and configuration setup
- database migration execution
- smoke tests and rollback guidance

This runbook also calls out the current gaps in the repo that still require manual work before a full production deployment is possible.

## Target Azure Architecture

The intended production shape for this project is:

- Azure Database for PostgreSQL Flexible Server
- Azure Storage Account
- Azure Queue Storage for ticket email messages
- Azure Blob Storage for poster uploads
- Azure Container Registry
- backend API hosted as a containerized app
- Azure Functions for:
  - `send-ticket-email`
  - `expire-reservations`
- Key Vault for secrets
- Log Analytics workspace for monitoring

## Current Automation Status

Automated today in `azure/main.bicep`:

- PostgreSQL Flexible Server
- Azure Container Registry
- Storage Account
- Key Vault
- Log Analytics Workspace

Not automated yet:

- Container App or App Service for the backend API
- Function App hosting resources
- storage queue creation
- blob container creation
- managed identity wiring
- Key Vault secret references in runtime resources
- deployment pipeline

## Important Current Gaps

Before relying on this in production, account for these repo gaps:

1. No backend `Dockerfile` exists yet.
2. No Bicep module exists yet for the backend host.
3. No Bicep module exists yet for the Function App host.
4. `azure/modules/postgres.bicep` currently hardcodes `administratorLoginPassword: 'replace-me'` and must be parameterized before serious use.
5. `azure/modules/storage.bicep` creates only the storage account. It does not create the queue or blob container the app expects.
6. `functions/send-ticket-email/function.json` is bound to queue name `ticket-emails`, which must match your deployed queue name exactly.
7. `functions/send-ticket-email/main.go` currently hardcodes `From: "CineReserve <tickets@yourdomain.com>"`, which will fail unless that sender is verified in Resend.
8. The backend uses Azure Queue only when `CINERESERVE_PRIMARY.ENV=production` and queue config is present. In development it uses Asynq + Redis instead.

## Prerequisites

Install locally:

- Azure CLI
- Bicep CLI
- Docker Desktop
- Go 1.25+
- Azure Functions Core Tools

Azure permissions required:

- permission to create resource groups
- permission to deploy Bicep templates
- permission to create Key Vault secrets
- permission to create and configure PostgreSQL Flexible Server
- permission to create and configure Storage resources
- permission to create and configure a backend host and Function App host

## Naming Assumptions

The repo Bicep currently uses a `prefix` parameter, defaulting to `cinereserve`.

With that prefix, current resource names resolve like this:

- PostgreSQL: `cinereserve-pg`
- ACR: `cinereserveacr`
- Key Vault: `cinereserve-kv`
- Log Analytics: `cinereserve-logs`
- Storage Account: `cinereservestorage`

Pick a prefix that is globally unique where required, especially for:

- storage account names
- container registry names
- key vault names

## Required Configuration

### Backend environment variables

These must be available to the deployed API:

- `CINERESERVE_PRIMARY.ENV=production`
- `CINERESERVE_SERVER.PORT`
- `CINERESERVE_SERVER.READ_TIMEOUT`
- `CINERESERVE_SERVER.WRITE_TIMEOUT`
- `CINERESERVE_SERVER.IDLE_TIMEOUT`
- `CINERESERVE_SERVER.CORS_ALLOWED_ORIGINS`
- `CINERESERVE_DATABASE.HOST`
- `CINERESERVE_DATABASE.PORT`
- `CINERESERVE_DATABASE.USER`
- `CINERESERVE_DATABASE.PASSWORD`
- `CINERESERVE_DATABASE.NAME`
- `CINERESERVE_DATABASE.SSL_MODE=require`
- `CINERESERVE_DATABASE.MAX_OPEN_CONNS`
- `CINERESERVE_DATABASE.MAX_IDLE_CONNS`
- `CINERESERVE_DATABASE.CONN_MAX_LIFETIME`
- `CINERESERVE_DATABASE.CONN_MAX_IDLE_TIME`
- `CINERESERVE_REDIS.ADDRESS`
- `CINERESERVE_AUTH.SECRET_KEY`
- `CINERESERVE_INTEGRATION.RESEND_API_KEY`
- `CINERESERVE_INTEGRATION.RESEND_FROM`
- `CINERESERVE_APP.BASE_URL`
- `CINERESERVE_APP.NAME`
- `CINERESERVE_AZURE.STORAGE_ACCOUNT_NAME`
- `CINERESERVE_AZURE.STORAGE_CONTAINER_NAME`
- `CINERESERVE_AZURE.STORAGE_QUEUE_NAME`
- `CINERESERVE_AZURE.STORAGE_CONNECTION_STRING`

### Function App environment variables

`send-ticket-email` needs:

- `RESEND_API_KEY`
- `RESEND_FROM`
- `AzureWebJobsStorage`

`expire-reservations` needs:

- `DATABASE_URL`

Both functions also need:

- `AzureWebJobsStorage`

## Recommended Secret Placement

Store these in Key Vault:

- postgres admin password
- application JWT secret
- Resend API key
- storage connection string
- function `DATABASE_URL`

Recommended secret names:

- `postgres-admin-password`
- `app-jwt-secret`
- `resend-api-key`
- `resend-from`
- `storage-connection-string`
- `function-database-url`

## Deployment Procedure

### 1. Log in and select the target subscription

```powershell
az login
az account set --subscription "<subscription-id-or-name>"
```

Verify:

```powershell
az account show --output table
```

### 2. Create the resource group

Set working variables:

```powershell
$RESOURCE_GROUP = "rg-cinereserve-prod"
$LOCATION = "eastus"
$PREFIX = "cinereserveprod"
```

Create the resource group:

```powershell
az group create --name $RESOURCE_GROUP --location $LOCATION
```

### 3. Review and patch Bicep before deployment

Before running the template, update these files:

- [main.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/main.bicep)
- [postgres.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/modules/postgres.bicep)
- [storage.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/modules/storage.bicep)

Required pre-deploy changes:

1. Parameterize the PostgreSQL admin password instead of using `replace-me`.
2. Add blob container creation for the poster container.
3. Add queue creation for the ticket email queue.
4. Make the queue name consistent with:
   - `CINERESERVE_AZURE.STORAGE_QUEUE_NAME`
   - `functions/send-ticket-email/function.json`

If you deploy without fixing those, infrastructure will come up partially but the app and functions will not be fully wired.

### 4. Deploy the base infrastructure

From the repo root:

```powershell
az deployment group create `
  --resource-group $RESOURCE_GROUP `
  --template-file .\azure\main.bicep `
  --parameters location=$LOCATION prefix=$PREFIX
```

### 5. Create missing storage resources

If the Bicep has not yet been extended to create these, create them manually.

Create blob container:

```powershell
az storage container create `
  --name "movie-posters" `
  --account-name "${PREFIX}storage" `
  --auth-mode login
```

Create queue:

```powershell
az storage queue create `
  --name "ticket-emails" `
  --account-name "${PREFIX}storage" `
  --auth-mode login
```

Important:

Use the same queue name everywhere. Right now the function binding expects `ticket-emails`.

### 6. Create Key Vault secrets

Create or load the needed secret values:

```powershell
az keyvault secret set --vault-name "${PREFIX}-kv" --name "postgres-admin-password" --value "<strong-password>"
az keyvault secret set --vault-name "${PREFIX}-kv" --name "app-jwt-secret" --value "<jwt-secret>"
az keyvault secret set --vault-name "${PREFIX}-kv" --name "resend-api-key" --value "<resend-api-key>"
az keyvault secret set --vault-name "${PREFIX}-kv" --name "resend-from" --value "<verified-from-address>"
az keyvault secret set --vault-name "${PREFIX}-kv" --name "storage-connection-string" --value "<storage-connection-string>"
az keyvault secret set --vault-name "${PREFIX}-kv" --name "function-database-url" --value "<postgres-connection-string>"
```

### 7. Prepare the backend container image

This repo does not yet contain a backend `Dockerfile`, so add one before this step.

Recommended image requirements:

- build the Go binary from `./apps/backend/cmd/movie-reservation`
- expose port `8080`
- copy `apps/backend/templates/emails`
- run the compiled binary as the container entrypoint

Build the image:

```powershell
docker build -t "${PREFIX}acr.azurecr.io/movie-reservation-api:latest" .
```

Authenticate to ACR:

```powershell
az acr login --name "${PREFIX}acr"
```

Push the image:

```powershell
docker push "${PREFIX}acr.azurecr.io/movie-reservation-api:latest"
```

### 8. Deploy the backend host

There is no backend host Bicep module yet, so deploy this manually for now.

Recommended target:

- Azure Container Apps

Create a Container App environment:

```powershell
az containerapp env create `
  --name "${PREFIX}-cae" `
  --resource-group $RESOURCE_GROUP `
  --location $LOCATION
```

Create the backend app:

```powershell
az containerapp create `
  --name "${PREFIX}-api" `
  --resource-group $RESOURCE_GROUP `
  --environment "${PREFIX}-cae" `
  --image "${PREFIX}acr.azurecr.io/movie-reservation-api:latest" `
  --target-port 8080 `
  --ingress external `
  --registry-server "${PREFIX}acr.azurecr.io"
```

Then set all required environment variables on the Container App, preferably from Key Vault references or secret values.

### 9. Deploy the Azure Functions host

There is no Function App Bicep module yet, so deploy this manually for now.

Create a Function App plan and app, or use Flex Consumption if preferred by your Azure standards.

Minimum function app settings required:

- `AzureWebJobsStorage`
- `FUNCTIONS_WORKER_RUNTIME`
- `DATABASE_URL`
- `RESEND_API_KEY`
- `RESEND_FROM`

Important repo note:

The function code in [main.go](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/functions/send-ticket-email/main.go) currently hardcodes the sender address instead of reading `RESEND_FROM`. Update that file before production deployment.

Publish the function app from the `functions` folder with Azure Functions Core Tools after the host exists.

### 10. Apply database migrations against Azure PostgreSQL

Point the app or your shell environment to Azure PostgreSQL and run:

```powershell
go run ./apps/backend/cmd/movie-reservation migrate up
```

Make sure these values are set to Azure, not local Docker:

- `CINERESERVE_DATABASE.HOST`
- `CINERESERVE_DATABASE.PORT`
- `CINERESERVE_DATABASE.USER`
- `CINERESERVE_DATABASE.PASSWORD`
- `CINERESERVE_DATABASE.NAME`
- `CINERESERVE_DATABASE.SSL_MODE=require`

### 11. Configure production mode queue behavior

The backend only uses Azure Queue for ticket email dispatch when:

- `CINERESERVE_PRIMARY.ENV=production`
- `CINERESERVE_AZURE.STORAGE_CONNECTION_STRING` is set
- queue client initialization succeeds

If any of those are missing, the app falls back to local-style job handling logic and will not use the Azure Function email path.

### 12. Post-deployment smoke tests

Run these checks in order.

#### API checks

- `GET /status`
- `POST /api/v1/auth/login`
- `GET /api/v1/genres`
- `GET /api/v1/movies`

#### Admin checks

- create a movie
- fetch admin screens
- create a showtime

#### Reservation checks

- reserve seats as a normal user
- confirm reservation row is created in PostgreSQL
- confirm `available_seats` decreases

#### Email checks

- confirm the backend writes a queue message
- confirm the Azure Function consumes the message
- confirm the email arrives in the inbox

#### Expiry checks

- insert or create a pending reservation with an old `expires_at`
- wait for the timer trigger
- confirm the reservation becomes `expired`
- confirm seats are restored to the showtime

## Operational Checks

After deployment, confirm:

- PostgreSQL server is reachable from the backend host
- Key Vault secrets are accessible
- storage queue exists and receives messages
- blob container exists and accepts uploads
- backend app logs show successful startup
- function logs show successful execution
- monitoring workspace is collecting logs

## Rollback Plan

If deployment fails after infrastructure but before application stabilization:

1. Stop external traffic to the backend app.
2. Roll back the backend image to the previous known-good tag.
3. Disable or stop the `send-ticket-email` function if it is misprocessing messages.
4. Restore application settings from the previous release set.
5. If a migration introduced breaking schema changes, use a forward-fix migration rather than editing live schema manually.

Do not delete the PostgreSQL server as a first rollback step unless the environment is disposable.

## Production Readiness Checklist

Before calling this production-ready, complete these repo items:

- add backend `Dockerfile`
- parameterize PostgreSQL admin password in Bicep
- add queue and blob container creation to Bicep
- add backend host module to Bicep
- add Function App module to Bicep
- wire managed identity and Key Vault references
- update `send-ticket-email` function to use `RESEND_FROM`
- verify email template path resolution inside the deployed function package
- add an Azure deployment pipeline
- run end-to-end validation in a real Azure environment

## Repo References

- [main.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/main.bicep)
- [postgres.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/modules/postgres.bicep)
- [storage.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/modules/storage.bicep)
- [container-registry.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/modules/container-registry.bicep)
- [keyvault.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/modules/keyvault.bicep)
- [monitoring.bicep](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure/modules/monitoring.bicep)
- [send-ticket-email function.json](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/functions/send-ticket-email/function.json)
- [send-ticket-email main.go](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/functions/send-ticket-email/main.go)
- [expire-reservations main.go](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/functions/expire-reservations/main.go)
- [Taskfile.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/Taskfile.yml)
