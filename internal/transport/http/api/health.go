package api

import (
	"errors"
	"net/http"
	"time"

	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	"github.com/valentinezhov/lifeos/internal/health/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
)

type weightLogJSON struct {
	ID       string  `json:"id"`
	WeightKg float64 `json:"weight_kg"`
	LoggedAt string  `json:"logged_at"`
}

func weightToJSON(dto healthapp.WeightLogDTO) weightLogJSON {
	return weightLogJSON{
		ID:       dto.ID.String(),
		WeightKg: dto.WeightKg,
		LoggedAt: dto.LoggedAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listWeights(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ListWeights == nil {
		writeError(w, http.StatusNotImplemented, "list weights is not configured")
		return
	}
	items, err := rt.deps.ListWeights.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]weightLogJSON, 0, len(items))
	for _, item := range items {
		out = append(out, weightToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"weights": out})
}

func (rt *Router) latestWeight(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.GetLatestWeight == nil {
		writeError(w, http.StatusNotImplemented, "latest weight is not configured")
		return
	}
	dto, err := rt.deps.GetLatestWeight.Execute(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no weight records")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, weightToJSON(dto))
}

type recordWeightRequest struct {
	WeightKg float64 `json:"weight_kg"`
}

func (rt *Router) recordWeight(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.RecordWeight == nil {
		writeError(w, http.StatusNotImplemented, "record weight is not configured")
		return
	}
	var req recordWeightRequest
	if err := decodeJSON(r, &req); err != nil || req.WeightKg <= 0 {
		writeError(w, http.StatusBadRequest, "weight_kg is required")
		return
	}
	dto, err := rt.deps.RecordWeight.Execute(r.Context(), healthapp.RecordWeightInput{
		UserID: userID, WeightKg: req.WeightKg, Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, weightToJSON(dto))
}

type stepLogJSON struct {
	ID       string `json:"id"`
	Steps    int32  `json:"steps"`
	LoggedAt string `json:"logged_at"`
}

func stepToJSON(dto healthapp.StepLogDTO) stepLogJSON {
	return stepLogJSON{
		ID:       dto.ID.String(),
		Steps:    dto.Steps,
		LoggedAt: dto.LoggedAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listSteps(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ListSteps == nil {
		writeError(w, http.StatusNotImplemented, "list steps is not configured")
		return
	}
	items, err := rt.deps.ListSteps.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]stepLogJSON, 0, len(items))
	for _, item := range items {
		out = append(out, stepToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"steps": out})
}

func (rt *Router) latestSteps(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.GetLatestSteps == nil {
		writeError(w, http.StatusNotImplemented, "latest steps is not configured")
		return
	}
	dto, err := rt.deps.GetLatestSteps.Execute(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no step records")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stepToJSON(dto))
}

type recordStepsRequest struct {
	Steps int32 `json:"steps"`
}

func (rt *Router) recordSteps(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.RecordSteps == nil {
		writeError(w, http.StatusNotImplemented, "record steps is not configured")
		return
	}
	var req recordStepsRequest
	if err := decodeJSON(r, &req); err != nil || req.Steps <= 0 {
		writeError(w, http.StatusBadRequest, "steps is required")
		return
	}
	dto, err := rt.deps.RecordSteps.Execute(r.Context(), healthapp.RecordStepsInput{
		UserID: userID, Steps: req.Steps, Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, stepToJSON(dto))
}

type sleepLogJSON struct {
	ID              string  `json:"id"`
	DurationMinutes int32   `json:"duration_minutes"`
	DurationHours   float64 `json:"duration_hours"`
	LoggedAt        string  `json:"logged_at"`
}

func sleepToJSON(dto healthapp.SleepLogDTO) sleepLogJSON {
	return sleepLogJSON{
		ID:              dto.ID.String(),
		DurationMinutes: dto.DurationMinutes,
		DurationHours:   dto.DurationHours(),
		LoggedAt:        dto.LoggedAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listSleep(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ListSleep == nil {
		writeError(w, http.StatusNotImplemented, "list sleep is not configured")
		return
	}
	items, err := rt.deps.ListSleep.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]sleepLogJSON, 0, len(items))
	for _, item := range items {
		out = append(out, sleepToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"sleep": out})
}

func (rt *Router) latestSleep(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.GetLatestSleep == nil {
		writeError(w, http.StatusNotImplemented, "latest sleep is not configured")
		return
	}
	dto, err := rt.deps.GetLatestSleep.Execute(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no sleep records")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sleepToJSON(dto))
}

type recordSleepRequest struct {
	DurationHours   *float64 `json:"duration_hours"`
	DurationMinutes *int32   `json:"duration_minutes"`
}

func (rt *Router) recordSleep(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.RecordSleep == nil {
		writeError(w, http.StatusNotImplemented, "record sleep is not configured")
		return
	}
	var req recordSleepRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var mins int32
	switch {
	case req.DurationMinutes != nil && *req.DurationMinutes > 0:
		mins = *req.DurationMinutes
	case req.DurationHours != nil && *req.DurationHours > 0:
		mins = domain.MinutesFromHours(*req.DurationHours)
	default:
		writeError(w, http.StatusBadRequest, "duration_hours or duration_minutes is required")
		return
	}
	dto, err := rt.deps.RecordSleep.Execute(r.Context(), healthapp.RecordSleepInput{
		UserID: userID, DurationMinutes: mins, Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sleepToJSON(dto))
}
