# Demo-streaming

Go + Gin baseline for a learning-focused streaming system.

## Prerequisites
- Go 1.25+
- Docker + Docker Compose

## Environment setup
```bash
cp .env.example .env
```

Both Go commands and `Makefile` read the same variables from `.env`.

## Run API locally
```bash
go run ./cmd/api
```

Health endpoints:
- `GET /healthz`
- `GET /api/v1/health`

## Swagger UI
Generate docs (one-time or after changing annotations):

```bash
make swagger-install
make swagger-gen
```

Open Swagger UI:
- `http://localhost:8080/swagger/index.html`

If you want demo data in Swagger, prepare local DB:

```bash
make swagger-db
```

## JWT setup (demo)
- Set app auth config in `.env`:
  - `JWT_SECRET`
  - optional `JWT_ISSUER`
  - optional `JWT_ACCESS_TOKEN_TTL_SECONDS` (default 3600)
  - optional `JWT_REFRESH_TOKEN_TTL_SECONDS` (default 604800, reserved for refresh-token flow)
- Issue a token:
  - `POST /api/v1/auth/token`
  - body example: `{"email":"streamer01@example.com","password":"password123"}`
  - role is assigned by server (`end_user`), TTL is loaded from env.
- Refresh token (rotation):
  - `POST /api/v1/auth/refresh`
  - body example: `{"refresh_token":"<refresh_token>"}`
- Revoke refresh token immediately:
  - `POST /api/v1/auth/revoke`
  - body example: `{"refresh_token":"<refresh_token>"}`
  - token state is stored in Redis, so revoked refresh tokens are rejected right away.
- Access protected endpoint:
  - `GET /api/v1/auth/me`
  - header: `Authorization: Bearer <access_token>`

## Run full local stack with Docker
```bash
docker compose up --build
```

Services:
- API: `http://localhost:8080`
- Postgres: `localhost:5432`
- Redis: `localhost:6379`
- SRS HTTP API: `http://localhost:1985`
- SRS HLS origin: `http://localhost:8081`
- Nginx local CDN edge: `http://localhost:8088`

Optional profile:
- Transcode worker: `docker compose --profile transcode up --build transcode`

## Streaming flow (demo)
See `documents/PLAYBACK_FLOW.md` for the full diagram. Summary:

- Publish from OBS via RTMP: `rtmp://<SRS_HOST>:1935/live/<stream_key>` (stream name = `playback_id`).
- SRS ingests → **transcoder** (optional profile) produces multi-bitrate HLS under `./tmp/transcode/<playback_id>/`.
- **Viewer (ABR master):** `http://<CDN_HOST>:8088/live/<playback_id>/master.m3u8` (served from disk when the pack exists).
- **SRS origin (single-bitrate):** `http://localhost:8081/live/<playback_id>.m3u8`
- **CDN flat playlist (API → SRS proxy):** `http://<CDN_HOST>:8088/live/<playback_id>.m3u8` — useful before transcode finishes or without the transcode profile.

Dev tip (LAN IP quick setup):
```bash
make local-ip
```
Set `.env`:
- `DEV_SRS_RTMP_PUBLISH_BASE_URL=rtmp://<LAN-IP>:1935/live`
- `DEV_SRS_PLAYBACK_CDN_BASE_URL=http://<LAN-IP>:8088/live`

## Verify CDN cache behavior
1. Start stream from OBS.
2. Open CDN URL in a player/browser (or with `hls.js` client).
3. Inspect response headers:
   - `X-Cache-Status: MISS` on first request
   - `X-Cache-Status: HIT` on repeated requests

## Transcode module (MVP)
- SRS `on_publish` webhook enqueues a transcode job into Redis list `transcode:publish_jobs`.
- Worker entrypoint: `go run ./cmd/transcode` (or `make transcode-run`).
- Enable ffmpeg processing by setting `TRANSCODE_ENABLED=true`.
- Output defaults to `./tmp/transcode/<playback_id>/master.m3u8`.

## Documentation
- `documents/PLAYBACK_FLOW.md`: OBS → SRS → transcode → multi-bitrate HLS → player.
- `documents/PROJECT_OVERVIEW.md`: architecture, components, and roadmap.
- `documents/SETUP_AND_DEMO.md`: local setup and step-by-step demo runbook.
- `documents/RTMP_AND_LL_HLS.md`: RTMP ingest and LL-HLS playback (see also `documents/README.md`).

## Database migration (golang-migrate)
This project uses `golang-migrate` via Go command at `cmd/migrate`.

### Migration commands
```bash
make migrate-version
make migrate-status
make migrate-create name=create_users_table
make migrate-up
make migrate-down
```

### Notes
- Default database URL: `postgres://streaming:streaming@localhost:5433/streaming?sslmode=disable`
- Override any variable at runtime, for example:
  - `make migrate-up DB_HOST=127.0.0.1`
  - `make migrate-up DB_PORT=5432 DB_NAME=streaming`
- Migrations are stored in `migrations/`.
