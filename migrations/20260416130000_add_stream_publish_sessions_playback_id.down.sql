DROP INDEX IF EXISTS idx_stream_publish_sessions_playback_id_unique;

ALTER TABLE stream_publish_sessions
  DROP COLUMN IF EXISTS playback_id;

