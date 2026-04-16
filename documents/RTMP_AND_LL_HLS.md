# RTMP and LL-HLS in streaming

This document explains **RTMP** and **LL-HLS**: what the acronyms stand for, their role in the pipeline, and how they behave at a system level. It matches the demo project architecture (OBS/ffmpeg publishing to SRS, viewers watching over HLS).

---

## RTMP

### What it stands for

**RTMP** = **Real-Time Messaging Protocol**.

It was originally tied to **Adobe Flash**; today RTMP is still widely used for **publish/ingest** from an encoder to a media server, not for most modern web players.

### Typical roles

| Role | Short description |
|------|-------------------|
| **Ingest** | Devices or apps (OBS, ffmpeg) **push a continuous stream** to the server over RTMP. |
| **Web playback** | Rarely used end-to-end (browsers no longer run Flash); playback usually **switches to HLS/DASH/WebRTC**. |

### How it works (conceptual)

1. **TCP connection** between the publisher (OBS) and the media server (e.g. SRS) to a **URL like** `rtmp://host:1935/app/stream_key`.
2. **Handshake** per the RTMP specification.
3. Data is framed as RTMP **chunks/messages**: **metadata**, **video codec** (often H.264), **audio codec** (often AAC), etc.
4. The server receives the RTMP stream and may:
   - **Record** it,
   - **Relay** it elsewhere,
   - **Transmux/transcode** to another format — most often **HLS output** (`.m3u8` playlist + segments) for viewers.

### Quick facts

- **Latency**: ingest can be lower than HLS playback, but **end-to-end delay** still depends on transcoding/packaging and the viewer’s buffer.
- **Firewall/NAT**: RTMP uses a fixed TCP port (often **1935**); HTTP (HLS) sometimes passes corporate networks more easily.

---

## LL-HLS

### What it stands for

**LL-HLS** = **Low-Latency HTTP Live Streaming**.

**Baseline HLS** = **HTTP Live Streaming** (delivery over HTTP: `.m3u8` playlist + segments).

LL-HLS is an **implementation profile** of HLS aimed at **reducing latency** versus traditional HLS, still over **HTTP** and friendly to CDNs.

### Why classic HLS is often high-latency

- Each **segment** may last **several seconds** (e.g. 6–10s).
- Players typically **buffer** multiple segments → total delay can reach **tens of seconds** behind “true” live.

### How LL-HLS works (main ideas)

1. **Shorter segments / partial segments (CMAF chunks)**  
   The origin can **publish part of** a segment on the playlist sooner; the player **fetches and plays incrementally** instead of waiting for one long complete segment file.

2. **Faster-updating playlist**  
   The `.m3u8` is refreshed often; modern HLS adds mechanisms such as **blocking playlist reload** and **preload hints** so the CDN/player knows which chunk is coming next.

3. **Still HTTP**  
   Like HLS: easy to put behind a **CDN**, cache, and align with web infrastructure.

4. **Typical outcome**  
   Delay is often **a few seconds** (depends on encoder, CDN, player, buffer settings) — **lower than classic HLS**, but usually **not as low as WebRTC** for ultra-low-latency interaction.

### Requirements

- The **origin** (SRS or encoder/packager) must **emit valid LL-HLS** (partial segments, correct playlist tags).
- The **player** must **support LL-HLS** (Apple’s stacks are strong on HLS; check each JS player’s docs).

---

## RTMP vs LL-HLS (quick comparison)

| Criterion | RTMP | LL-HLS |
|-----------|------|--------|
| **Full name** | Real-Time Messaging Protocol | Low-Latency HTTP Live Streaming |
| **Protocol / transport** | TCP, dedicated RTMP stream | HTTP(S), `.m3u8` + segments |
| **Typical use** | **Ingest** from OBS/ffmpeg | **Playback** via web/CDN |
| **CDN** | Not the classic “HTTP file” model | Fits HTTP caching very well |
| **Viewer latency** | Not directly applicable (viewers rarely use RTMP) | Lower than ordinary HLS, still > 0 |

In this project’s demo pipeline: **RTMP into SRS** → **HLS (can be tuned toward LL-HLS)** → **Nginx CDN** → **player**.

---

## More in this repo

- [PROJECT_OVERVIEW.md](./PROJECT_OVERVIEW.md) — SRS, HLS, and CDN flow.
- [SETUP_AND_DEMO.md](./SETUP_AND_DEMO.md) — run locally and try ingest/playback.
