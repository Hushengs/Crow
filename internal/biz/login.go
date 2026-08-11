package biz

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	// ErrLoginInvalidArgument is returned when the login request is invalid.
	ErrLoginInvalidArgument = errors.BadRequest("LOGIN_INVALID_ARGUMENT", "invalid login argument")
	// ErrLoginInvalidCredentials is returned when the account or password is incorrect.
	ErrLoginInvalidCredentials = errors.Unauthorized("LOGIN_INVALID_CREDENTIALS", "invalid account or password")
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

// LoginUser is the user snapshot loaded from persistence.
type LoginUser struct {
	ID           int64
	Account      string
	PasswordHash string
}

// LoginRepo provides the persistence boundary for looking up users.
type LoginRepo interface {
	FindByAccount(context.Context, string) (*LoginUser, error)
}

// TokenGenerator signs login results for authenticated users.
type TokenGenerator interface {
	Generate(context.Context, *LoginUser) (*LoginResult, error)
}

// LoginUsecase is the login usecase.
type LoginUsecase struct {
	repo   LoginRepo
	token  TokenGenerator
}

// NewLoginUsecase creates a new LoginUsecase.
func NewLoginUsecase(repo LoginRepo, token TokenGenerator) *LoginUsecase {
	return &LoginUsecase{repo: repo, token: token}
}

// Login authenticates a user and returns login tokens.
func (uc *LoginUsecase) Login(ctx context.Context, in *LoginInput) (*LoginResult, error) {
	if in == nil || strings.TrimSpace(in.Account) == "" || strings.TrimSpace(in.Password) == "" {
		return nil, ErrLoginInvalidArgument
	}

	user, err := uc.repo.FindByAccount(ctx, strings.TrimSpace(in.Account))
	if err != nil {
		return nil, err
	}
	if user == nil || user.PasswordHash != in.Password {
		return nil, ErrLoginInvalidCredentials
	}
	return uc.token.Generate(ctx, user)
}
