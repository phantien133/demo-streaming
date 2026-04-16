-- Long-lived ingest secret per broadcaster (OBS stream key) before any publish session exists.
CREATE TABLE stream_keys (
  id BIGSERIAL PRIMARY KEY,
  owner_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
  stream_key_secret TEXT NOT NULL UNIQUE,
  media_provider_id BIGINT NOT NULL REFERENCES media_providers (id) ON DELETE RESTRICT,
  label TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_stream_keys_owner_user_id ON stream_keys (owner_user_id);
