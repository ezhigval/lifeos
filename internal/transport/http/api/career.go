package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type contactJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Company   string `json:"company"`
	Role      string `json:"role"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func contactToJSON(dto careerapp.ContactDTO) contactJSON {
	return contactJSON{
		ID:        dto.ID.String(),
		Name:      dto.Name,
		Company:   dto.Company,
		Role:      dto.Role,
		Notes:     dto.Notes,
		CreatedAt: dto.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listContacts(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	var items []careerapp.ContactDTO
	var err error
	if query != "" {
		items, err = rt.deps.SearchContacts.Execute(r.Context(), careerapp.SearchContactsInput{
			UserID: userID, Query: query,
		})
	} else {
		items, err = rt.deps.ListContacts.Execute(r.Context(), userID)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]contactJSON, 0, len(items))
	for _, item := range items {
		out = append(out, contactToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"contacts": out, "query": query})
}

type createContactRequest struct {
	Name    string `json:"name"`
	Company string `json:"company"`
	Role    string `json:"role"`
	Notes   string `json:"notes"`
}

func (rt *Router) createContact(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createContactRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	dto, err := rt.deps.CreateContact.Execute(r.Context(), careerapp.CreateContactInput{
		UserID: userID,
		Name:   strings.TrimSpace(req.Name),
		Company: strings.TrimSpace(req.Company),
		Role:    strings.TrimSpace(req.Role),
		Notes:   strings.TrimSpace(req.Notes),
		Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, contactToJSON(dto))
}

func (rt *Router) deleteContact(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	contactID, err := ids.ParseContactID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid contact id")
		return
	}
	dto, err := rt.deps.DeleteContact.Execute(r.Context(), careerapp.DeleteContactInput{
		UserID: userID, ContactID: contactID, Source: events.SourceHTTP,
	})
	if err != nil {
		if errors.Is(err, careerapp.ErrContactNotFound) {
			writeError(w, http.StatusNotFound, "contact not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, contactToJSON(dto))
}

type skillJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Level     string `json:"level"`
	CreatedAt string `json:"created_at"`
}

func skillToJSON(dto careerapp.SkillDTO) skillJSON {
	return skillJSON{
		ID:        dto.ID.String(),
		Name:      dto.Name,
		Level:     dto.Level,
		CreatedAt: dto.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listSkills(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	var items []careerapp.SkillDTO
	var err error
	if query != "" {
		items, err = rt.deps.SearchSkills.Execute(r.Context(), careerapp.SearchSkillsInput{
			UserID: userID, Query: query,
		})
	} else {
		items, err = rt.deps.ListSkills.Execute(r.Context(), userID)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]skillJSON, 0, len(items))
	for _, item := range items {
		out = append(out, skillToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"skills": out, "query": query})
}

type createSkillRequest struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

func (rt *Router) createSkill(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createSkillRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	dto, err := rt.deps.CreateSkill.Execute(r.Context(), careerapp.CreateSkillInput{
		UserID: userID,
		Name:   strings.TrimSpace(req.Name),
		Level:  strings.TrimSpace(req.Level),
		Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, skillToJSON(dto))
}

func (rt *Router) deleteSkill(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	skillID, err := ids.ParseSkillID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid skill id")
		return
	}
	dto, err := rt.deps.DeleteSkill.Execute(r.Context(), careerapp.DeleteSkillInput{
		UserID: userID, SkillID: skillID, Source: events.SourceHTTP,
	})
	if err != nil {
		if errors.Is(err, careerapp.ErrSkillNotFound) {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, skillToJSON(dto))
}
