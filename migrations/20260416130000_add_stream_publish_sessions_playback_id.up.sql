ALTER TABLE stream_publish_sessions
  ADD COLUMN playback_id TEXT;

-- Backfill existing rows (no pgcrypto dependency).
UPDATE stream_publish_sessions
SET playback_id = substring(md5(id::text || clock_timestamp()::text || random()::text) for 24)
WHERE playback_id IS NULL OR length(trim(playback_id)) = 0;

ALTER TABLE stream_publish_sessions
  ALTER COLUMN playback_id SET NOT NULL;

CREATE UNIQUE INDEX idx_stream_publish_sessions_playback_id_unique
  ON stream_publish_sessions (playback_id);

