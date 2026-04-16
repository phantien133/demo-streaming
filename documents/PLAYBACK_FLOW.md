# Playback stream flow (demo stack)

End-to-end path from publisher to viewer when the **transcode** profile is enabled and the worker has produced an ABR pack.

```
OBS (RTMP push)
        ↓
   [ SRS ingest ]
        ↓
   (internal HLS / RTMP — same stream name as playback_id)
        ↓
   [ Transcoder (FFmpeg) ]
   on_publish → Redis job → worker pulls SRS (RTMP or HLS)
        ↓
   multi-bitrate HLS on disk
   ./tmp/transcode/<playback_id>/
     master.m3u8
     720p/index.m3u8, …
     480p/index.m3u8, …
        ↓
   Nginx CDN (8088) serves files under
   /live/<stream_id>/…
        ↓
   Player opens master playlist:
   http://<edge>:8088/live/<stream_id>/master.m3u8
        ↓
       Player
```

## URLs

| Stage | URL pattern |
|--------|----------------|
| **Viewer (ABR, recommended)** | `{CDN_BASE}/live/<playback_id>/master.m3u8` — Nginx serves the file from disk when the transcode pack exists; **until then** the same URL falls back to the API (SRS single-bitrate playlist), so the player always gets a valid master. |
| **Viewer (flat, optional)** | `{CDN_BASE}/live/<playback_id>.m3u8` → API proxies SRS (same playlist as pre-transcode fallback above). |

`playback_id` equals the RTMP stream name / SRS path segment for this demo (see stream publish session create).

## Components

- **SRS**: ingest + origin HLS used as **transcoder input** and as **fallback** playback.
- **Transcoder**: writes `master.m3u8` and variant folders under `tmp/transcode/<playback_id>/`.
- **CDN**: mounts `tmp/transcode` read-only at `/var/transcode` and maps `/live/<hex>/…` to those files; hex-only IDs avoid colliding with other `/live/` paths.

See also: [PROJECT_OVERVIEW.md](./PROJECT_OVERVIEW.md), [README.md](../README.md) (Docker / transcode profile).
