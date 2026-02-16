package middleware

type contextKey string

const (
	CtxUserID      contextKey = "user_id"
	CtxTenantID    contextKey = "tenant_id"
	CtxPermissions contextKey = "permissions"
)
