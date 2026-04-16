package transcode

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormMetadataStore_RecordUpsert(t *testing.T) {
	t.Parallel()
	db := newMetadataTestDB(t)
	store := NewGormMetadataStore(db)

	job := PublishJob{SessionID: 10, PlaybackID: "a", StreamKeySecret: "sk-not-playback"}
	first := TranscodeResult{
		Renditions: []RenditionOutput{
			{Name: "720p", PlaylistPath: "playback-10/720p/index.m3u8", Status: "ready"},
		},
	}
	if err := store.Record(context.Background(), job, first); err != nil {
		t.Fatalf("record first failed: %v", err)
	}

	second := TranscodeResult{
		Renditions: []RenditionOutput{
			{Name: "720p", PlaylistPath: "playback-10/720p/v2/index.m3u8", Status: "failed"},
			{Name: "480p", PlaylistPath: "playback-10/480p/index.m3u8", Status: "ready"},
		},
	}
	if err := store.Record(context.Background(), job, second); err != nil {
		t.Fatalf("record second failed: %v", err)
	}

	type row struct {
		RenditionName string
		PlaylistPath  string
		Status        string
	}
	var rows []row
	if err := db.Table("transcode_renditions").Select("rendition_name, playlist_path, status").Order("rendition_name asc").Find(&rows).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1].RenditionName != "720p" || rows[1].PlaylistPath != "playback-10/720p/v2/index.m3u8" || rows[1].Status != "failed" {
		t.Fatalf("unexpected upserted 720p row: %+v", rows[1])
	}
}

func newMetadataTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:transcode_metadata_%d?mode=memory&cache=shared&_foreign_keys=1&_busy_timeout=5000", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	stmts := []string{
		`CREATE TABLE stream_publish_sessions (id INTEGER PRIMARY KEY AUTOINCREMENT);`,
		`CREATE TABLE transcode_renditions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			publish_session_id INTEGER NOT NULL,
			playback_id TEXT NOT NULL,
			rendition_name TEXT NOT NULL,
			playlist_path TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'ready',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (publish_session_id, rendition_name)
		);`,
		`INSERT INTO stream_publish_sessions(id) VALUES (10);`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema exec failed: %v", err)
		}
	}
	return db
}
