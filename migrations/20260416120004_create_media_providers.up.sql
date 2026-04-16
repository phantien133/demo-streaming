CREATE TABLE media_providers (
  id BIGSERIAL PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL DEFAULT '',
  api_base_url TEXT,
  config JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO media_providers (code, display_name)
VALUES ('srs', 'SRS (Simple Realtime Server)');
