package streamkeys

import (
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestCreateStreamKeyService_Execute(t *testing.T) {
	t.Parallel()

	t.Run("creates new key when missing", func(t *testing.T) {
		t.Parallel()
		service, client := newCreateService(t)

		key, expiresIn, err := service.Execute(t.Context(), 123)
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

	t.Run("reuses existing key for same user", func(t *testing.T) {
		t.Parallel()
		service, client := newCreateService(t)
		if err := client.Set(t.Context(), redisKey(123), "existing", time.Minute).Err(); err != nil {
			t.Fatalf("seed failed: %v", err)
		}

		key, _, err := service.Execute(t.Context(), 123)
		if err != nil {
			t.Fatalf("execute failed: %v", err)
		}
		if key != "existing" {
			t.Fatalf("expected existing key, got %s", key)
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
		if err := client.Set(t.Context(), redisKey(123), "old", time.Minute).Err(); err != nil {
			t.Fatalf("seed failed: %v", err)
		}

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
	return NewCreateStreamKeyService(client), client
}

func newRefreshService(t *testing.T) (*RefreshStreamKeyService, *goredis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return NewRefreshStreamKeyService(client), client
}

func newRevokeService(t *testing.T) (*RevokeStreamKeyService, *goredis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return NewRevokeStreamKeyService(client), client
}
