# Movie Reservation Service

`movie-reservation` is a Go backend for browsing movies and showtimes, reserving seats, handling admin workflows, and sending ticket confirmations.

Project brief:

- https://roadmap.sh/projects/movie-reservation-system

The project is built around:

- Echo for the HTTP API
- PostgreSQL for core data
- Redis + Asynq for local background jobs
- Azure Queue + Azure Functions for production-oriented async processing
- Resend for transactional email
- Azure Blob Storage for poster uploads
- Bicep for infrastructure
- GitHub Actions for CI/CD
- GitHub Container Registry for the API image

## Current Status

Working in the repo today:

- local API startup with config validation
- PostgreSQL schema and seed migrations
- local auth with register/login endpoints
- public movies, genres, showtimes, and seat lookup endpoints
- authenticated reservation create/list/get/cancel flow
- admin movie, showtime, screen, reporting, and promote-user endpoints
- reservation locking with `SELECT ... FOR UPDATE`
- local ticket email dispatch with Redis + Asynq
- Azure Functions custom-handler packaging scaffold
- Azure Bicep modules for PostgreSQL, storage, Key Vault, monitoring, and optional ACR
- GitHub Actions workflows for infra, API, and Functions deployment

Still incomplete:

- live Clerk integration
- Bicep modules for the Container Apps host and Function App host
- full Azure runtime secret wiring through Key Vault references
- end-to-end Azure validation in a real subscription
- broader integration and concurrency test coverage

## Repository Layout

```text
movie-reservation/
|-- apps/
|   `-- backend/
|       |-- cmd/movie-reservation/
|       |-- internal/
|       `-- templates/emails/
|-- functions/
|   |-- cmd/handler/
|   |-- expire-reservations/
|   `-- send-ticket-email/
|-- azure/
|   |-- main.bicep
|   `-- modules/
|-- .github/workflows/
|-- docs/
|-- docker-compose.yml
|-- Dockerfile
`-- Taskfile.yml
```

## Local Development

### Prerequisites

- Go 1.25+
- Docker Desktop
- `task` CLI

### Environment

Local development uses [.env](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.env). The sample version is [.env.sample](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.env.sample).

Most important values:

- `CINERESERVE_AUTH.SECRET_KEY`
- `CINERESERVE_INTEGRATION.RESEND_API_KEY`
- `CINERESERVE_INTEGRATION.RESEND_FROM`
- `CINERESERVE_APP.BASE_URL`
- `CINERESERVE_AZURE.STORAGE_CONNECTION_STRING`
- `DATABASE_URL`
- `RESEND_API_KEY`
- `AzureWebJobsStorage`

### Start dependencies

```bash
task up
```

This starts:

- PostgreSQL on `localhost:5434`
- Redis on `localhost:6379`

### Apply migrations

```bash
task migrate-up
```

### Run the API

```bash
task run
```

The API listens on:

```text
http://localhost:8080
```

Health check:

```text
GET /status
```

## Task Commands

The root [Taskfile.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/Taskfile.yml) provides:

```bash
task run
task build
task test
task fmt
task up
task down
task logs
task migrate-up
task migrate-down
task functions-test
task dev
```

## API Surface

### Public

- `GET /status`
- `GET /api/v1/genres`
- `GET /api/v1/movies`
- `GET /api/v1/movies/:id`
- `GET /api/v1/showtimes`
- `GET /api/v1/showtimes/:id/seats`
- `GET /api/v1/screens/:id/seats`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

### Authenticated

- `POST /api/v1/reservations`
- `GET /api/v1/reservations/me`
- `GET /api/v1/reservations/:id`
- `DELETE /api/v1/reservations/:id`

### Admin

- `POST /api/v1/movies`
- `PUT /api/v1/movies/:id`
- `DELETE /api/v1/movies/:id`
- `POST /api/v1/movies/:id/poster`
- `POST /api/v1/showtimes`
- `GET /api/v1/admin/screens`
- `GET /api/v1/admin/reservations`
- `GET /api/v1/admin/reports/revenue`
- `GET /api/v1/admin/reports/capacity`
- `PUT /api/v1/admin/users/:id/promote`

## Reservation and Email Flow

Current reservation behavior:

1. validate the request and authenticated user
2. lock the showtime row
3. verify the requested seats belong to that screen
4. reject already-reserved seats and return the exact conflicting seat labels
5. create the reservation inside a transaction
6. decrement `available_seats`
7. commit
8. enqueue ticket confirmation work after commit

Email dispatch differs by environment:

- `development`: Redis + Asynq handles the ticket confirmation job
- `production`: the backend can enqueue to Azure Queue, and Azure Functions can process the queue message

The email template is:

- [ticket_confirmation.html](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/apps/backend/templates/emails/ticket_confirmation.html)

## Azure and GitHub Deployment

The repo now supports a GitHub-first deployment direction:

- GitHub Actions deploys base infrastructure
- GitHub Actions builds the backend Docker image
- GitHub Actions pushes the API image to GHCR
- GitHub Actions deploys the API image to Azure Container Apps
- GitHub Actions builds and publishes the Azure Functions package

Current workflow files:

- [deploy-infra.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.github/workflows/deploy-infra.yml)
- [deploy-api.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.github/workflows/deploy-api.yml)
- [deploy-functions.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.github/workflows/deploy-functions.yml)

Important current limitation:

- the Bicep in this repo does not yet create the Container Apps environment, the API Container App, or the Function App host

So the workflows are in place, but those host resources still need to exist before full deployment succeeds.

## Infrastructure

Azure infrastructure definitions live under [azure](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure).

Current modules:

- PostgreSQL Flexible Server
- Storage Account
- Blob container creation
- Queue creation
- Key Vault
- Log Analytics Workspace
- optional Azure Container Registry creation

## Testing

Run tests with:

```bash
task test
```

Current coverage includes:

- config loading and parsing
- health endpoint checks
- job payload tests

Still recommended:

- repository integration tests
- reservation concurrency tests
- handler auth and admin tests
- Azure Functions integration tests
- Azure deployment validation tests

## Documentation

Deployment docs:

- [AZURE_DEPLOYMENT_RUNBOOK.md](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/docs/AZURE_DEPLOYMENT_RUNBOOK.md)
- [GITHUB_AZURE_DEPLOYMENT_SETUP.md](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/docs/GITHUB_AZURE_DEPLOYMENT_SETUP.md)

Testing assets:

- [movie-reservation.postman_collection.json](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/postman/movie-reservation.postman_collection.json)
- [movie-reservation.test-data.json](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/postman/movie-reservation.test-data.json)
- [movie-reservation.runner-data.json](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/postman/movie-reservation.runner-data.json)
