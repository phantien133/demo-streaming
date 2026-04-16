package auth

import "context"

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	UserID int64
	Email  string
}

type LoginService interface {
	Execute(ctx context.Context, input LoginInput) (LoginOutput, error)
}
