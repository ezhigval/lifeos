package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	projectsdomain "github.com/valentinezhov/lifeos/internal/projects/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type projectJSON struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Outcome     string   `json:"outcome,omitempty"`
	Status      string   `json:"status"`
	TargetValue *string  `json:"target_value,omitempty"`
	CurrentValue string  `json:"current_value,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	TargetDate  *string  `json:"target_date,omitempty"`
	SphereIDs   []string `json:"sphere_ids"`
}

func projectToJSON(dto projectsapp.ProjectDTO) projectJSON {
	out := projectJSON{
		ID:           dto.ID.String(),
		Name:         dto.Name,
		Outcome:      dto.Outcome,
		Status:       string(dto.Status),
		CurrentValue: dto.CurrentValue.String(),
		Unit:         dto.Unit,
		SphereIDs:    sphereIDsToStrings(dto.SphereIDs),
	}
	if dto.TargetValue != nil {
		s := dto.TargetValue.String()
		out.TargetValue = &s
	}
	if dto.TargetDate != nil {
		s := dto.TargetDate.Format("2006-01-02")
		out.TargetDate = &s
	}
	return out
}

func sphereIDsToStrings(ids []ids.SphereID) []string {
	if len(ids) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func parseSphereIDs(raw []string) ([]ids.SphereID, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]ids.SphereID, 0, len(raw))
	for _, s := range raw {
		id, err := ids.ParseSphereID(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func (rt *Router) listProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	in := projectsapp.ListProjectsInput{UserID: userID}
	if raw := r.URL.Query().Get("sphere_id"); raw != "" {
		sphereID, err := ids.ParseSphereID(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid sphere_id")
			return
		}
		in.SphereID = sphereID
	}
	items, err := rt.deps.ListProjects.Execute(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]projectJSON, 0, len(items))
	for _, item := range items {
		out = append(out, projectToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": out})
}

type createProjectRequest struct {
	Name        string   `json:"name"`
	SphereIDs   []string `json:"sphere_ids"`
	Outcome     string   `json:"outcome"`
	TargetValue *string  `json:"target_value"`
	Unit        string   `json:"unit"`
	TargetDate  *string  `json:"target_date"`
}

func (rt *Router) createProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createProjectRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	sphereIDs, err := parseSphereIDs(req.SphereIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid sphere_ids")
		return
	}
	var targetValue *decimal.Decimal
	if req.TargetValue != nil && *req.TargetValue != "" {
		targetValue, err = projectsapp.ParseTarget(*req.TargetValue)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid target_value")
			return
		}
	}
	var targetDate *time.Time
	if req.TargetDate != nil && *req.TargetDate != "" {
		t, err := time.Parse("2006-01-02", *req.TargetDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid target_date")
			return
		}
		targetDate = &t
	}
	dto, err := rt.deps.CreateProject.Execute(r.Context(), projectsapp.CreateProjectInput{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		SphereIDs:   sphereIDs,
		Outcome:     strings.TrimSpace(req.Outcome),
		TargetValue: targetValue,
		Unit:        strings.TrimSpace(req.Unit),
		TargetDate:  targetDate,
		Source:      events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, projectToJSON(dto))
}

func (rt *Router) projectProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var projectID ids.ProjectID
	if raw := r.URL.Query().Get("project_id"); raw != "" {
		var err error
		projectID, err = ids.ParseProjectID(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid project_id")
			return
		}
	}
	p, err := rt.deps.ProjectProg.Execute(r.Context(), userID, projectID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project_id": p.ProjectID.String(),
		"name":       p.Name,
		"target":     p.Target,
		"current":    p.Current,
		"remaining":  p.Remaining,
		"percent":    p.Percent,
		"unit":       p.Unit,
		"has_target": p.HasTarget,
	})
}

func (rt *Router) listProjectTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	projectID, err := ids.ParseProjectID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	items, err := rt.deps.ListProjectTasks.Execute(r.Context(), userID, projectID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out := make([]taskJSON, 0, len(items))
	for _, item := range items {
		out = append(out, taskToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": out})
}

func (rt *Router) archiveProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	projectID, err := ids.ParseProjectID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	dto, err := rt.deps.ArchiveProject.Execute(r.Context(), projectsapp.ArchiveProjectInput{
		UserID: userID, ProjectID: projectID, Source: events.SourceHTTP,
	})
	if errors.Is(err, projectsdomain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if errors.Is(err, projectsdomain.ErrNotActive) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, projectToJSON(dto))
}
