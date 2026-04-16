# Documents index

- `system flow.jpg`: initial architecture sketch from project kickoff.
- `system-flow-with-cdn.puml`: PlantUML diagram (SRS, CDN, gateway, backend).
- `livestream-session-sequence.puml`: sequence diagram — stream key (early or at go-live), RTMP, SRS webhook, viewer join, HLS via CDN.
- `database-design.puml`: ERD — `stream_publish_sessions` vs `view_sessions`.
- `PLAYBACK_FLOW.md`: playback path OBS → SRS → FFmpeg transcode → `/live/<id>/master.m3u8` → player.
- `PROJECT_OVERVIEW.md`: project scope, architecture, modules, and roadmap.
- `SETUP_AND_DEMO.md`: local setup and end-to-end demo instructions.
- `RTMP_AND_LL_HLS.md`: RTMP ingest and LL-HLS playback concepts for this demo stack.
