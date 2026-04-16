# Livestream Readiness Checklist

Checklist to validate end-to-end streaming works locally (OBS/ffmpeg → SRS RTMP ingest → HLS playback → backend state).

## 1) Network / RTMP reachability

- [x] `SRS_RTMP_BASE_URL` uses a host reachable by the streamer client (avoid `localhost` unless client runs on same machine as SRS)
- [x] Docker/compose exposes RTMP port `1935` (e.g. `1935:1935`)
- [x] Firewall / security group allows inbound TCP `1935` from streamer networks (LAN/public as needed)
- [x] Verify publish from the streamer machine:
  - [x] OBS: set Server = `SRS_RTMP_BASE_URL`, Stream Key = `<stream_key>`
- [x] ffmpeg: publish to `rtmp://host:1935/live/<stream_key>`
- [x] Confirm SRS logs show a publish connection for `<stream_key>`

## 2) SRS webhook / reliable live state

- [ ] Configure SRS HTTP callback `on_publish` → backend endpoint
- [ ] Implement backend endpoint:
  - [ ] `POST /api/v1/media/webhooks/srs/on-publish`
  - [ ] Validate shared secret/signature to prevent spoofing
  - [ ] Map SRS `stream` (stream key credential) → stream key → publish session (created)
  - [ ] Update `stream_publish_sessions`:
    - [ ] `status = live`
    - [ ] `started_at = now()`
  - [ ] Upsert `media_ingest_bindings` for traceability (publish_session_id + provider_publish_id + param)
- [ ] (Optional) Add `on_unpublish` to set `status = ended` and `ended_at`
- [ ] Add unit tests for webhook handler/service (validation + mapping + state transition)

## 3) Playback correctness (HLS)

- [x] Confirm the actual HLS path output by SRS / Nginx (playlist location + naming)
- [x] Set `SRS_PLAYBACK_BASE_URL` to the correct base URL for HLS playback
- [x] Validate playback URL matches the output convention:
  - [x] `GET {SRS_PLAYBACK_BASE_URL}/{playback_id}.m3u8` returns `200`
  - [x] segments referenced by the playlist are reachable (`200`)
- [ ] If using CDN later: confirm CDN origin + cache behavior matches expected playback URL

## 4) API contracts (streamer UX)

- [ ] `POST /api/v1/stream-publish-sessions` returns:
  - [ ] `playback_url_cdn`
- [ ] Ensure `playback_url_cdn` is derived from environment/provider config (no hardcoded localhost in services)
- [ ] Ensure `/api/v1/stream-publish-sessions/:id/start` is optional/manual (prefer webhook-driven state in real flow)

