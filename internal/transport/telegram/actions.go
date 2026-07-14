package telegram

import (
	"context"
	"html"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
)

func (h *MessageHandler) runAction(ctx context.Context, userID ids.UserID, action string) (dispatchResult, error) {
	switch action {
	case ActionAddTask:
		if err := h.sessions.SetState(ctx, userID, tginfra.StateAwaitTaskTitle, h.basePayload(ctx, userID)); err != nil {
			return dispatchResult{}, err
		}
	case ActionAddProject:
		if err := h.sessions.SetState(ctx, userID, tginfra.StateAwaitProjectTitle, h.basePayload(ctx, userID)); err != nil {
			return dispatchResult{}, err
		}
	default:
		// basePayload drops pending_delete_* / draft_* so reply-keyboard navigation
		// cannot leave an account wipe still armed after the user walks away.
		if err := h.sessions.SetState(ctx, userID, tginfra.StateIdle, h.basePayload(ctx, userID)); err != nil {
			return dispatchResult{}, err
		}
	}

	switch action {
	case ActionHome:
		items, err := h.listToday.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatHomeSummary(len(items)), inline: InlineHomeActions()}, nil
	case ActionTasksToday:
		return h.tasksTodayView(ctx, userID)
	case ActionPriorities:
		items, err := h.priorities.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatPriorities(items)}, nil
	case ActionAddTask:
		return dispatchResult{text: FormatPromptTaskTitle()}, nil
	case ActionProjectProgress:
		p, err := h.projectProg.Execute(ctx, userID, ids.ProjectID{})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatProjectProgress(p)}, nil
	case ActionAddProject:
		return dispatchResult{text: FormatPromptProjectTitle()}, nil
	case ActionTriage:
		msg, low, err := h.triage.Propose(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		out := dispatchResult{text: msg, deferTasks: low}
		if len(low) > 0 {
			out.inline = [][]InlineButton{{{Text: "Перенести low", CallbackData: CBTriageDefer}}}
		}
		return out, nil
	case ActionHabits:
		return h.habitsTodayView(ctx, userID)
	case ActionProjects:
		return h.projectsPickerView(ctx, userID)
	case ActionCalendar:
		return h.calendarTodayView(ctx, userID)
	case ActionAnalytics:
		return h.analyticsView(ctx, userID)
	case ActionSettings:
		return h.settingsView(ctx, userID)
	case ActionMiniApp:
		if h.miniAppURL == "" {
			return dispatchResult{text: "Mini App ещё не настроен (нет LIFEOS_MINIAPP_URL)."}, nil
		}
		return dispatchResult{text: "📱 Mini App: " + html.EscapeString(h.miniAppURL) + "\nНажми кнопку <b>📱 Mini App</b> на клавиатуре, чтобы открыть."}, nil
	default:
		return dispatchResult{text: FormatFallback()}, nil
	}
}
