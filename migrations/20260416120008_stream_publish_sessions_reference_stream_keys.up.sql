ALTER TABLE stream_publish_sessions
  ADD COLUMN stream_key_id BIGINT REFERENCES stream_keys (id) ON DELETE RESTRICT;

INSERT INTO stream_keys (owner_user_id, stream_key_secret, media_provider_id)
SELECT DISTINCT ON (stream_key)
  streamer_user_id,
  stream_key,
  media_provider_id
FROM stream_publish_sessions
ORDER BY stream_key, id DESC;

UPDATE stream_publish_sessions s
SET stream_key_id = k.id
FROM stream_keys k
WHERE k.stream_key_secret = s.stream_key;

ALTER TABLE stream_publish_sessions
  ALTER COLUMN stream_key_id SET NOT NULL;

ALTER TABLE stream_publish_sessions
  DROP COLUMN stream_key;

CREATE INDEX idx_stream_publish_sessions_stream_key_id ON stream_publish_sessions (stream_key_id);
