CREATE TABLE stream_publish_sessions (
  id BIGSERIAL PRIMARY KEY,
  streamer_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
  stream_key TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'created'
    CHECK (status IN ('created', 'live', 'ended', 'failed')),
  playback_url_cdn TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  started_at TIMESTAMPTZ,
  ended_at TIMESTAMPTZ
);

CREATE INDEX idx_stream_publish_sessions_streamer_user_id ON stream_publish_sessions (streamer_user_id);
CREATE INDEX idx_stream_publish_sessions_status ON stream_publish_sessions (status);
CREATE INDEX idx_stream_publish_sessions_created_at ON stream_publish_sessions (created_at DESC);
