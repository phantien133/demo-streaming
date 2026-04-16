DROP INDEX IF EXISTS idx_stream_publish_sessions_stream_key_id;

ALTER TABLE stream_publish_sessions
  ADD COLUMN stream_key TEXT;

UPDATE stream_publish_sessions s
SET stream_key = k.stream_key_secret
FROM stream_keys k
WHERE k.id = s.stream_key_id;

ALTER TABLE stream_publish_sessions
  ALTER COLUMN stream_key SET NOT NULL;

CREATE UNIQUE INDEX stream_publish_sessions_stream_key_key ON stream_publish_sessions (stream_key);

ALTER TABLE stream_publish_sessions
  DROP COLUMN stream_key_id;

-- stream_keys (from 20260416120007) may still hold rows; drop or prune manually if you fully rollback both migrations.
