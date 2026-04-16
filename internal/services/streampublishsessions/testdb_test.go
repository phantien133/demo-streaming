package streampublishsessions

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use a unique in-memory database per test to avoid cross-test interference.
	dsn := fmt.Sprintf("file:streampublishsessions_%d?mode=memory&cache=shared&_foreign_keys=1&_busy_timeout=5000", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	// Use SQLite-compatible schema (avoid Postgres-specific defaults/types from GORM tags).
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL DEFAULT '',
			password_hash TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS media_providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL DEFAULT '',
			api_base_url TEXT,
			config TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS stream_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_user_id INTEGER NOT NULL,
			stream_key_secret TEXT NOT NULL UNIQUE,
			media_provider_id INTEGER NOT NULL,
			label TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			revoked_at DATETIME,
			FOREIGN KEY(owner_user_id) REFERENCES users(id) ON DELETE RESTRICT,
			FOREIGN KEY(media_provider_id) REFERENCES media_providers(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE IF NOT EXISTS stream_publish_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			streamer_user_id INTEGER NOT NULL,
			playback_id TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'created',
			playback_url_cdn TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			ended_at DATETIME,
			media_provider_id INTEGER NOT NULL,
			stream_key_id INTEGER NOT NULL,
			FOREIGN KEY(streamer_user_id) REFERENCES users(id) ON DELETE RESTRICT,
			FOREIGN KEY(media_provider_id) REFERENCES media_providers(id) ON DELETE RESTRICT,
			FOREIGN KEY(stream_key_id) REFERENCES stream_keys(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE IF NOT EXISTS view_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			publish_session_id INTEGER NOT NULL,
			viewer_user_id INTEGER,
			viewer_ref TEXT NOT NULL DEFAULT '',
			client_type TEXT NOT NULL DEFAULT 'web',
			joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			left_at DATETIME,
			last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(publish_session_id) REFERENCES stream_publish_sessions(id) ON DELETE CASCADE,
			FOREIGN KEY(viewer_user_id) REFERENCES users(id) ON DELETE SET NULL
		);`,
		`CREATE TABLE IF NOT EXISTS transcode_renditions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			publish_session_id INTEGER NOT NULL,
			playback_id TEXT NOT NULL,
			rendition_name TEXT NOT NULL,
			playlist_path TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'ready',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (publish_session_id, rendition_name),
			FOREIGN KEY(publish_session_id) REFERENCES stream_publish_sessions(id) ON DELETE CASCADE
		);`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema exec failed: %v (stmt=%s)", err, stmt)
		}
	}
	return db
}

