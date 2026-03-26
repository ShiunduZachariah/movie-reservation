FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY apps ./apps

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/movie-reservation ./apps/backend/cmd/movie-reservation

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/movie-reservation ./movie-reservation
COPY apps/backend/templates ./apps/backend/templates

EXPOSE 8080

ENTRYPOINT ["./movie-reservation"]
