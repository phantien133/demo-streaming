package transcode

import (
	"context"
	"errors"
)

var (
	ErrQueueEmpty = errors.New("transcode queue empty")
)

type PublishJob struct {
	SessionID       int64  `json:"session_id"`
	PlaybackID      string `json:"playback_id"`
	StreamKeySecret string `json:"stream_key_secret"`
}

type RenditionOutput struct {
	Name         string
	PlaylistPath string
	Status       string
}

type TranscodeResult struct {
	Renditions []RenditionOutput
}

type PublishJobEnqueuer interface {
	EnqueuePublishJob(ctx context.Context, job PublishJob) error
}

type PublishJobDequeuer interface {
	DequeuePublishJob(ctx context.Context) (PublishJob, error)
}

type MetadataRecorder interface {
	Record(ctx context.Context, job PublishJob, result TranscodeResult) error
}
