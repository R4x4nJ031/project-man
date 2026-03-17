package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appcontext "github.com/R4x4nJ031/project-man/internal/context"
)

func TestRequirePermission_ForbiddenForViewer(t *testing.T) {
	handler := RequirePermission(CreateProject)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/projects", nil)
	req = req.WithContext(appcontext.WithSecurityContext(req.Context(), &appcontext.SecurityContext{
		UserID:   "user-123",
		TenantID: "tenant-acme",
		Role:     "viewer",
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["error"] != "forbidden" {
		t.Fatalf("expected forbidden error, got %q", body["error"])
	}
}

func TestRequirePermission_AllowsAdmin(t *testing.T) {
	var called bool

	handler := RequirePermission(DeleteProject)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/projects/delete", nil)
	req = req.WithContext(appcontext.WithSecurityContext(context.Background(), &appcontext.SecurityContext{
		UserID:   "user-123",
		TenantID: "tenant-acme",
		Role:     "admin",
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
