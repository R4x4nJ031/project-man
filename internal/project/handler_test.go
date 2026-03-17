package project

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appcontext "github.com/R4x4nJ031/project-man/internal/context"
)

func TestCreateProjectHandler_MissingSecurityContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"name":"demo"}`))
	rec := httptest.NewRecorder()

	CreateProjectHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	assertJSONError(t, rec.Body.Bytes(), "security context missing")
}

func TestCreateProjectHandler_RejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"name":`))
	req = req.WithContext(appcontext.WithSecurityContext(req.Context(), &appcontext.SecurityContext{
		UserID:   "user-123",
		TenantID: "tenant-acme",
		Role:     "admin",
	}))
	rec := httptest.NewRecorder()

	CreateProjectHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	assertJSONError(t, rec.Body.Bytes(), "invalid request")
}

func TestUpdateProjectHandler_RejectsEmptyName(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/projects/update?id=project-1", bytes.NewBufferString(`{"name":""}`))
	req = req.WithContext(appcontext.WithSecurityContext(req.Context(), &appcontext.SecurityContext{
		UserID:   "user-123",
		TenantID: "tenant-acme",
		Role:     "admin",
	}))
	rec := httptest.NewRecorder()

	UpdateProjectHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	assertJSONError(t, rec.Body.Bytes(), "name cannot be empty")
}

func TestDeleteProjectHandler_MissingProjectID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/projects/delete", nil)
	req = req.WithContext(appcontext.WithSecurityContext(req.Context(), &appcontext.SecurityContext{
		UserID:   "user-123",
		TenantID: "tenant-acme",
		Role:     "admin",
	}))
	rec := httptest.NewRecorder()

	DeleteProjectHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	assertJSONError(t, rec.Body.Bytes(), "missing project id")
}

func assertJSONError(t *testing.T, body []byte, want string) {
	t.Helper()

	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["error"] != want {
		t.Fatalf("expected error %q, got %q", want, payload["error"])
	}
}
