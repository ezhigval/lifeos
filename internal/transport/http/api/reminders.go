package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
)

type reminderJSON struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	FireAt  string `json:"fire_at"`
	Status  string `json:"status"`
}

func reminderToJSON(dto notifapp.ReminderDTO) reminderJSON {
	return reminderJSON{
		ID:      dto.ID,
		Message: dto.Message,
		FireAt:  dto.FireAt.UTC().Format(time.RFC3339),
		Status:  dto.Status,
	}
}

func (rt *Router) listReminders(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ListReminders == nil {
		writeError(w, http.StatusNotImplemented, "list reminders is not configured")
		return
	}
	items, err := rt.deps.ListReminders.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]reminderJSON, 0, len(items))
	for _, item := range items {
		out = append(out, reminderToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"reminders": out})
}

type createReminderRequest struct {
	Message string `json:"message"`
	FireAt  string `json:"fire_at"`
}

func (rt *Router) createReminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ScheduleReminder == nil {
		writeError(w, http.StatusNotImplemented, "schedule reminder is not configured")
		return
	}
	var req createReminderRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Message) == "" || req.FireAt == "" {
		writeError(w, http.StatusBadRequest, "message and fire_at are required")
		return
	}
	fireAt, err := time.Parse(time.RFC3339, req.FireAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fire_at, use RFC3339")
		return
	}
	if fireAt.Before(time.Now().UTC()) {
		writeError(w, http.StatusBadRequest, "fire_at must be in the future")
		return
	}
	dto, err := rt.deps.ScheduleReminder.Execute(r.Context(), notifapp.ScheduleReminderInput{
		UserID:  userID,
		Message: strings.TrimSpace(req.Message),
		FireAt:  fireAt,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, reminderToJSON(dto))
}

func (rt *Router) cancelReminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.CancelReminder == nil {
		writeError(w, http.StatusNotImplemented, "cancel reminder is not configured")
		return
	}
	reminderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid reminder id")
		return
	}
	dto, err := rt.deps.CancelReminder.Execute(r.Context(), notifapp.CancelReminderInput{
		UserID: userID, ReminderID: reminderID,
	})
	if err != nil {
		if errors.Is(err, notifapp.ErrReminderNotFound) {
			writeError(w, http.StatusNotFound, "reminder not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, reminderToJSON(dto))
}
