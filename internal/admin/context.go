package admin

import (
	"context"
)

type contextKey string

const currentAdminContextKey contextKey = "current_admin"

func ContextWithAdmin(ctx context.Context, a Admin) context.Context {
	return context.WithValue(ctx, currentAdminContextKey, a)
}

func CurrentAdmin(ctx context.Context) (Admin, bool) {
	a, ok := ctx.Value(currentAdminContextKey).(Admin)
	return a, ok
}
