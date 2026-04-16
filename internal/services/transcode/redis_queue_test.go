package transcode

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisQueue_EnqueueDequeue(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	queue := NewRedisQueue(client, "test:queue")

	want := PublishJob{
		SessionID:       1,
		PlaybackID:      "1",
		StreamKeySecret: "ingest-secret-1",
	}
	if err := queue.EnqueuePublishJob(context.Background(), want); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	got, err := queue.DequeuePublishJob(context.Background())
	if err != nil {
		t.Fatalf("dequeue failed: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected job: got %+v want %+v", got, want)
	}
}

func TestRedisQueue_DequeueEmptyReturnsErrQueueEmpty(t *testing.T) {
	t.Parallel()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	queue := NewRedisQueue(client, "test:queue-empty")
	queue.dequeueTimeout = 10 * time.Millisecond

	_, err := queue.DequeuePublishJob(context.Background())
	if !errors.Is(err, ErrQueueEmpty) {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}
}
