ALTER TABLE stream_publish_sessions
  ADD COLUMN media_provider_id BIGINT REFERENCES media_providers (id) ON DELETE RESTRICT;

UPDATE stream_publish_sessions
SET media_provider_id = (SELECT id FROM media_providers WHERE code = 'srs' LIMIT 1);

ALTER TABLE stream_publish_sessions
  ALTER COLUMN media_provider_id SET NOT NULL;

CREATE INDEX idx_stream_publish_sessions_media_provider_id ON stream_publish_sessions (media_provider_id);
