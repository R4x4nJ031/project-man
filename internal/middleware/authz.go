package middleware

import (
	"net/http"

	"github.com/R4x4nJ031/project-man/internal/context"
	"github.com/R4x4nJ031/project-man/internal/response"
)

type Permission string

const (
	ReadProject   Permission = "read_project"
	CreateProject Permission = "create_project"
	UpdateProject Permission = "update_project"
	DeleteProject Permission = "delete_project"
)

var rolePermissions = map[string][]Permission{
	"admin": {
		ReadProject,
		CreateProject,
		UpdateProject,
		DeleteProject,
	},
	"developer": {
		ReadProject,
		CreateProject,
		UpdateProject,
	},
	"viewer": {
		ReadProject,
	},
}

func RequirePermission(permission Permission) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			secCtx := context.GetSecurityContext(r.Context())
			if secCtx == nil {
				response.Error(w, http.StatusInternalServerError, "security context missing")
				return
			}

			role := secCtx.Role

			perms := rolePermissions[role]

			allowed := false

			for _, p := range perms {
				if p == permission {
					allowed = true
					break
				}
			}

			if !allowed {
				response.Error(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
