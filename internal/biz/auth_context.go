package biz

import "context"

type adminContextKey struct{}

// AdminPrincipal describes the authenticated administrator attached to a request.
type AdminPrincipal struct {
	ID      int64
	Account string
}

// WithAdminPrincipal stores the current administrator in context.
func WithAdminPrincipal(ctx context.Context, principal *AdminPrincipal) context.Context {
	if principal == nil {
		return ctx
	}
	return context.WithValue(ctx, adminContextKey{}, principal)
}

// AdminPrincipalFromContext loads the current administrator from context.
func AdminPrincipalFromContext(ctx context.Context) (*AdminPrincipal, bool) {
	principal, ok := ctx.Value(adminContextKey{}).(*AdminPrincipal)
	return principal, ok
}
