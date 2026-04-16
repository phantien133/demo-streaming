DROP INDEX IF EXISTS idx_stream_publish_sessions_media_provider_id;

ALTER TABLE stream_publish_sessions
  DROP COLUMN IF EXISTS media_provider_id;
