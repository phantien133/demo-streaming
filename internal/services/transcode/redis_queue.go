package transcode

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultQueueKey       = "transcode:publish_jobs"
	defaultDequeueTimeout = 3 * time.Second
)

type RedisQueue struct {
	client         *redis.Client
	key            string
	dequeueTimeout time.Duration
}

func NewRedisQueue(client *redis.Client, key string) *RedisQueue {
	queueKey := key
	if queueKey == "" {
		queueKey = defaultQueueKey
	}
	return &RedisQueue{
		client:         client,
		key:            queueKey,
		dequeueTimeout: defaultDequeueTimeout,
	}
}

func (q *RedisQueue) EnqueuePublishJob(ctx context.Context, job PublishJob) error {
	raw, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, q.key, raw).Err()
}

func (q *RedisQueue) DequeuePublishJob(ctx context.Context) (PublishJob, error) {
	values, err := q.client.BLPop(ctx, q.dequeueTimeout, q.key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return PublishJob{}, ErrQueueEmpty
		}
		// When context deadline is exceeded we surface empty queue semantics.
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return PublishJob{}, ErrQueueEmpty
		}
		return PublishJob{}, err
	}
	if len(values) < 2 {
		return PublishJob{}, ErrQueueEmpty
	}

	var job PublishJob
	if err := json.Unmarshal([]byte(values[1]), &job); err != nil {
		return PublishJob{}, err
	}
	return job, nil
}
