package streampublishsessions

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPublishSessionNotFound  = errors.New("publish session not found")
	ErrPublishSessionForbidden = errors.New("publish session forbidden")
	ErrPublishSessionBadState  = errors.New("publish session invalid state")
	ErrPublishSessionLiveExists = errors.New("publish session live exists")
	ErrMediaProviderNotFound   = errors.New("media provider not found")
	ErrMediaProviderMisconfigured = errors.New("media provider misconfigured")
)

type CreateInput struct {
	StreamerUserID int64
	Title          string
}

type CreateOutput struct {
	SessionID      int64
	PlaybackID     string
	Status         string
	PlaybackURLCDN string
	Ingest         CreateOutputIngest
}

type CreateOutputIngest struct {
	Provider  string
	RTMPURL   string
	StreamKey string
}

type CreateService interface {
	Execute(ctx context.Context, input CreateInput) (CreateOutput, error)
}

type StartInput struct {
	SessionID      int64
	StreamerUserID int64
}

type StartOutput struct {
	SessionID int64
	Status    string
	StartedAt time.Time
}

type StartService interface {
	Execute(ctx context.Context, input StartInput) (StartOutput, error)
}

type StopLiveInput struct {
	// StreamerUserID is always required (used for authorization).
	StreamerUserID int64

	// SessionID is optional:
	// - nil: stop ALL live sessions for this streamer (used by CreateService)
	// - non-nil: stop exactly this session (used by Stop API)
	SessionID *int64
}

type StopLiveOutput struct {
	StoppedSessionIDs []int64
	EndedAt           time.Time
}

type StopLiveService interface {
	Execute(ctx context.Context, input StopLiveInput) (StopLiveOutput, error)
}

type ListInput struct {
	StreamerUserID int64
	Page           int
	Limit          int
}

type ListItem struct {
	SessionID      int64
	PlaybackID     string
	Title          string
	Status         string
	PlaybackURLCDN string
	CreatedAt      time.Time
	StartedAt      *time.Time
	EndedAt        *time.Time
	Renditions     []ListRenditionItem
}

type ListRenditionItem struct {
	Name         string
	PlaylistPath string
	Status       string
}

type ListOutput struct {
	Items []ListItem
	Page  int
	Limit int
	Total int64
}

type ListService interface {
	Execute(ctx context.Context, input ListInput) (ListOutput, error)
}
