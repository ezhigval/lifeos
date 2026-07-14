package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type habitJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Frequency string `json:"frequency"`
}

func habitToJSON(dto habitsapp.HabitDTO) habitJSON {
	return habitJSON{
		ID:        dto.ID.String(),
		Name:      dto.Name,
		Frequency: string(dto.Frequency),
	}
}

type habitDayJSON struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	TodayCompleted bool   `json:"today_completed"`
	Streak         int    `json:"streak"`
}

func habitDayToJSON(dto habitsapp.HabitDayDTO) habitDayJSON {
	return habitDayJSON{
		ID:             dto.ID.String(),
		Name:           dto.Name,
		TodayCompleted: dto.TodayCompleted,
		Streak:         dto.Streak,
	}
}

func (rt *Router) listHabitsToday(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
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
	Name string `json:"name"`
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
	dto, err := rt.deps.CreateHabit.Execute(r.Context(), habitsapp.CreateHabitInput{
		UserID: userID,
		Name:   strings.TrimSpace(req.Name),
		Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, habitToJSON(dto))
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
	result, err := rt.deps.TrackHabit.ExecuteByID(r.Context(), userID, habitID, events.SourceHTTP)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name":   result.Name,
		"streak": result.Streak,
	})
}
