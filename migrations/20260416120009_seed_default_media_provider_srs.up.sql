INSERT INTO media_providers (code, display_name, api_base_url, config)
VALUES (
  'srs',
  'SRS Local',
  'http://localhost:1985',
  jsonb_build_object(
    'rtmp_base_url', 'rtmp://localhost:1935/live',
    'playback_base_url', 'http://localhost:8080/live'
  )
)
ON CONFLICT (code) DO UPDATE
SET
  display_name = EXCLUDED.display_name,
  api_base_url = EXCLUDED.api_base_url,
  config = EXCLUDED.config;
