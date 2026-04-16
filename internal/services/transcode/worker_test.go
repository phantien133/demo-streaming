package transcode

import (
	"context"
	"errors"
	"testing"
)

func TestWorker_RunProcessesJobAndRecordsMetadata(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue := &stubQueue{
		jobs: []PublishJob{
			{SessionID: 1, PlaybackID: "1", StreamKeySecret: "rtmp-stream-name-1"},
		},
	}
	exec := &stubExecutor{
		result: TranscodeResult{
			Renditions: []RenditionOutput{
				{Name: "720p", PlaylistPath: "p1/720p/index.m3u8", Status: "ready"},
			},
		},
		onExecute: cancel,
	}
	rec := &stubRecorder{}

	w := NewWorker(queue, exec, rec)
	err := w.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if exec.calls != 1 {
		t.Fatalf("expected 1 execute call, got %d", exec.calls)
	}
	if rec.calls != 1 {
		t.Fatalf("expected 1 record call, got %d", rec.calls)
	}
}

type stubQueue struct {
	jobs []PublishJob
}

func (s *stubQueue) DequeuePublishJob(_ context.Context) (PublishJob, error) {
	if len(s.jobs) == 0 {
		return PublishJob{}, ErrQueueEmpty
	}
	job := s.jobs[0]
	s.jobs = s.jobs[1:]
	return job, nil
}

type stubExecutor struct {
	calls    int
	result   TranscodeResult
	err      error
	onExecute func()
}

func (s *stubExecutor) Execute(_ context.Context, _ PublishJob) (TranscodeResult, error) {
	s.calls++
	if s.onExecute != nil {
		s.onExecute()
	}
	return s.result, s.err
}

type stubRecorder struct {
	calls int
}

func (s *stubRecorder) Record(_ context.Context, _ PublishJob, _ TranscodeResult) error {
	s.calls++
	return nil
}
