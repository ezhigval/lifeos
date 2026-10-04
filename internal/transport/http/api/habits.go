package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	habitsdomain "github.com/valentinezhov/lifeos/internal/habits/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// parseDatePtr parses an optional "YYYY-MM-DD" body field into a UTC day pointer.
// ok=false means the value was present but malformed.
func parseDatePtr(field string, v *string) (*time.Time, bool) {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil, true
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(*v))
	if err != nil {
		return nil, false
	}
	return &t, true
}

type habitJSON struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Frequency string  `json:"frequency"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
}

func dateStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func habitToJSON(dto habitsapp.HabitDTO) habitJSON {
	return habitJSON{
		ID:        dto.ID.String(),
		Name:      dto.Name,
		Frequency: string(dto.Frequency),
		StartDate: dateStr(dto.StartDate),
		EndDate:   dateStr(dto.EndDate),
	}
}

type habitDayJSON struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	TodayCompleted bool    `json:"today_completed"`
	Streak         int     `json:"streak"`
	StartDate      *string `json:"start_date,omitempty"`
	EndDate        *string `json:"end_date,omitempty"`
	Active         bool    `json:"active"`
}

func habitDayToJSON(dto habitsapp.HabitDayDTO) habitDayJSON {
	return habitDayJSON{
		ID:             dto.ID.String(),
		Name:           dto.Name,
		TodayCompleted: dto.TodayCompleted,
		Streak:         dto.Streak,
		StartDate:      dateStr(dto.StartDate),
		EndDate:        dateStr(dto.EndDate),
		Active:         dto.Active,
	}
}

func (rt *Router) listHabitsToday(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ListHabits == nil {
		writeError(w, http.StatusNotImplemented, "list habits is not configured")
		return
	}
	items, err := rt.deps.ListHabits.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]habitDayJSON, 0, len(items))
	for _, item := range items {
		out = append(out, habitDayToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"habits": out})
}

type createHabitRequest struct {
	Name      string  `json:"name"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

func (rt *Router) createHabit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createHabitRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	start, okStart := parseDatePtr("start_date", req.StartDate)
	if !okStart {
		writeError(w, http.StatusBadRequest, "start_date must be YYYY-MM-DD")
		return
	}
	end, okEnd := parseDatePtr("end_date", req.EndDate)
	if !okEnd {
		writeError(w, http.StatusBadRequest, "end_date must be YYYY-MM-DD")
		return
	}
	dto, err := rt.deps.CreateHabit.Execute(r.Context(), habitsapp.CreateHabitInput{
		UserID:    userID,
		Name:      strings.TrimSpace(req.Name),
		StartDate: start,
		EndDate:   end,
		Source:    events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, habitToJSON(dto))
}

type updateHabitRequest struct {
	Name      *string `json:"name"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

func (rt *Router) updateHabit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.UpdateHabit == nil {
		writeError(w, http.StatusNotImplemented, "update habit is not configured")
		return
	}
	habitID, err := ids.ParseHabitID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}
	var req updateHabitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := ""
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}
	start, okStart := parseDatePtr("start_date", req.StartDate)
	if !okStart {
		writeError(w, http.StatusBadRequest, "start_date must be YYYY-MM-DD")
		return
	}
	end, okEnd := parseDatePtr("end_date", req.EndDate)
	if !okEnd {
		writeError(w, http.StatusBadRequest, "end_date must be YYYY-MM-DD")
		return
	}
	dto, err := rt.deps.UpdateHabit.Execute(r.Context(), habitsapp.UpdateHabitInput{
		UserID:    userID,
		HabitID:   habitID,
		Name:      name,
		StartDate: start,
		EndDate:   end,
		Source:    events.SourceHTTP,
	})
	if err != nil {
		switch {
		case errors.Is(err, habitsdomain.ErrNotFound):
			writeError(w, http.StatusNotFound, "habit not found")
		case errors.Is(err, habitsdomain.ErrEmptyName), errors.Is(err, habitsdomain.ErrInvalidDeadline):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, habitToJSON(dto))
}

func (rt *Router) deleteHabit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.DeleteHabit == nil {
		writeError(w, http.StatusNotImplemented, "delete habit is not configured")
		return
	}
	habitID, err := ids.ParseHabitID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}
	if err := rt.deps.DeleteHabit.Execute(r.Context(), habitsapp.DeleteHabitInput{
		UserID:  userID,
		HabitID: habitID,
		Source:  events.SourceHTTP,
	}); err != nil {
		if errors.Is(err, habitsdomain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habit not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *Router) trackHabit(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	habitID, err := ids.ParseHabitID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}
	if rt.deps.TrackHabit == nil {
		writeError(w, http.StatusNotImplemented, "track habit is not configured")
		return
	}
	result, err := rt.deps.TrackHabit.ExecuteByID(r.Context(), userID, habitID, events.SourceHTTP)
	if err != nil {
		if errors.Is(err, habitsdomain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habit not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":   result.Name,
		"streak": result.Streak,
	})
}
