package redis

import (
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestRedisUtils_GetSetDeleteString(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	utils := NewRedisUtils(client)

	if err := utils.SetString(t.Context(), "k1", "v1", time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	got, err := utils.GetString(t.Context(), "k1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got != "v1" {
		t.Fatalf("unexpected value: got %q want %q", got, "v1")
	}

	if err := utils.DeleteKey(t.Context(), "k1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = utils.GetString(t.Context(), "k1")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestRedisUtils_GetString_NotFound(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	utils := NewRedisUtils(client)

	_, err := utils.GetString(t.Context(), "missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}
