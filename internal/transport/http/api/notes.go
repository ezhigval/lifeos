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
	ID        string   `json:"id"`
	Body      string   `json:"body"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

func noteToJSON(dto knowledgeapp.NoteDTO) noteJSON {
	tags := dto.Tags
	if tags == nil {
		tags = []string{}
	}
	return noteJSON{
		ID:        dto.ID.String(),
		Body:      dto.Body,
		Tags:      tags,
		CreatedAt: dto.CreatedAt.UTC().Format(time.RFC3339),
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
	var items []knowledgeapp.NoteDTO
	var err error
	if query != "" {
		items, err = rt.deps.SearchNotes.Execute(r.Context(), knowledgeapp.SearchNotesInput{
			UserID: userID, Query: query,
		})
	} else {
		items, err = rt.deps.ListNotes.Execute(r.Context(), knowledgeapp.ListNotesInput{
			UserID: userID, Tag: tag,
		})
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]noteJSON, 0, len(items))
	for _, item := range items {
		out = append(out, noteToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"notes": out, "query": query, "tag": tag})
}

type createNoteRequest struct {
	Body string   `json:"body"`
	Tags []string `json:"tags"`
}

func (rt *Router) createNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createNoteRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Body) == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}
	dto, err := rt.deps.CreateNote.Execute(r.Context(), knowledgeapp.CreateNoteInput{
		UserID: userID,
		Body:   strings.TrimSpace(req.Body),
		Tags:   req.Tags,
		Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, noteToJSON(dto))
}

func (rt *Router) deleteNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
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
