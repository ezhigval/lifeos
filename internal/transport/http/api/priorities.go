package api

import (
	"net/http"
)

func (rt *Router) listPriorities(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := rt.deps.Priorities.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"kind":   item.Kind,
			"title":  item.Title,
			"score":  item.Score,
			"detail": item.Detail,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"priorities": out})
}
