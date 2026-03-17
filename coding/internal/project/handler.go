package project

import (
	"encoding/json"
	"net/http"

	appcontext "github.com/R4x4nJ031/project-man/internal/context"
	"github.com/R4x4nJ031/project-man/internal/response"
	"github.com/google/uuid"
)

func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {

	secCtx := appcontext.GetSecurityContext(r.Context())
	if secCtx == nil {
		response.Error(w, http.StatusInternalServerError, "security context missing")
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, "name cannot be empty")
		return
	}

	p := Project{
		ID:       uuid.New().String(),
		TenantID: secCtx.TenantID,
		Name:     req.Name,
	}

	err = CreateProjectWithAudit(r.Context(), &p, secCtx.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	response.JSON(w, http.StatusOK, p)
}

func ListProjectsHandler(w http.ResponseWriter, r *http.Request) {

	secCtx := appcontext.GetSecurityContext(r.Context())
	if secCtx == nil {
		response.Error(w, http.StatusInternalServerError, "security context missing")
		return
	}

	projects, err := ListProjects(r.Context(), secCtx.TenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list projects")
		return
	}

	response.JSON(w, http.StatusOK, projects)
}
func GetProjectHandler(w http.ResponseWriter, r *http.Request) {

	secCtx := appcontext.GetSecurityContext(r.Context())
	if secCtx == nil {
		response.Error(w, http.StatusInternalServerError, "security context missing")
		return
	}

	projectID := r.URL.Query().Get("id")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "missing project id")
		return
	}

	project, err := GetProjectByID(
		r.Context(),
		projectID,
		secCtx.TenantID,
	)

	if err != nil {
		response.Error(w, http.StatusNotFound, "project not found")
		return
	}

	response.JSON(w, http.StatusOK, project)
}

func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {

	secCtx := appcontext.GetSecurityContext(r.Context())
	if secCtx == nil {
		response.Error(w, http.StatusInternalServerError, "security context missing")
		return
	}

	projectID := r.URL.Query().Get("id")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "missing project id")
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, "name cannot be empty")
		return
	}

	updated, err := UpdateProjectWithAudit(r.Context(), projectID, secCtx.TenantID, req.Name, secCtx.UserID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "project not found")
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {

	secCtx := appcontext.GetSecurityContext(r.Context())
	if secCtx == nil {
		response.Error(w, http.StatusInternalServerError, "security context missing")
		return
	}

	projectID := r.URL.Query().Get("id")
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "missing project id")
		return
	}

	err := DeleteProjectWithAudit(r.Context(), projectID, secCtx.TenantID, secCtx.UserID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "project not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
