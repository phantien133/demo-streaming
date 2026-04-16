# TODO API Resource Checklist

Checklist API resources for the streaming demo project.

## 1) System
- [x] `GET /healthz`
- [x] `GET /api/v1/health`
- [ ] `GET /api/v1/version`

## 2) Auth (MVP can be stubbed)
- [ ] `POST /api/v1/auth/login`
- [ ] `POST /api/v1/auth/refresh`
- [ ] `GET /api/v1/auth/me`

## 3) Users / Streamers
- [ ] `GET /api/v1/users/:id`
- [ ] `GET /api/v1/streamers/:id`
- [ ] `GET /api/v1/streamers/:id/live-status`

## 4) Stream Sessions (MVP core)
- [ ] `POST /api/v1/streams`
- [ ] `GET /api/v1/streams/:streamId`
- [ ] `PATCH /api/v1/streams/:streamId`
- [ ] `POST /api/v1/streams/:streamId/start`
- [ ] `POST /api/v1/streams/:streamId/end`
- [ ] `GET /api/v1/streams/:streamId/playback`
- [ ] `GET /api/v1/streams?status=live|ended&cursor=...`

## 5) Media Webhooks (SRS -> backend)
- [ ] `POST /api/v1/media/webhooks/srs/on-publish`
- [ ] `POST /api/v1/media/webhooks/srs/on-unpublish`
- [ ] `POST /api/v1/media/webhooks/srs/on-play`
- [ ] `POST /api/v1/media/webhooks/srs/on-record`

## 6) Chat / Realtime
- [ ] `GET /api/v1/streams/:streamId/chat/history`
- [ ] `POST /api/v1/streams/:streamId/chat/messages`
- [ ] `WS /ws/streams/:streamId/chat`

## 7) Gifts / Wallet
- [ ] `GET /api/v1/wallets/me/balance`
- [ ] `POST /api/v1/streams/:streamId/gifts`
- [ ] `GET /api/v1/streams/:streamId/gifts`
- [ ] `GET /api/v1/transactions?cursor=...`

## 8) Stats / Leaderboard
- [ ] `GET /api/v1/streams/:streamId/stats`
- [ ] `GET /api/v1/streams/:streamId/leaderboard`
- [ ] `GET /api/v1/streamers/:id/analytics`

## 9) Admin / Moderation
- [ ] `POST /api/v1/admin/streams/:streamId/terminate`
- [ ] `POST /api/v1/admin/streams/:streamId/mute-user`
- [ ] `GET /api/v1/admin/reports`

## Notes
- [ ] Add request/response schema for each endpoint in `documents/API_RESOURCES.md`.
- [ ] Define standard error format for all REST APIs.
- [ ] Define websocket message schema: `type`, `trace_id`, `payload`, `ts`.
