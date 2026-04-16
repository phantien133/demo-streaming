package mediawebhooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"demo-streaming/internal/services/transcode"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var mediaWebhookTestDBSeq int64

func TestSRSOnPublishService_Unauthorized(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	svc := NewGormSRSOnPublishService(db, nil)

	_, err := svc.Execute(context.Background(), SRSOnPublishInput{
		WebhookSecretHeader: "wrong",
		ExpectedSecret:      "expected",
		Stream:              "stream-key",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSRSOnPublishService_BadRequest(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	svc := NewGormSRSOnPublishService(db, nil)

	_, err := svc.Execute(context.Background(), SRSOnPublishInput{
		Stream: "   ",
	})
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestSRSOnPublishService_ForbiddenWhenMissingStreamKey(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	svc := NewGormSRSOnPublishService(db, nil)

	_, err := svc.Execute(context.Background(), SRSOnPublishInput{
		Stream: "not-found",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestSRSOnPublishService_SuccessCreatesBindingAndEnqueuesJob(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	seedMediaWebhookFixtures(t, db)

	enq := &fakeEnqueuer{}
	svc := NewGormSRSOnPublishService(db, enq)

	out, err := svc.Execute(context.Background(), SRSOnPublishInput{
		Stream: "unified-stream-id-1",
		Param:  "token=abc",
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out.SessionID == 0 {
		t.Fatalf("expected non-zero session id")
	}
	if out.Status != "created" {
		t.Fatalf("unexpected status: %s", out.Status)
	}

	if len(enq.jobs) != 1 {
		t.Fatalf("expected 1 enqueued job, got %d", len(enq.jobs))
	}
	if enq.jobs[0].SessionID != out.SessionID {
		t.Fatalf("unexpected enqueued session id: %d", enq.jobs[0].SessionID)
	}
	if enq.jobs[0].PlaybackID != "1" {
		t.Fatalf("unexpected playback id: %s", enq.jobs[0].PlaybackID)
	}
	if enq.jobs[0].StreamKeySecret != "unified-stream-id-1" {
		t.Fatalf("unexpected stream key in job: %s", enq.jobs[0].StreamKeySecret)
	}

	var count int64
	if err := db.Table("media_ingest_bindings").Where("publish_session_id = ?", out.SessionID).Count(&count).Error; err != nil {
		t.Fatalf("count bindings failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 binding, got %d", count)
	}
}

func TestSRSOnPublishService_LegacyResolvesByStreamKeySecret(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	seed := []string{
		`INSERT INTO users(id, email, display_name, password_hash) VALUES (1, 's1@example.com', 's1', 'hash');`,
		`INSERT INTO media_providers(id, code, display_name, config) VALUES (1, 'srs', 'SRS', '{}');`,
		`INSERT INTO stream_keys(id, owner_user_id, stream_key_secret, media_provider_id, label) VALUES (1, 1, 'old-secret-only', 1, 'default');`,
		`INSERT INTO stream_publish_sessions(id, streamer_user_id, playback_id, title, status, playback_url_cdn, media_provider_id, stream_key_id) VALUES (1, 1, 'old-playback-id', 'demo', 'created', 'http://localhost/live/old-playback-id/master.m3u8', 1, 1);`,
	}
	for _, stmt := range seed {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}

	enq := &fakeEnqueuer{}
	svc := NewGormSRSOnPublishService(db, enq)

	out, err := svc.Execute(context.Background(), SRSOnPublishInput{Stream: "old-secret-only"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out.SessionID != 1 {
		t.Fatalf("session id: %d", out.SessionID)
	}
	if len(enq.jobs) != 1 || enq.jobs[0].PlaybackID != "old-playback-id" || enq.jobs[0].StreamKeySecret != "old-secret-only" {
		t.Fatalf("unexpected job: %+v", enq.jobs)
	}
}

func TestSRSOnPublishService_LegacyResolvesByPlaybackIDWhenStreamNameNotAKey(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	seed := []string{
		`INSERT INTO users(id, email, display_name, password_hash) VALUES (1, 's1@example.com', 's1', 'hash');`,
		`INSERT INTO media_providers(id, code, display_name, config) VALUES (1, 'srs', 'SRS', '{}');`,
		`INSERT INTO stream_keys(id, owner_user_id, stream_key_secret, media_provider_id, label) VALUES (1, 1, 'ingest-secret-xyz', 1, 'default');`,
		`INSERT INTO stream_publish_sessions(id, streamer_user_id, playback_id, title, status, playback_url_cdn, media_provider_id, stream_key_id) VALUES (1, 1, 'public-id-legacy', 'demo', 'created', 'http://localhost/live/public-id-legacy/master.m3u8', 1, 1);`,
	}
	for _, stmt := range seed {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}

	enq := &fakeEnqueuer{}
	svc := NewGormSRSOnPublishService(db, enq)

	out, err := svc.Execute(context.Background(), SRSOnPublishInput{Stream: "public-id-legacy"})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out.SessionID != 1 {
		t.Fatalf("session id: %d", out.SessionID)
	}
	if len(enq.jobs) != 1 || enq.jobs[0].PlaybackID != "public-id-legacy" || enq.jobs[0].StreamKeySecret != "ingest-secret-xyz" {
		t.Fatalf("unexpected job: %+v", enq.jobs)
	}
}

func TestSRSOnPublishService_SuccessWhenSessionAlreadyLive(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	seedMediaWebhookFixtures(t, db)
	if err := db.Exec(`UPDATE stream_publish_sessions SET status = 'live' WHERE id = 1`).Error; err != nil {
		t.Fatalf("update session: %v", err)
	}

	enq := &fakeEnqueuer{}
	svc := NewGormSRSOnPublishService(db, enq)

	out, err := svc.Execute(context.Background(), SRSOnPublishInput{
		Stream: "unified-stream-id-1",
		Param:  "token=live",
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if out.SessionID != 1 {
		t.Fatalf("unexpected session id: %d", out.SessionID)
	}
	if out.Status != "live" {
		t.Fatalf("unexpected status: %s", out.Status)
	}
	if len(enq.jobs) != 1 || enq.jobs[0].SessionID != 1 {
		t.Fatalf("unexpected enqueue: %+v", enq.jobs)
	}
}

type fakeEnqueuer struct {
	jobs []transcode.PublishJob
}

func (f *fakeEnqueuer) EnqueuePublishJob(_ context.Context, job transcode.PublishJob) error {
	f.jobs = append(f.jobs, job)
	return nil
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	id := atomic.AddInt64(&mediaWebhookTestDBSeq, 1)
	dsn := fmt.Sprintf("file:mw_%s_%d_%d?mode=memory&cache=private&_foreign_keys=1&_busy_timeout=5000",
		strings.ReplaceAll(t.Name(), "/", "_"), id, time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	stmts := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL UNIQUE, display_name TEXT NOT NULL DEFAULT '', password_hash TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE media_providers (id INTEGER PRIMARY KEY AUTOINCREMENT, code TEXT NOT NULL UNIQUE, display_name TEXT NOT NULL DEFAULT '', api_base_url TEXT, config TEXT NOT NULL DEFAULT '{}', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE stream_keys (id INTEGER PRIMARY KEY AUTOINCREMENT, owner_user_id INTEGER NOT NULL, stream_key_secret TEXT NOT NULL UNIQUE, media_provider_id INTEGER NOT NULL, label TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, revoked_at DATETIME);`,
		`CREATE TABLE stream_publish_sessions (id INTEGER PRIMARY KEY AUTOINCREMENT, streamer_user_id INTEGER NOT NULL, playback_id TEXT NOT NULL UNIQUE, title TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'created', playback_url_cdn TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, started_at DATETIME, ended_at DATETIME, media_provider_id INTEGER NOT NULL, stream_key_id INTEGER NOT NULL);`,
		`CREATE TABLE media_ingest_bindings (id INTEGER PRIMARY KEY AUTOINCREMENT, publish_session_id INTEGER NOT NULL UNIQUE, provider_publish_id TEXT, provider_vhost TEXT, provider_app TEXT, ingest_query_param TEXT, record_local_uri TEXT, last_callback_action TEXT, last_callback_at DATETIME, provider_context TEXT NOT NULL DEFAULT '{}', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema exec failed: %v", err)
		}
	}
	return db
}

func seedMediaWebhookFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()
	seed := []string{
		`INSERT INTO users(id, email, display_name, password_hash) VALUES (1, 's1@example.com', 's1', 'hash');`,
		`INSERT INTO media_providers(id, code, display_name, config) VALUES (1, 'srs', 'SRS', '{}');`,
		`INSERT INTO stream_keys(id, owner_user_id, stream_key_secret, media_provider_id, label) VALUES (1, 1, 'unified-stream-id-1', 1, 'default');`,
		`INSERT INTO stream_publish_sessions(id, streamer_user_id, playback_id, title, status, playback_url_cdn, media_provider_id, stream_key_id) VALUES (1, 1, '1', 'demo', 'created', 'http://localhost/live/1/master.m3u8', 1, 1);`,
	}
	for _, stmt := range seed {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}
}
