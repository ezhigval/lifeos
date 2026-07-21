package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	"github.com/valentinezhov/lifeos/internal/spheres/domain"
)

type sphereJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int32  `json:"sort_order"`
	CreatedAt string `json:"created_at"`
}

func sphereToJSON(dto spheresapp.SphereDTO) sphereJSON {
	return sphereJSON{
		ID:        dto.ID.String(),
		Name:      dto.Name,
		SortOrder: dto.SortOrder,
		CreatedAt: dto.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listSpheres(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := rt.deps.ListSpheres.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]sphereJSON, 0, len(items))
	for _, item := range items {
		out = append(out, sphereToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"spheres": out})
}

type createSphereRequest struct {
	Name      string `json:"name"`
	SortOrder *int32 `json:"sort_order"`
}

func (rt *Router) createSphere(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createSphereRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	dto, err := rt.deps.CreateSphere.Execute(r.Context(), spheresapp.CreateSphereInput{
		UserID: userID, Name: strings.TrimSpace(req.Name), SortOrder: req.SortOrder, Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sphereToJSON(dto))
}

type updateSphereRequest struct {
	Name      string `json:"name"`
	SortOrder int32  `json:"sort_order"`
}

func (rt *Router) updateSphere(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	sphereID, err := ids.ParseSphereID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid sphere id")
		return
	}
	var req updateSphereRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	dto, err := rt.deps.UpdateSphere.Execute(r.Context(), spheresapp.UpdateSphereInput{
		UserID: userID, SphereID: sphereID,
		Name: strings.TrimSpace(req.Name), SortOrder: req.SortOrder, Source: events.SourceHTTP,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "sphere not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sphereToJSON(dto))
}

func (rt *Router) deleteSphere(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	sphereID, err := ids.ParseSphereID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid sphere id")
		return
	}
	dto, err := rt.deps.DeleteSphere.Execute(r.Context(), spheresapp.DeleteSphereInput{
		UserID: userID, SphereID: sphereID, Source: events.SourceHTTP,
	})
	if err != nil {
		if errors.Is(err, spheresapp.ErrSphereNotFound) {
			writeError(w, http.StatusNotFound, "sphere not found")
			return
		}
		if errors.Is(err, domain.ErrHasProjects) {
			writeError(w, http.StatusConflict, "sphere has linked projects")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sphereToJSON(dto))
}
