package biz

import (
	"context"
	"strconv"
	"time"

	"crow/internal/conf"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultTokenType       = "Bearer"
	defaultAccessExpiresIn = 7200
	defaultRefreshTTL      = 24 * time.Hour
)

var (
	// ErrLoginTokenConfig is returned when JWT configuration is missing.
	ErrLoginTokenConfig = errors.InternalServer("LOGIN_TOKEN_CONFIG", "jwt config is invalid")
)

type loginTokenGenerator struct {
	secret    []byte
	expiresIn int64
}

// NewLoginTokenGenerator creates a JWT-backed TokenGenerator.
func NewLoginTokenGenerator(c *conf.Auth) TokenGenerator {
	expiresIn := c.GetJwt().GetExpiresIn()
	if expiresIn <= 0 {
		expiresIn = defaultAccessExpiresIn
	}
	return &loginTokenGenerator{
		secret:    []byte(c.GetJwt().GetSecret()),
		expiresIn: expiresIn,
	}
}

func (g *loginTokenGenerator) Generate(_ context.Context, user *LoginUser) (*LoginResult, error) {
	if user == nil || len(g.secret) == 0 {
		return nil, ErrLoginTokenConfig
	}

	now := time.Now()
	accessToken, err := g.signToken(user, now, time.Duration(g.expiresIn)*time.Second, "access")
	if err != nil {
		return nil, err
	}
	refreshToken, err := g.signToken(user, now, defaultRefreshTTL, "refresh")
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    defaultTokenType,
		ExpiresIn:    g.expiresIn,
		UserID:       user.ID,
	}, nil
}

func (g *loginTokenGenerator) signToken(user *LoginUser, now time.Time, ttl time.Duration, tokenUse string) (string, error) {
	claims := jwt.MapClaims{
		"sub":     strconv.FormatInt(user.ID, 10),
		"account": user.Account,
		"use":     tokenUse,
		"iat":     now.Unix(),
		"exp":     now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secret)
}
