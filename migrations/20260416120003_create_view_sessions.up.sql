CREATE TABLE view_sessions (
  id BIGSERIAL PRIMARY KEY,
  publish_session_id BIGINT NOT NULL REFERENCES stream_publish_sessions (id) ON DELETE CASCADE,
  viewer_user_id BIGINT REFERENCES users (id) ON DELETE SET NULL,
  viewer_ref TEXT NOT NULL DEFAULT '',
  client_type TEXT NOT NULL DEFAULT 'web',
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  left_at TIMESTAMPTZ,
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT view_sessions_viewer_identity_chk CHECK (
    viewer_user_id IS NOT NULL OR length(trim(viewer_ref)) > 0
  )
);

CREATE INDEX idx_view_sessions_publish_session_id ON view_sessions (publish_session_id);
CREATE INDEX idx_view_sessions_viewer_user_id ON view_sessions (viewer_user_id);
CREATE INDEX idx_view_sessions_joined_at ON view_sessions (joined_at DESC);
