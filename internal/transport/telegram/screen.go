package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
)

type Screen struct {
	client   *Client
	sessions *tginfra.Sessions
}

func NewScreen(client *Client, sessions *tginfra.Sessions) *Screen {
	return &Screen{client: client, sessions: sessions}
}

func (s *Screen) Show(
	ctx context.Context,
	userID ids.UserID,
	chatID int64,
	body string,
	extraInline [][]InlineButton,
	replyKB [][]ReplyButton,
) error {
	sess, err := s.sessions.Get(ctx, userID)
	if err != nil {
		return err
	}
	sess.ChatID = chatID

	text := FormatDashboard(body)
	inline := extraInline
	if inline == nil {
		inline = [][]InlineButton{}
	}

	// When reply keyboard must be (re)installed, force a new dashboard message:
	// editMessage* cannot attach ReplyKeyboardMarkup.
	forceNew := len(replyKB) > 0 && !replyKeyboardInstalled(sess.StatePayload)

	if sess.DashboardMessageID > 0 && !forceNew {
		if err := s.client.EditScreen(ctx, chatID, sess.DashboardMessageID, text, inline); err == nil {
			return s.sessions.Save(ctx, sess)
		}
		sess.DashboardMessageID = 0
	}

	// Attach persistent reply keyboard on every NEW dashboard send — no separate
	// "⌨️" carrier. If the user deletes this message, the next action resends and
	// restores the keyboard.
	id, err := s.client.SendScreen(ctx, chatID, text, inline, replyKB)
	if err != nil {
		return err
	}
	sess.DashboardMessageID = id
	if len(replyKB) > 0 {
		if sess.StatePayload == nil {
			sess.StatePayload = map[string]any{}
		}
		sess.StatePayload[replyKBSetKey] = true
		sess.StatePayload[replyKBVersionKey] = float64(replyKBVersion)
	}
	return s.sessions.Save(ctx, sess)
}

func menuActionLabel(action string) (string, bool) {
	switch action {
	case ActionHome:
		return MenuHome, true
	case ActionTasksToday:
		return MenuTasksToday, true
	case ActionPriorities:
		return MenuPriorities, true
	case ActionAddTask:
		return MenuAddTask, true
	case ActionProjectProgress:
		return MenuProjectProgress, true
	case ActionAddProject:
		return MenuAddProject, true
	case ActionTriage:
		return MenuTriage, true
	case ActionHabits:
		return MenuHabits, true
	case ActionProjects:
		return MenuProjects, true
	case ActionCalendar:
		return MenuCalendar, true
	case ActionAnalytics:
		return MenuAnalytics, true
	case ActionSettings:
		return MenuSettings, true
	case ActionMiniApp:
		return MenuMiniApp, true
	default:
		return "", false
	}
}

// TextToAction maps menu label text to action id (exported for tests).
func TextToAction(text string) (string, bool) {
	return textToAction(text)
}

func textToAction(text string) (string, bool) {
	text = strings.TrimSpace(text)
	labels := map[string]string{
		MenuHome:       ActionHome,
		MenuTasksToday: ActionTasksToday,
		MenuProjects:   ActionProjects,
		MenuHabits:     ActionHabits,
		MenuCalendar:   ActionCalendar,
		MenuAnalytics:  ActionAnalytics,
		MenuSettings:   ActionSettings,
		MenuMiniApp:    ActionMiniApp,
		// Старые reply-кнопки (до реорганизации) — на случай если клавиатура не обновилась.
		MenuAddTask:         ActionAddTask,
		MenuPriorities:      ActionPriorities,
		MenuTriage:          ActionTriage,
		MenuAddProject:      ActionAddProject,
		MenuProjectProgress: ActionProjectProgress,
	}
	action, ok := labels[text]
	return action, ok
}

func formatActionName(action string) string {
	if label, ok := menuActionLabel(action); ok {
		return label
	}
	return action
}

func actionError(action string, err error) string {
	return fmt.Sprintf("%s: %v", formatActionName(action), err)
}
