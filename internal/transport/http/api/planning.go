package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// TriageProposer exposes the overloaded-day triage use case over HTTP
// (MA-C6). Implemented by *planningapp.TriageOverloadedDay.
type TriageProposer interface {
	Propose(ctx context.Context, userID ids.UserID) (string, []ids.TaskID, error)
	ApplyDefer(ctx context.Context, userID ids.UserID, taskIDs []ids.TaskID) (int, error)
}

func (rt *Router) triageProposal(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.Triage == nil {
		writeError(w, http.StatusServiceUnavailable, "triage unavailable")
		return
	}
	text, lowIDs, err := rt.deps.Triage.Propose(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	idsOut := make([]string, 0, len(lowIDs))
	for _, id := range lowIDs {
		idsOut = append(idsOut, id.String())
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"text":             text,
		"low_priority_ids": idsOut,
	})
}

func (rt *Router) triageDefer(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.Triage == nil {
		writeError(w, http.StatusServiceUnavailable, "triage unavailable")
		return
	}
	var body struct {
		TaskIDs []string `json:"task_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	taskIDs := make([]ids.TaskID, 0, len(body.TaskIDs))
	for _, raw := range body.TaskIDs {
		id, err := ids.ParseTaskID(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid task id: "+raw)
			return
		}
		taskIDs = append(taskIDs, id)
	}
	moved, err := rt.deps.Triage.ApplyDefer(r.Context(), userID, taskIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"moved": moved})
}
