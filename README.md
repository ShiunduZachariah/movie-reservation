# Movie Reservation Service

A Go-based movie reservation backend for browsing movies and showtimes, reserving seats, sending ticket confirmations, and preparing the service for Azure deployment.

The project is built around a layered architecture:

- Echo for the HTTP API
- PostgreSQL for core data
- Redis + Asynq for local background jobs
- Azure Functions for timer and queue-driven background processing
- Resend for transactional email
- Azure Blob Storage for poster uploads
- Bicep for infrastructure scaffolding

## Current Status

Implemented in the repo today:

- Project scaffold with backend, functions, and Azure folders
- Environment-based config loading and validation
- PostgreSQL schema and seed migrations
- Movie, showtime, reservation, and admin reporting service layers
- Seat reservation flow with `SELECT ... FOR UPDATE`
- Ticket confirmation email template and Asynq task
- Azure Function stubs for reservation expiry and queue-based email sending
- Azure Blob client scaffold for poster uploads
- Basic tests for config, health, and job payloads

Not fully complete yet:

- Live Clerk authentication integration
- End-to-end Azure deployment wiring
- Full repository/service concurrency and integration test coverage
- Production-ready Azure Function runtime integration
- Complete infrastructure modules for Container Apps and Function Apps

## Repository Structure

```text
movie-reservation/
├── apps/
│   └── backend/
│       ├── cmd/movie-reservation/
│       ├── internal/
│       └── templates/emails/
├── functions/
│   ├── expire-reservations/
│   └── send-ticket-email/
├── azure/
│   ├── main.bicep
│   └── modules/
├── docker-compose.yml
├── Taskfile.yml
├── .env
└── .env.sample
```

## Tech Stack

- Go 1.25
- Echo
- PostgreSQL
- Redis
- Asynq
- Resend
- Azure Storage Queue
- Azure Blob Storage
- Azure Functions
- Bicep

## Quick Start

### 1. Prerequisites

Install:

- Go 1.25+
- Docker Desktop
- `task` CLI

### 2. Configure environment

The repo already includes a local [.env](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/.env) for development. Before real use, replace the placeholder values for secrets and external services.

The most important variables are:

- `CINERESERVE_AUTH.SECRET_KEY`
- `CINERESERVE_INTEGRATION.RESEND_API_KEY`
- `CINERESERVE_APP.BASE_URL`
- `CINERESERVE_AZURE.STORAGE_CONNECTION_STRING`

For Azure Functions you will also need:

- `DATABASE_URL`
- `RESEND_API_KEY`
- `AzureWebJobsStorage`

### 3. Start local dependencies

```bash
task up
```

This starts:

- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

### 4. Run migrations

```bash
task migrate-up
```

### 5. Start the API

```bash
task run
```

The service starts on:

```text
http://localhost:8080
```

Health check:

```text
GET /status
```

## Task Commands

The root [Taskfile.yml](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/Taskfile.yml) exposes simple commands:

```bash
task run            # start the backend
task build          # build everything
task test           # run Go tests
task fmt            # format Go files
task up             # start postgres + redis
task down           # stop local services
task logs           # tail docker logs
task migrate-up     # apply database migrations
task migrate-down   # rollback migrations
task functions-test # run function package tests
task dev            # start deps and run migrations
```

## Environment Variables

### Core application

- `CINERESERVE_PRIMARY.ENV`
- `CINERESERVE_SERVER.PORT`
- `CINERESERVE_SERVER.READ_TIMEOUT`
- `CINERESERVE_SERVER.WRITE_TIMEOUT`
- `CINERESERVE_SERVER.IDLE_TIMEOUT`
- `CINERESERVE_SERVER.CORS_ALLOWED_ORIGINS`

### Database

- `CINERESERVE_DATABASE.HOST`
- `CINERESERVE_DATABASE.PORT`
- `CINERESERVE_DATABASE.USER`
- `CINERESERVE_DATABASE.PASSWORD`
- `CINERESERVE_DATABASE.NAME`
- `CINERESERVE_DATABASE.SSL_MODE`
- `CINERESERVE_DATABASE.MAX_OPEN_CONNS`
- `CINERESERVE_DATABASE.MAX_IDLE_CONNS`
- `CINERESERVE_DATABASE.CONN_MAX_LIFETIME`
- `CINERESERVE_DATABASE.CONN_MAX_IDLE_TIME`

### Redis

- `CINERESERVE_REDIS.ADDRESS`

### Auth

- `CINERESERVE_AUTH.SECRET_KEY`

Note:

The current implementation uses shared-secret JWT validation for development. The original product plan expects Clerk-based auth, which is still pending.

### Email

- `CINERESERVE_INTEGRATION.RESEND_API_KEY`

### App

- `CINERESERVE_APP.BASE_URL`
- `CINERESERVE_APP.NAME`

### Azure Storage

- `CINERESERVE_AZURE.STORAGE_ACCOUNT_NAME`
- `CINERESERVE_AZURE.STORAGE_CONTAINER_NAME`
- `CINERESERVE_AZURE.STORAGE_QUEUE_NAME`
- `CINERESERVE_AZURE.STORAGE_CONNECTION_STRING`

## Database and Migrations

Migrations live in [apps/backend/internal/database/migrations](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/apps/backend/internal/database/migrations).

Current migration set:

- `001_setup.sql` creates the core schema
- `002_seed_admin.sql` seeds an admin placeholder user
- `003_seed_genres.sql` seeds movie genres
- `004_seed_screens_and_seats.sql` seeds default screens and seats

The schema includes:

- users
- genres
- movies
- movie_genres
- screens
- seats
- showtimes
- reservations
- reservation_seats

## API Overview

### Public routes

- `GET /status`
- `GET /api/v1/movies`
- `GET /api/v1/movies/:id`
- `GET /api/v1/showtimes`
- `GET /api/v1/showtimes/:id/seats`

### Authenticated routes

- `POST /api/v1/reservations`
- `GET /api/v1/reservations/me`
- `GET /api/v1/reservations/:id`
- `DELETE /api/v1/reservations/:id`

### Admin routes

- `POST /api/v1/movies`
- `PUT /api/v1/movies/:id`
- `DELETE /api/v1/movies/:id`
- `POST /api/v1/movies/:id/poster`
- `POST /api/v1/showtimes`
- `GET /api/v1/admin/reservations`
- `GET /api/v1/admin/reports/revenue`
- `GET /api/v1/admin/reports/capacity`
- `PUT /api/v1/admin/users/:id/promote`

## Reservation Flow

The current reservation flow is:

1. Validate the authenticated user and input.
2. Lock the target showtime row with `SELECT ... FOR UPDATE`.
3. Validate the seats belong to the showtime screen.
4. Check that requested seats are not already reserved.
5. Create the reservation and reservation-seat links in one transaction.
6. Decrement `available_seats`.
7. Commit the transaction.
8. Enqueue a ticket confirmation email after commit.

This keeps seat availability consistent under concurrent requests and avoids sending emails for rolled-back bookings.

## Background Jobs

### Local development

Local email dispatch uses Asynq with Redis:

- Task type: `reservation:ticket_confirmation`

### Azure-oriented flow

The repo also includes function scaffolds for:

- `functions/expire-reservations`
- `functions/send-ticket-email`

These are intended for:

- expiring stale pending reservations
- consuming queue messages and sending ticket emails

## Email Templates

Ticket confirmation email template:

- [ticket_confirmation.html](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/apps/backend/templates/emails/ticket_confirmation.html)

It includes:

- reservation ID
- movie title
- show date and time
- screen name
- seat labels
- total price
- reservation deep link

## Azure Infrastructure

Bicep files live under [azure](c:/Users/ZacH/Documents/Personal-Projects/movie-reservation/azure).

Current modules:

- PostgreSQL
- Container Registry
- Storage
- Monitoring
- Key Vault

Still to be completed:

- Container App module
- Function App module
- managed identity wiring
- Key Vault secret references
- deployment pipeline and runtime configuration

## Testing

Run all current tests:

```bash
task test
```

What is currently covered:

- config loading and parsing
- health endpoint response
- email task payload creation

What still needs to be added:

- repository integration tests
- reservation concurrency tests
- handler auth/admin tests
- function integration tests
- blob upload tests

## Development Notes

- The repo root `.env` is ignored by git.
- Go build/test caches are expected to stay local.
- The service currently builds successfully with `go build ./...`.
- If your editor shows stale dependency or import errors, reload the Go language server.

## Roadmap

High-priority next steps:

- replace development JWT auth with Clerk integration
- add robust repository and service integration tests
- finish Azure Container App and Function App Bicep modules
- wire Azure queue client fully into production mode
- validate the Azure Functions flow locally with Functions Core Tools
- harden reservation lifecycle rules for pending, confirmed, cancelled, and expired states

## License

No license has been added yet. If this project is going to be public, add one before publishing.
