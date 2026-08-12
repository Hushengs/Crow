package biz

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type stubLoginRepo struct {
	user *LoginUser
}

func (r *stubLoginRepo) FindByAccount(context.Context, string) (*LoginUser, error) {
	return r.user, nil
}

type stubTokenGenerator struct{}

func (g *stubTokenGenerator) Generate(context.Context, *LoginUser) (*LoginResult, error) {
	return &LoginResult{AccessToken: "token"}, nil
}

func TestLoginAcceptsBcryptPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate bcrypt hash: %v", err)
	}

	uc := NewLoginUsecase(
		&stubLoginRepo{user: &LoginUser{ID: 1, Account: "admin", PasswordHash: string(hash)}},
		&stubTokenGenerator{},
	)

	result, err := uc.Login(context.Background(), &LoginInput{
		Account:  "admin",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if result == nil || result.AccessToken == "" {
		t.Fatalf("expected access token, got %#v", result)
	}
}

func TestLoginAcceptsPlaintextPasswordFallback(t *testing.T) {
	uc := NewLoginUsecase(
		&stubLoginRepo{user: &LoginUser{ID: 1, Account: "admin", PasswordHash: "123456"}},
		&stubTokenGenerator{},
	)

	result, err := uc.Login(context.Background(), &LoginInput{
		Account:  "admin",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if result == nil || result.AccessToken == "" {
		t.Fatalf("expected access token, got %#v", result)
	}
}
