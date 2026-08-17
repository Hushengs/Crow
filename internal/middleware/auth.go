package middleware

import (
	"context"
	"strconv"
	"strings"

	"crow/internal/biz"
	"crow/internal/conf"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v3/middleware"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/golang-jwt/jwt/v5"
)

const bearerPrefix = "Bearer "

// NewAdminAuth authenticates admin HTTP requests and stores the current principal in context.
func NewAdminAuth(c *conf.Auth) kratosmiddleware.Middleware {
	secret := []byte(c.GetJwt().GetSecret())

	return func(next kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			httpReq, ok := kratoshttp.RequestFromServerContext(ctx)
			if !ok || !shouldProtectAdminPath(httpReq.URL.Path) {
				return next(ctx, req)
			}
			if len(secret) == 0 {
				return nil, kratoserrors.InternalServer("AUTH_TOKEN_CONFIG", "jwt config is invalid")
			}

			authHeader := strings.TrimSpace(httpReq.Header.Get("Authorization"))
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return nil, kratoserrors.Unauthorized("AUTH_UNAUTHORIZED", "missing bearer token")
			}

			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
			principal, err := parseAdminPrincipal(tokenString, secret)
			if err != nil {
				return nil, err
			}

			return next(biz.WithAdminPrincipal(ctx, principal), req)
		}
	}
}

func parseAdminPrincipal(tokenString string, secret []byte) (*biz.AdminPrincipal, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, kratoserrors.Unauthorized("AUTH_UNAUTHORIZED", "invalid token signature")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, kratoserrors.Unauthorized("AUTH_UNAUTHORIZED", "invalid access token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, kratoserrors.Unauthorized("AUTH_UNAUTHORIZED", "invalid access token")
	}

	tokenUse, _ := claims["use"].(string)
	if tokenUse != "access" {
		return nil, kratoserrors.Unauthorized("AUTH_UNAUTHORIZED", "invalid access token")
	}

	sub, _ := claims["sub"].(string)
	account, _ := claims["account"].(string)
	adminID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil || adminID <= 0 {
		return nil, kratoserrors.Unauthorized("AUTH_UNAUTHORIZED", "invalid access token")
	}

	return &biz.AdminPrincipal{
		ID:      adminID,
		Account: strings.TrimSpace(account),
	}, nil
}

func shouldProtectAdminPath(path string) bool {
	switch {
	case strings.HasPrefix(path, "/v1/admins"):
		return true
	case strings.HasPrefix(path, "/v1/roles"):
		return true
	case strings.HasPrefix(path, "/v1/permissions"):
		return true
	case strings.HasPrefix(path, "/v1/admin-roles"):
		return true
	case strings.HasPrefix(path, "/v1/group-permissions"):
		return true
	case strings.HasPrefix(path, "/v1/admin-operation-logs"):
		return true
	case strings.HasPrefix(path, "/v1/system-logs"):
		return true
	case strings.HasPrefix(path, "/v1/cps"):
		return true
	case strings.HasPrefix(path, "/v1/sps"):
		return true
	case strings.HasPrefix(path, "/v1/cp-sps"):
		return true
	default:
		return false
	}
}
