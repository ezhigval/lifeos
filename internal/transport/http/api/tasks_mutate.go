package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
)

func (rt *Router) getTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.GetTask == nil {
		writeError(w, http.StatusNotImplemented, "get task is not configured")
		return
	}
	taskID, err := ids.ParseTaskID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	dto, err := rt.deps.GetTask.Execute(r.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, tasksapp.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, taskToJSON(dto))
}

func (rt *Router) archiveTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ArchiveTask == nil {
		writeError(w, http.StatusNotImplemented, "archive task is not configured")
		return
	}
	taskID, err := ids.ParseTaskID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	dto, err := rt.deps.ArchiveTask.Execute(r.Context(), tasksapp.ArchiveTaskInput{
		UserID: userID, TaskID: taskID, Source: events.SourceHTTP,
	})
	if err != nil {
		if errors.Is(err, tasksapp.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rt.cancelTaskReminders(r.Context(), userID, taskID)
	writeJSON(w, http.StatusOK, taskToJSON(dto))
}

func (rt *Router) deleteTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.DeleteTask == nil {
		writeError(w, http.StatusNotImplemented, "delete task is not configured")
		return
	}
	taskID, err := ids.ParseTaskID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	err = rt.deps.DeleteTask.Execute(r.Context(), tasksapp.DeleteTaskInput{
		UserID: userID, TaskID: taskID, Source: events.SourceHTTP,
	})
	if err != nil {
		if errors.Is(err, tasksapp.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rt.cancelTaskReminders(r.Context(), userID, taskID)
	w.WriteHeader(http.StatusNoContent)
}
