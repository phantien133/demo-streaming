# Streaming Learn

Go + Gin baseline for a learning-focused streaming system.

## Prerequisites
- Go 1.25+
- Docker + Docker Compose

## Run API locally
```bash
go run ./cmd/api
```

Health endpoints:
- `GET /healthz`
- `GET /api/v1/health`

## Run local stack with Docker
```bash
docker compose up --build
```

Services:
- API: `http://localhost:8080`
- Postgres: `localhost:5432`
- Redis: `localhost:6379`
- SRS HTTP API: `http://localhost:1985`
- SRS HLS port mapped to `http://localhost:8081`
