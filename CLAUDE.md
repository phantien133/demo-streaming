# demo-streaming - Agent Guide

## Project Goal
Build a learning-focused live streaming backend in Go with a web client and dockerized local stack.

## System Components
- Web client for viewer and streamer demo flows.
- API gateway for routing, auth middleware, and rate limiting.
- Backend services for stream sessions, chat events, gifts, and wallet stubs.
- Media server (SRS) for RTMP ingest and HLS/LL-HLS playback.
- Transcode worker for multi-quality outputs (480p/720p/1080p).
- Data layer: PostgreSQL for durable state, Redis for cache and distributed locks.

## Preferred Tech Stack
- Language: Go
- HTTP framework: Gin (fast setup, strong ecosystem)
- Realtime: Gorilla WebSocket
- Database access: sqlc + pgx
- Async jobs/events: NATS (simple and lightweight for learning)
- Observability: zap logger + Prometheus metrics

## Repository Layout (target)
- `cmd/api-gateway`
- `cmd/backend`
- `internal/<module>`
- `pkg/shared`
- `deployments/docker`
- `web`

## Coding Expectations
- Keep handlers thin and move logic to services.
- Validate all input at API boundary.
- Keep domain logic testable without network dependencies.
- Add tests for core use cases before expanding features.

## Docker and Local Run Expectations
- Use `docker-compose.yml` as the entrypoint for local stack.
- Start with minimal services: gateway, backend, postgres, redis, srs.
- Add transcoder and optional modules incrementally.
- Document startup commands and sample URLs in README.

## Immediate Milestone
Step 1: baseline project setup and developer docs.
Step 2: implement stream session create/join and chat over WebSocket.
Step 3: integrate SRS callback flow and playback endpoint.
