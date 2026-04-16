package seed

import "context"

type SeedLocalInput struct {
	Reset           bool
	UserCount       int
	SessionsPerUser int
	StreamKeysPerUser int
	UsersFile       string
}

type SeedLocalOutput struct {
	UsersCreated          int
	MediaProvidersEnsured int
	StreamKeysCreated     int
	PublishSessionsCreated int
}

type SeedLocalService interface {
	Execute(ctx context.Context, input SeedLocalInput) (SeedLocalOutput, error)
}
