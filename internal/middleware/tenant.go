package middleware

import (
	stdcontext "context"
	"net/http"

	appcontext "github.com/R4x4nJ031/project-man/internal/context"
	"github.com/R4x4nJ031/project-man/internal/database"
	"github.com/R4x4nJ031/project-man/internal/response"
)

func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			response.Error(w, http.StatusForbidden, "missing tenant id")
			return
		}

		secCtx := appcontext.GetSecurityContext(r.Context())
		if secCtx == nil {
			response.Error(w, http.StatusInternalServerError, "security context missing")
			return
		}

		var role string

		err := database.DB.QueryRow(
			stdcontext.Background(),
			`SELECT role FROM memberships WHERE user_id=$1 AND tenant_id=$2`,
			secCtx.UserID,
			tenantID,
		).Scan(&role)

		if err != nil {
			response.Error(w, http.StatusForbidden, "forbidden tenant access")
			return
		}

		secCtx.TenantID = tenantID
		secCtx.Role = role

		next.ServeHTTP(w, r)
	})
}
