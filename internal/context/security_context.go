package context

import "context"

type SecurityContext struct {
	UserID   string
	Email    string
	TenantID string
	Role     string
}

type securityContextKeyType struct{}

var securityContextKey = securityContextKeyType{}

func WithSecurityContext(ctx context.Context, sc *SecurityContext) context.Context {
	return context.WithValue(ctx, securityContextKey, sc)
}

func GetSecurityContext(ctx context.Context) *SecurityContext {
	val := ctx.Value(securityContextKey)
	if val == nil {
		return nil
	}
	return val.(*SecurityContext)
}
