package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type noteJSON struct {
	ID         string   `json:"id"`
	Body       string   `json:"body"`
	Tags       []string `json:"tags"`
	TargetType string   `json:"target_type,omitempty"`
	TargetID   string   `json:"target_id,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

func noteToJSON(dto knowledgeapp.NoteDTO) noteJSON {
	tags := dto.Tags
	if tags == nil {
		tags = []string{}
	}
	return noteJSON{
		ID:         dto.ID.String(),
		Body:       dto.Body,
		Tags:       tags,
		TargetType: dto.TargetType,
		TargetID:   dto.TargetID,
		CreatedAt:  dto.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (rt *Router) listNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	tt := strings.TrimSpace(r.URL.Query().Get("target_type"))
	ti := strings.TrimSpace(r.URL.Query().Get("target_id"))
	var items []knowledgeapp.NoteDTO
	var err error
	switch {
	case tt != "" && ti != "":
		// TASK-011 item 5: reverse sync — notes linked to a task/event/reminder.
		if rt.deps.ListNotesByTarget == nil {
			writeError(w, http.StatusNotImplemented, "list notes by target is not configured")
			return
		}
		items, err = rt.deps.ListNotesByTarget.Execute(r.Context(), userID, tt, ti)
	case query != "":
		if rt.deps.SearchNotes == nil {
			writeError(w, http.StatusNotImplemented, "search notes is not configured")
			return
		}
		items, err = rt.deps.SearchNotes.Execute(r.Context(), knowledgeapp.SearchNotesInput{
			UserID: userID, Query: query,
		})
	default:
		if rt.deps.ListNotes == nil {
			writeError(w, http.StatusNotImplemented, "list notes is not configured")
			return
		}
		items, err = rt.deps.ListNotes.Execute(r.Context(), knowledgeapp.ListNotesInput{
			UserID: userID, Tag: tag,
		})
	}
	if err != nil {
		status := http.StatusInternalServerError
		if tt != "" || strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}
	out := make([]noteJSON, 0, len(items))
	for _, item := range items {
		out = append(out, noteToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"notes": out, "query": query, "tag": tag})
}

type createNoteRequest struct {
	Body       string   `json:"body"`
	Tags       []string `json:"tags"`
	TargetType string   `json:"target_type"`
	TargetID   string   `json:"target_id"`
}

func (rt *Router) createNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.CreateNote == nil {
		writeError(w, http.StatusNotImplemented, "create note is not configured")
		return
	}
	var req createNoteRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Body) == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}
	dto, err := rt.deps.CreateNote.Execute(r.Context(), knowledgeapp.CreateNoteInput{
		UserID:     userID,
		Body:       strings.TrimSpace(req.Body),
		Tags:       req.Tags,
		Source:     events.SourceHTTP,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, noteToJSON(dto))
}

func (rt *Router) getNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.GetNote == nil {
		writeError(w, http.StatusNotImplemented, "get note is not configured")
		return
	}
	noteID, err := ids.ParseNoteID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	dto, err := rt.deps.GetNote.Execute(r.Context(), userID, noteID)
	if errors.Is(err, knowledgeapp.ErrNoteNotFound) {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, noteToJSON(dto))
}

type updateNoteRequest struct {
	Body string `json:"body"`
}

func (rt *Router) updateNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.UpdateNote == nil {
		writeError(w, http.StatusNotImplemented, "update note is not configured")
		return
	}
	noteID, err := ids.ParseNoteID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	var req updateNoteRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Body) == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}
	dto, err := rt.deps.UpdateNote.Execute(r.Context(), userID, noteID, req.Body)
	if errors.Is(err, knowledgeapp.ErrNoteNotFound) {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, noteToJSON(dto))
}

func (rt *Router) deleteNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.DeleteNote == nil {
		writeError(w, http.StatusNotImplemented, "delete note is not configured")
		return
	}
	noteID, err := ids.ParseNoteID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	dto, err := rt.deps.DeleteNote.Execute(r.Context(), knowledgeapp.DeleteNoteInput{
		UserID: userID, NoteID: noteID, Source: events.SourceHTTP,
	})
	if err != nil {
		if errors.Is(err, knowledgeapp.ErrNoteNotFound) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, noteToJSON(dto))
}
