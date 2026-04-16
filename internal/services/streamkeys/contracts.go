package streamkeys

import "context"

type CreateService interface {
	Execute(ctx context.Context, userID int64, expiresInSeconds *int64) (streamKey string, expiresIn int64, err error)
}

type RefreshService interface {
	Execute(ctx context.Context, userID int64) (streamKey string, expiresIn int64, err error)
}

type RevokeService interface {
	Execute(ctx context.Context, userID int64) error
}
