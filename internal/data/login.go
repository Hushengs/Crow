
package data

import (
	"context"

	"crow/internal/biz"
)

const (
	defaultTokenType = "Bearer"
	defaultExpiresIn = 7200
)

type loginRepo struct {
	data *Data
}

// NewLoginRepo creates a new LoginRepo instance.
func NewLoginRepo(data *Data) biz.LoginRepo {
	return &loginRepo{data: data}
}

func (r *loginRepo) Login(_ context.Context, in *biz.LoginInput) (*biz.LoginResult, error) {
	if in == nil {
		return nil, biz.ErrLoginInvalidArgument
	}

	// Placeholder implementation: return mock tokens until a real user store
	// and token generator are wired in.
	return &biz.LoginResult{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		TokenType:    defaultTokenType,
		ExpiresIn:    defaultExpiresIn,
		UserID:       1,
	}, nil
}
