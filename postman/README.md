# Postman (demo-streaming)

## Import
- Import collection: `postman/demo-streaming.postman_collection.json`
- Import environment: `postman/demo-streaming.local.postman_environment.json`

## Demo data (from seed)
- Email: `streamer01@example.com`
- Password: `password123`

If you haven't seeded yet:

```bash
make seed
```

## Recommended run order
1) **System** → `GET /healthz` (optional)
2) **Auth** → `POST /api/v1/auth/token (login)` (auto-saves `accessToken`, `refreshToken`)
3) **Auth** → `GET /api/v1/auth/me`
4) **Stream Keys** → `POST /api/v1/stream-keys` (auto-saves `streamKey`)
5) **Stream Publish Sessions** → `POST /api/v1/stream-publish-sessions (create)` (auto-saves `publishSessionId`, `playbackUrl`)
6) **Media Webhooks** → `POST /api/v1/media/webhooks/srs/on-publish (manual)` (optional manual trigger for webhook/transcode queue)
7) **Stream Publish Sessions** → `POST /api/v1/stream-publish-sessions/:id/start`
8) **Stream Publish Sessions** → `POST /api/v1/stream-publish-sessions/:id/stop`

Notes:
- Authorization header uses `Bearer {{accessToken}}` stored as **collection variables**.
- You can change user credentials in the environment (`email`, `password`).
- Stream key is managed separately via **Stream Keys**.
- Viewer ABR playback uses `playbackUrl` → `{CDN}/live/<playback_id>/master.m3u8` (after transcode); flat `/live/<id>.m3u8` remains available via CDN as SRS fallback.
- `playbackId` is auto-extracted from `playbackUrl` and saved to collection variables.

