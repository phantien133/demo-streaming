CREATE TABLE transcode_renditions (
    id BIGSERIAL PRIMARY KEY,
    publish_session_id BIGINT NOT NULL REFERENCES stream_publish_sessions(id) ON DELETE CASCADE,
    playback_id TEXT NOT NULL,
    rendition_name TEXT NOT NULL,
    playlist_path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('processing', 'ready', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (publish_session_id, rendition_name)
);

CREATE INDEX idx_transcode_renditions_publish_session_id ON transcode_renditions(publish_session_id);
CREATE INDEX idx_transcode_renditions_playback_id ON transcode_renditions(playback_id);
