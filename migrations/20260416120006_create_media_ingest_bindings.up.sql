-- Provider-native ingest identity and webhook-derived fields (SRS today; other adapters later).
CREATE TABLE media_ingest_bindings (
  id BIGSERIAL PRIMARY KEY,
  publish_session_id BIGINT NOT NULL UNIQUE REFERENCES stream_publish_sessions (id) ON DELETE CASCADE,
  provider_publish_id TEXT,
  provider_vhost TEXT,
  provider_app TEXT,
  ingest_query_param TEXT,
  record_local_uri TEXT,
  last_callback_action TEXT,
  last_callback_at TIMESTAMPTZ,
  provider_context JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_ingest_bindings_provider_publish_id ON media_ingest_bindings (provider_publish_id);
