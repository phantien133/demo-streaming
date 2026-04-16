# Setup and Demo Runbook

## 1) Prerequisites
- Docker Desktop (or Docker Engine + Compose plugin)
- OBS Studio (for RTMP publishing)

## 2) Start the stack
From repository root:

```bash
docker compose up --build
```

Wait until these services are healthy/ready:
- `streaming_api`
- `streaming_postgres`
- `streaming_redis`
- `streaming_srs`
- `streaming_cdn`

## 3) Publish a live stream from OBS
In OBS settings:
- **Service:** Custom
- **Server:** `rtmp://localhost:1935/live`
- **Stream key:** choose a value, e.g. `demo123`

Start streaming from OBS.

## 4) Playback URLs
- **SRS origin:** `http://localhost:8081/live/demo123.m3u8`
- **CDN edge:** `http://localhost:8088/live/demo123.m3u8`

Use a browser player supporting HLS or a simple page with `hls.js`.

## 5) CDN cache check
Request the CDN URL repeatedly and inspect headers:
- `X-Cache-Status: MISS` on first request
- `X-Cache-Status: HIT` on next requests during TTL window

Caching policy in this demo:
- Playlist (`.m3u8`): `max-age=2`
- Segments (`.ts`, `.m4s`, `.mp4`, `.aac`): `max-age=30`

## 6) API checks
- `GET http://localhost:8080/healthz`
- `GET http://localhost:8080/api/v1/health`

## 7) Troubleshooting quick list
- **Playback 404:** verify stream key and OBS publish status.
- **No cache HIT:** request the same segment URL multiple times.
- **High latency:** reduce HLS segment duration in origin config (future tuning).
- **CORS issues in web client:** CDN already sets permissive CORS headers.
