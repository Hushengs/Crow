package biz

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	// ErrLoginInvalidArgument is returned when the login request is invalid.
	ErrLoginInvalidArgument = errors.BadRequest("LOGIN_INVALID_ARGUMENT", "invalid login argument")
)

// LoginInput is the domain input for a login request.
type LoginInput struct {
	Account  string
	Password string
}

// LoginResult is the domain result for a successful login.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
	UserID       int64
}

// LoginRepo provides the persistence boundary for authentication.
type LoginRepo interface {
	Login(context.Context, *LoginInput) (*LoginResult, error)
}

// LoginUsecase is the login usecase.
type LoginUsecase struct {
	repo LoginRepo
}

// NewLoginUsecase creates a new LoginUsecase.
func NewLoginUsecase(repo LoginRepo) *LoginUsecase {
	return &LoginUsecase{repo: repo}
}

// Login authenticates a user and returns login tokens.
func (uc *LoginUsecase) Login(ctx context.Context, in *LoginInput) (*LoginResult, error) {
	if in == nil || strings.TrimSpace(in.Account) == "" || strings.TrimSpace(in.Password) == "" {
		return nil, ErrLoginInvalidArgument
	}
	return uc.repo.Login(ctx, in)
}
