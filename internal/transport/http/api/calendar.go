package api

import (
	"net/http"
	"strings"
	"time"

	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
)

type eventJSON struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	StartsAt string `json:"starts_at"`
}

func eventToJSON(dto calendarapp.EventDTO) eventJSON {
	return eventJSON{
		ID:       dto.ID.String(),
		Title:    dto.Title,
		StartsAt: dto.StartsAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listCalendarToday(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := rt.deps.ListCalendar.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]eventJSON, 0, len(items))
	for _, item := range items {
		out = append(out, eventToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": out})
}

type createEventRequest struct {
	Title    string `json:"title"`
	StartsAt string `json:"starts_at"`
}

func (rt *Router) createCalendarEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createEventRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Title) == "" || req.StartsAt == "" {
		writeError(w, http.StatusBadRequest, "title and starts_at are required")
		return
	}
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid starts_at, use RFC3339")
		return
	}
	dto, err := rt.deps.CreateEvent.Execute(r.Context(), calendarapp.CreateEventInput{
		UserID:   userID,
		Title:    strings.TrimSpace(req.Title),
		StartsAt: startsAt,
		Source:   events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, eventToJSON(dto))
}
