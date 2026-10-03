package telegram

import (
	"context"
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
	miniAppURL string,
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
	wantURL := strings.TrimSpace(miniAppURL)
	wantMiniApp := replyKeyboardHasMiniApp(replyKB)
	forceNew := len(replyKB) > 0 && !replyKeyboardInstalled(sess.StatePayload, wantMiniApp, wantURL)

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
		sess.StatePayload[replyKBMiniAppKey] = wantMiniApp
		if wantMiniApp {
			sess.StatePayload[replyKBMiniAppURLKey] = wantURL
		} else {
			delete(sess.StatePayload, replyKBMiniAppURLKey)
		}
	}
	return s.sessions.Save(ctx, sess)
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
