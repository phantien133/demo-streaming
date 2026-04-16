package streamkeys

import (
	"errors"
	"fmt"
	"testing"
	"time"
	"sync/atomic"

	"demo-streaming/internal/database"
	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateStreamKeyService_Execute(t *testing.T) {
	t.Parallel()

	t.Run("creates new key when missing", func(t *testing.T) {
		t.Parallel()
		service, client := newCreateService(t)

		key, expiresIn, err := service.Execute(t.Context(), 123, nil)
		if err != nil {
			t.Fatalf("execute failed: %v", err)
		}
		if key == "" {
			t.Fatalf("expected non-empty key")
		}
		if expiresIn != int64(defaultStreamKeyTTL.Seconds()) {
			t.Fatalf("unexpected expiresIn: got %d", expiresIn)
		}

		stored, err := client.Get(t.Context(), redisKey(123)).Result()
		if err != nil {
			t.Fatalf("failed to read redis key: %v", err)
		}
		if stored != key {
			t.Fatalf("unexpected stored key: got %s want %s", stored, key)
		}
	})

	t.Run("reuses existing key for same user (from DB)", func(t *testing.T) {
		t.Parallel()
		service, client := newCreateService(t)

		seedStreamKey(t, service.db, 123, "existing")

		key, _, err := service.Execute(t.Context(), 123, nil)
		if err != nil {
			t.Fatalf("execute failed: %v", err)
		}
		if key != "existing" {
			t.Fatalf("expected existing key, got %s", key)
		}

		stored, err := client.Get(t.Context(), redisKey(123)).Result()
		if err != nil {
			t.Fatalf("failed to read redis key: %v", err)
		}
		if stored != "existing" {
			t.Fatalf("unexpected redis cached key: got %s want %s", stored, "existing")
		}
	})

	t.Run("respects custom expires_in when provided", func(t *testing.T) {
		t.Parallel()
		service, client := newCreateService(t)
		ttlSeconds := int64(10)

		_, expiresIn, err := service.Execute(t.Context(), 123, &ttlSeconds)
		if err != nil {
			t.Fatalf("execute failed: %v", err)
		}
		if expiresIn != ttlSeconds {
			t.Fatalf("unexpected expiresIn: got %d want %d", expiresIn, ttlSeconds)
		}
		ttl, err := client.TTL(t.Context(), redisKey(123)).Result()
		if err != nil {
			t.Fatalf("ttl failed: %v", err)
		}
		if ttl <= 0 || ttl > 10*time.Second {
			t.Fatalf("unexpected ttl: %v", ttl)
		}
	})
}

func TestRefreshStreamKeyService_Execute(t *testing.T) {
	t.Parallel()

	t.Run("returns not found when key missing", func(t *testing.T) {
		t.Parallel()
		service, _ := newRefreshService(t)

		_, _, err := service.Execute(t.Context(), 123)
		if !errors.Is(err, ErrStreamKeyNotFound) {
			t.Fatalf("expected ErrStreamKeyNotFound, got %v", err)
		}
	})

	t.Run("rotates existing key", func(t *testing.T) {
		t.Parallel()
		service, client := newRefreshService(t)
		seedStreamKey(t, service.db, 123, "old")

		newKey, expiresIn, err := service.Execute(t.Context(), 123)
		if err != nil {
			t.Fatalf("execute failed: %v", err)
		}
		if newKey == "" || newKey == "old" {
			t.Fatalf("expected rotated key, got %s", newKey)
		}
		if expiresIn != int64(defaultStreamKeyTTL.Seconds()) {
			t.Fatalf("unexpected expiresIn: got %d", expiresIn)
		}

		stored, err := client.Get(t.Context(), redisKey(123)).Result()
		if err != nil {
			t.Fatalf("failed to read redis key: %v", err)
		}
		if stored != newKey {
			t.Fatalf("unexpected stored key: got %s want %s", stored, newKey)
		}
	})
}

func TestRevokeStreamKeyService_Execute(t *testing.T) {
	t.Parallel()

	service, client := newRevokeService(t)
	if err := client.Set(t.Context(), redisKey(123), "k", time.Minute).Err(); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	seedStreamKey(t, service.db, 123, "k")

	if err := service.Execute(t.Context(), 123); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if _, err := client.Get(t.Context(), redisKey(123)).Result(); err == nil {
		t.Fatalf("expected key to be deleted")
	}
}

func newCreateService(t *testing.T) (*CreateStreamKeyService, *goredis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	db := newTestDB(t)
	return NewCreateStreamKeyService(db, client), client
}

func newRefreshService(t *testing.T) (*RefreshStreamKeyService, *goredis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	db := newTestDB(t)
	return NewRefreshStreamKeyService(db, client), client
}

func newRevokeService(t *testing.T) (*RevokeStreamKeyService, *goredis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	db := newTestDB(t)
	return NewRevokeStreamKeyService(db, client), client
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Unique in-memory DB per test to avoid parallel locking.
	id := atomic.AddUint64(&testDBCounter, 1)
	dsn := fmt.Sprintf("file:streamkeys_services_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), id)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sqlite sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	// Minimal schema needed by services (media_providers + stream_keys).
	if err := db.Exec(`
PRAGMA busy_timeout = 5000;
CREATE TABLE IF NOT EXISTS media_providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL DEFAULT '',
  api_base_url TEXT,
  config TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS stream_keys (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  owner_user_id INTEGER NOT NULL,
  stream_key_secret TEXT NOT NULL UNIQUE,
  media_provider_id INTEGER NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at DATETIME
);
`).Error; err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// Seed SRS provider row.
	cfg := `{"rtmp_base_url":"rtmp://localhost:1935/live"}`
	if err := db.Create(&database.MediaProvider{
		Code:        "srs",
		DisplayName: "SRS",
		Config:      []byte(cfg),
	}).Error; err != nil {
		t.Fatalf("failed to seed media provider: %v", err)
	}
	return db
}

var testDBCounter uint64

func seedStreamKey(t *testing.T, db *gorm.DB, userID int64, secret string) {
	t.Helper()
	var provider database.MediaProvider
	if err := db.Where("code = ?", "srs").First(&provider).Error; err != nil {
		t.Fatalf("failed to load media provider: %v", err)
	}
	if err := db.Create(&database.StreamKey{
		OwnerUserID:     userID,
		StreamKeySecret: secret,
		MediaProviderID: provider.ID,
	}).Error; err != nil {
		t.Fatalf("failed to seed stream key: %v", err)
	}
}
