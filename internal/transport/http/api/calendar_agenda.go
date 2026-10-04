package api

import (
	"net/http"
	"sort"
	"strings"
	"time"

	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

// agendaItemJSON is a unified calendar item: tasks, events, reminders and notes
// merged into one time-sorted stream for the month/week/day views.
type agendaItemJSON struct {
	Type       string   `json:"type"` // task | event | reminder | note
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	StartsAt   string   `json:"starts_at"` // RFC3339 UTC
	AllDay     bool     `json:"all_day,omitempty"`
	ProjectIDs []string `json:"project_ids,omitempty"`
	SphereIDs  []string `json:"sphere_ids,omitempty"`
	Done       bool     `json:"done,omitempty"`
}

// parseAgendaWindow reads from/to (YYYY-MM-DD) and view (month|week|day).
// Defaults: view=week, window = current week (Mon..next Mon exclusive).
func parseAgendaWindow(r *http.Request, now time.Time) (from, to time.Time, view string, ok bool) {
	view = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("view")))
	switch view {
	case "day", "week", "month":
	default:
		view = "week"
	}

	fromStr := strings.TrimSpace(r.URL.Query().Get("from"))
	toStr := strings.TrimSpace(r.URL.Query().Get("to"))

	const dateLayout = "2006-01-02"
	if fromStr == "" || toStr == "" {
		base := now.UTC()
		if fromStr != "" {
			parsed, err := time.Parse(dateLayout, fromStr)
			if err != nil {
				return time.Time{}, time.Time{}, "", false
			}
			base = parsed
		}
		switch view {
		case "day":
			from = base
			to = base.AddDate(0, 0, 1)
		case "month":
			from = time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, time.UTC)
			to = from.AddDate(0, 1, 0)
		default: // week starting Monday
			wd := int(base.Weekday()) // Sunday=0
			offset := (wd + 6) % 7
			from = base.AddDate(0, 0, -offset)
			to = from.AddDate(0, 0, 7)
		}
		return from, to, view, true
	}

	parsedFrom, err := time.Parse(dateLayout, fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, "", false
	}
	parsedTo, err := time.Parse(dateLayout, toStr)
	if err != nil {
		return time.Time{}, time.Time{}, "", false
	}
	if parsedTo.Before(parsedFrom) {
		return time.Time{}, time.Time{}, "", false
	}
	// to is an inclusive date -> exclusive bound.
	return parsedFrom, parsedTo.AddDate(0, 0, 1), view, true
}

func queryCSV(r *http.Request, key string) map[string]bool {
	set := map[string]bool{}
	for _, raw := range r.URL.Query()[key] {
		for _, part := range strings.Split(raw, ",") {
			if v := strings.TrimSpace(part); v != "" {
				set[v] = true
			}
		}
	}
	return set
}

// listCalendarAgenda returns merged items for the requested window with
// optional filters by projects and item types (workspace filter arrives
// with Stage 4 workspaces).
func (rt *Router) listCalendarAgenda(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	from, to, view, ok := parseAgendaWindow(r, time.Now().UTC())
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid from/to, use YYYY-MM-DD")
		return
	}

	projectFilter := queryCSV(r, "projects")
	sphereFilter := queryCSV(r, "spheres")
	typesFilter := queryCSV(r, "types")
	wantsType := func(t string) bool {
		return len(typesFilter) == 0 || typesFilter[t]
	}

	items := make([]agendaItemJSON, 0, 64)

	// Tasks with due_date in [from, to), including completed ones.
	if wantsType("task") && rt.deps.ListTasksBetween != nil {
		tasks, err := rt.deps.ListTasksBetween.Execute(r.Context(), userID, from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, t := range tasks {
			if t.DueDate == nil {
				continue
			}
			projectIDs := make([]string, 0, len(t.ProjectIDs))
			for _, pid := range t.ProjectIDs {
				projectIDs = append(projectIDs, pid.String())
			}
			if len(projectFilter) > 0 {
				match := false
				for _, pid := range projectIDs {
					if projectFilter[pid] {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
			sphereIDs := make([]string, 0, len(t.SphereIDs))
			for _, sid := range t.SphereIDs {
				sphereIDs = append(sphereIDs, sid.String())
			}
			if len(sphereFilter) > 0 {
				match := false
				for _, sid := range sphereIDs {
					if sphereFilter[sid] {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
			items = append(items, agendaItemJSON{
				Type: "task", ID: t.ID.String(), Title: t.Title,
				StartsAt: t.DueDate.UTC().Format(time.RFC3339), AllDay: true,
				ProjectIDs: projectIDs, SphereIDs: sphereIDs, Done: t.Status == taskdomain.StatusDone,
			})
		}
	}

	// Calendar events with starts_at in [from, to).
	if wantsType("event") && rt.deps.ListEventsBetween != nil {
		events, err := rt.deps.ListEventsBetween.Execute(r.Context(), userID, from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, e := range events {
			items = append(items, agendaItemJSON{
				Type: "event", ID: e.ID.String(), Title: e.Title,
				StartsAt: e.StartsAt.UTC().Format(time.RFC3339),
			})
		}
	}

	// Reminders pending fire_at in [from, to).
	if wantsType("reminder") && rt.deps.ListReminders != nil {
		reminders, err := rt.deps.ListReminders.Execute(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, rem := range reminders {
			fireAt := rem.FireAt.UTC()
			if fireAt.Before(from) || !fireAt.Before(to) {
				continue
			}
			items = append(items, agendaItemJSON{
				Type: "reminder", ID: rem.ID, Title: rem.Message,
				StartsAt: fireAt.Format(time.RFC3339),
			})
		}
	}

	// Notes created in window (so linked notes surface on their day).
	if wantsType("note") && rt.deps.ListNotesBetween != nil {
		notes, err := rt.deps.ListNotesBetween.Execute(r.Context(), userID, from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		for _, n := range notes {
			title := truncateRunes(n.Body, 80)
			items = append(items, agendaItemJSON{
				Type: "note", ID: n.ID.String(), Title: title,
				StartsAt: n.CreatedAt.UTC().Format(time.RFC3339), AllDay: true,
			})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].StartsAt != items[j].StartsAt {
			return items[i].StartsAt < items[j].StartsAt
		}
		return items[i].Title < items[j].Title
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"view":  view,
		"from":  from.Format("2006-01-02"),
		"to":    to.AddDate(0, 0, -1).Format("2006-01-02"),
		"items": items,
	})
}

// truncateRunes shortens s to max runes (UTF-8 safe), cutting at the first
// line break when present — used for note titles in the agenda feed.
func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '\n'); idx > 0 {
		s = strings.TrimSpace(s[:idx])
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
