package api

import (
	"context"
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
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	SortOrder  int32    `json:"sort_order"`
	CreatedAt  string   `json:"created_at"`
	DomainLink string   `json:"domain_link,omitempty"`
	RefID      string   `json:"ref_id,omitempty"`
	RefName    string   `json:"ref_name,omitempty"`
	Career     bool     `json:"career,omitempty"`
	Contacts   []string `json:"contacts,omitempty"`
}

func sphereToJSON(dto spheresapp.SphereDTO) sphereJSON {
	return sphereJSON{
		ID:        dto.ID.String(),
		Name:      dto.Name,
		SortOrder: dto.SortOrder,
		CreatedAt: dto.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// careerBridge enriches a Career sphere with its workspace domain link
// (sphere_domain_links, TASK-010 WS-15 rule 3 / TASK-011 п.8 мост Карьера→воркспейс).
// Best-effort: any store error just leaves the sphere un-enriched.
func (rt *Router) careerBridge(ctx context.Context, userID ids.UserID, item *sphereJSON) {
	if rt.deps.GetSphereDomainLink == nil || !rt.deps.GetSphereDomainLink.SphereIsCareer(ctx, userID, item.Name) {
		return
	}
	item.Career = true
	sphereID, perr := ids.ParseSphereID(item.ID)
	if perr != nil {
		return
	}
	link, found, err := rt.deps.GetSphereDomainLink.Execute(ctx, userID, sphereID)
	if err != nil {
		return
	}
	if found {
		item.DomainLink = link.LinkType
		item.RefID = link.RefID.String()
		item.RefName = link.RefName
	}
	if rt.deps.ListCareerContacts != nil {
		names, err := rt.deps.ListCareerContacts.Execute(ctx, userID)
		if err == nil && len(names) > 0 {
			item.Contacts = names
		}
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
		j := sphereToJSON(item)
		rt.careerBridge(r.Context(), userID, &j)
		out = append(out, j)
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
