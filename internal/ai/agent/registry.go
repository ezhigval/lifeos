package agent

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// Built-in tool names. Handlers are injected via Register from outside.
const (
	ToolTaskCreate           = "task.create"
	ToolTaskListToday        = "task.list_today"
	ToolTaskComplete         = "task.complete"
	ToolFinanceExpense       = "finance.expense"
	ToolFinanceIncome        = "finance.income"
	ToolFinanceListDebts     = "finance.list_debts"
	ToolFinanceCreateDebt    = "finance.create_debt"
	ToolFinancePayDebt       = "finance.pay_debt"
	ToolFinanceCashFlow      = "finance.cash_flow"
	ToolReminderCreate       = "reminder.create"
	ToolReminderCancel       = "reminder.cancel"
	ToolHabitCreate          = "habit.create"
	ToolHabitTrack           = "habit.track"
	ToolHabitList             = "habit.list"
	ToolNoteCreate           = "note.create"
	ToolNoteList             = "note.list"
	ToolNoteSearch           = "note.search"
	ToolCalendarCreate       = "calendar.create"
	ToolCalendarListToday    = "calendar.list_today"
	ToolProjectCreate        = "project.create"
	ToolProjectList          = "project.list"
	ToolHealthRecordWeight   = "health.record_weight"
	ToolHealthLatestWeight   = "health.latest_weight"
	ToolCareerContactCreate  = "career.contact_create"
	ToolCareerContactList    = "career.contact_list"
	ToolMemorySave           = "memory.save"
	ToolMemoryRecall         = "memory.recall"
)

type Tool struct {
	Name        string
	Description string
	// Parameters is a JSON schema-ish description for the prompt.
	Parameters string
}

type ToolHandler func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error)

type Registry struct {
	mu    sync.RWMutex
	tools map[string]registeredTool
}

type registeredTool struct {
	Tool
	handler ToolHandler
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]registeredTool)}
}

func (r *Registry) Register(name, desc, params string, h ToolHandler) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[name] = registeredTool{
		Tool: Tool{
			Name:        name,
			Description: desc,
			Parameters:  params,
		},
		handler: h,
	}
}

func (r *Registry) Get(name string) (Tool, ToolHandler, bool) {
	if r == nil {
		return Tool{}, nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	if !ok {
		return Tool{}, nil, false
	}
	return t.Tool, t.handler, true
}

func (r *Registry) List() []Tool {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t.Tool)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) Call(ctx context.Context, userID ids.UserID, name string, args map[string]any) (string, error) {
	_, h, ok := r.Get(name)
	if !ok || h == nil {
		return "", fmt.Errorf("unknown tool %q", name)
	}
	if args == nil {
		args = map[string]any{}
	}
	return h(ctx, userID, args)
}
