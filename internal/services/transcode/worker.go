package transcode

import (
	"context"
	"errors"
	"log"
	"time"
)

type JobExecutor interface {
	Execute(ctx context.Context, job PublishJob) (TranscodeResult, error)
}

type Worker struct {
	queue    PublishJobDequeuer
	executor JobExecutor
	recorder MetadataRecorder
}

func NewWorker(queue PublishJobDequeuer, executor JobExecutor, recorder MetadataRecorder) *Worker {
	return &Worker{
		queue:    queue,
		executor: executor,
		recorder: recorder,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		job, err := w.queue.DequeuePublishJob(ctx)
		if err != nil {
			if errors.Is(err, ErrQueueEmpty) {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return err
		}

		log.Printf("[transcode] processing publish session=%d playback_id=%s", job.SessionID, job.PlaybackID)
		result, err := w.executor.Execute(ctx, job)
		if err != nil {
			log.Printf("[transcode] failed session=%d playback_id=%s err=%v", job.SessionID, job.PlaybackID, err)
			continue
		}
		if w.recorder != nil {
			if err := w.recorder.Record(ctx, job, result); err != nil {
				log.Printf("[transcode] metadata record failed session=%d playback_id=%s err=%v", job.SessionID, job.PlaybackID, err)
				continue
			}
		}
		log.Printf("[transcode] completed session=%d playback_id=%s", job.SessionID, job.PlaybackID)
	}
}
