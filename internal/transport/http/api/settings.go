package api

import (
	"net/http"

	"github.com/valentinezhov/lifeos/internal/settings/domain"
)

type timeOfDayJSON struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

func timeOfDayToJSON(t domain.TimeOfDay) timeOfDayJSON {
	return timeOfDayJSON{Hour: t.Hour, Minute: t.Minute}
}

func timeOfDayPtrToJSON(t *domain.TimeOfDay) *timeOfDayJSON {
	if t == nil {
		return nil
	}
	out := timeOfDayToJSON(*t)
	return &out
}

func (rt *Router) getSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.GetSettings == nil {
		writeError(w, http.StatusNotImplemented, "settings is not configured")
		return
	}
	s, err := rt.deps.GetSettings.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"morning_review_at": timeOfDayToJSON(s.MorningReviewAt),
		"evening_review_at": timeOfDayToJSON(s.EveningReviewAt),
		"weekly_review_at":  timeOfDayToJSON(s.WeeklyReviewAt),
		"monthly_review_at": timeOfDayToJSON(s.MonthlyReviewAt),
		"quiet_hours_start": timeOfDayPtrToJSON(s.QuietHoursStart),
		"quiet_hours_end":   timeOfDayPtrToJSON(s.QuietHoursEnd),
		"language":          s.Language,
	})
}

type updateReviewTimeRequest struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

func (rt *Router) updateMorningReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req updateReviewTimeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	at, err := rt.deps.UpdateMorning.Execute(r.Context(), userID, domain.TimeOfDay{
		Hour: req.Hour, Minute: req.Minute,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"morning_review_at": timeOfDayToJSON(at)})
}

func (rt *Router) updateEveningReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req updateReviewTimeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	at, err := rt.deps.UpdateEvening.Execute(r.Context(), userID, domain.TimeOfDay{
		Hour: req.Hour, Minute: req.Minute,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"evening_review_at": timeOfDayToJSON(at)})
}

type updateQuietHoursRequest struct {
	StartHour   int `json:"start_hour"`
	StartMinute int `json:"start_minute"`
	EndHour     int `json:"end_hour"`
	EndMinute   int `json:"end_minute"`
}

func (rt *Router) updateQuietHours(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req updateQuietHoursRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	err := rt.deps.UpdateQuiet.Execute(r.Context(), userID,
		domain.TimeOfDay{Hour: req.StartHour, Minute: req.StartMinute},
		domain.TimeOfDay{Hour: req.EndHour, Minute: req.EndMinute},
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"quiet_hours_start": timeOfDayJSON{Hour: req.StartHour, Minute: req.StartMinute},
		"quiet_hours_end":   timeOfDayJSON{Hour: req.EndHour, Minute: req.EndMinute},
	})
}
