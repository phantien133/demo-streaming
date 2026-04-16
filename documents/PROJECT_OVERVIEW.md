# Project Overview - Streaming Learn

## Goal
Build a learning-oriented live streaming platform using Go backend services, SRS media server, and a local CDN layer.

## Core Components
- **Web client (demo first):** stream viewer/player and simple chat UI.
- **API backend (Go + Gin):** health, stream session metadata, chat/event APIs.
- **Media origin (SRS):** RTMP ingest and HLS output.
- **CDN edge (Nginx):** cache HLS playlist/segments to reduce origin load.
- **Redis:** cache stream state, counters, and pub/sub for realtime fan-out.
- **PostgreSQL:** durable storage for users, stream sessions, gifts, and history.

## High-Level Flow
1. Streamer publishes to SRS over RTMP (`/live/<stream_key>`).
2. SRS generates HLS output (`.m3u8`, `.ts`/`.m4s`).
3. Viewer requests playback from Nginx CDN edge.
4. CDN serves cached segments or fetches from SRS origin on cache miss.
5. Backend handles non-media data: stream metadata, chat, gifts, leaderboard.

## Why Local CDN Layer
- Demonstrate realistic architecture without external cloud CDN.
- Validate caching strategy before deploying cloud edge.
- Keep media path independent from API/backend path.

## Proposed Backend Modules
- `cmd/api`: bootstrap HTTP server.
- `internal/server`: routing + middleware.
- `internal/stream`: stream session lifecycle.
- `internal/chat`: websocket room and message fan-out.
- `internal/media`: media provider adapter (SRS now, replaceable later).
- `internal/storage`: postgres + redis repositories.

## Phased Roadmap
- **Phase 1 (current):** base API + docker stack + local CDN edge.
- **Phase 2:** stream session APIs + SRS callback endpoint + Redis cache.
- **Phase 3:** web demo (viewer page with `hls.js`, simple chat).
- **Phase 4:** auth stub, gifts/wallet mock, observability enhancements.

## Non-Goals for MVP
- Production-grade multi-region CDN orchestration.
- DRM, anti-piracy, and advanced video moderation.
- Full billing system.
