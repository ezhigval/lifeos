package telegram

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"

	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
)

const (
	CBDeleteOK     = "delete:ok"
	CBDeleteCancel = "delete:cancel"

	PayloadPendingDeleteTG   = "pending_delete_tg"
	PayloadPendingDeleteName = "pending_delete_name"
)

func parseDeleteTarget(text string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) < 2 {
		return ""
	}
	return strings.TrimPrefix(fields[1], "@")
}

func isDeleteConfirmText(text string) bool {
	t := strings.TrimSpace(strings.ToLower(text))
	switch t {
	case "confirm", "подтвердить", "удалить", "yes", "да":
		return true
	default:
		return false
	}
}

func (h *MessageHandler) beginDeleteUser(ctx context.Context, actor ids.UserID, actorTG int64, actorUsername, rawText string) (dispatchResult, error) {
	targetRaw := parseDeleteTarget(rawText)
	if targetRaw == "" {
		return dispatchResult{text: "Использование: <code>/delete @username</code>\nЗатем подтверди кнопкой или напиши <code>confirm</code>."}, nil
	}

	targetTG, label, err := h.resolveDeleteTarget(ctx, targetRaw, actorTG, actorUsername)
	if err != nil {
		return dispatchResult{text: "Не удалось найти пользователя: " + html.EscapeString(err.Error())}, nil
	}
	if !h.canDeleteUser(actorTG, targetTG) {
		return dispatchResult{text: "Недостаточно прав: можно удалить только свой аккаунт (чужой — только admin seed)."}, nil
	}

	payload := h.basePayload(ctx, actor)
	payload[PayloadPendingDeleteTG] = float64(targetTG)
	payload[PayloadPendingDeleteName] = label
	if err := h.sessions.SetState(ctx, actor, tginfra.StateIdle, payload); err != nil {
		return dispatchResult{}, err
	}

	text := fmt.Sprintf(
		"⚠️ <b>Удаление аккаунта</b>\nЦель: <b>%s</b> (tg:%d)\n\n"+
			"Будут удалены <b>все данные</b> пользователя из БД (задачи, привычки, проекты…).\n"+
			"Переписка будет сброшена.\n\n"+
			"Подтверди кнопкой ниже или напиши <code>confirm</code>.",
		html.EscapeString(label), targetTG,
	)
	inline := [][]InlineButton{{
		{Text: "✅ Удалить навсегда", CallbackData: CBDeleteOK},
		{Text: "❌ Отмена", CallbackData: CBDeleteCancel},
	}}
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) resolveDeleteTarget(ctx context.Context, targetRaw string, actorTG int64, actorUsername string) (int64, string, error) {
	lower := strings.ToLower(strings.TrimSpace(targetRaw))
	if lower == "me" || lower == "self" || lower == "я" {
		label := actorUsername
		if label == "" {
			label = fmt.Sprintf("tg:%d", actorTG)
		} else {
			label = "@" + label
		}
		return actorTG, label, nil
	}
	if id, err := strconv.ParseInt(targetRaw, 10, 64); err == nil && id > 0 {
		return id, fmt.Sprintf("tg:%d", id), nil
	}
	if actorUsername != "" && strings.EqualFold(actorUsername, targetRaw) {
		return actorTG, "@" + actorUsername, nil
	}
	tgID, err := h.client.ResolveUsername(ctx, targetRaw)
	if err != nil {
		return 0, "", err
	}
	return tgID, "@" + targetRaw, nil
}

func (h *MessageHandler) canDeleteUser(actorTG, targetTG int64) bool {
	if actorTG == targetTG {
		return true
	}
	return h.adminTelegramID > 0 && actorTG == h.adminTelegramID
}

func (h *MessageHandler) cancelPendingDelete(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	payload := h.basePayload(ctx, userID)
	delete(payload, PayloadPendingDeleteTG)
	delete(payload, PayloadPendingDeleteName)
	_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, payload)
	out, err := h.runAction(ctx, userID, ActionHome)
	if err != nil {
		return dispatchResult{}, err
	}
	out.text = "Удаление отменено.\n\n" + out.text
	return out, nil
}

func (h *MessageHandler) confirmPendingDelete(ctx context.Context, actor ids.UserID, actorTG, chatID, lastMsgID int64) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, actor)
	if err != nil {
		return dispatchResult{}, err
	}
	targetTG, ok := payloadInt64(sess.StatePayload, PayloadPendingDeleteTG)
	if !ok || targetTG <= 0 {
		return dispatchResult{text: "Нет ожидающего удаления. Сначала: <code>/delete @username</code>"}, nil
	}
	label, _ := sess.StatePayload[PayloadPendingDeleteName].(string)
	if label == "" {
		label = fmt.Sprintf("tg:%d", targetTG)
	}
	if !h.canDeleteUser(actorTG, targetTG) {
		return dispatchResult{text: "Недостаточно прав для удаления."}, nil
	}
	if h.deleteUser == nil {
		return dispatchResult{text: "Удаление пользователя не настроено на сервере."}, nil
	}

	targetChatID := chatID
	high := lastMsgID
	if targetTG != actorTG {
		if targetUser, lerr := h.deleteUser.Lookup(ctx, targetTG); lerr == nil {
			if ts, serr := h.sessions.Get(ctx, targetUser.ID); serr == nil {
				if ts.ChatID != 0 {
					targetChatID = ts.ChatID
				}
				if ts.DashboardMessageID > high {
					high = ts.DashboardMessageID
				}
			}
		}
	} else if sess.DashboardMessageID > high {
		high = sess.DashboardMessageID
	}
	if high < 1 {
		high = 1
	}
	low := high - clearChatMessageWindow + 1
	if low < 1 {
		low = 1
	}
	if targetChatID != 0 {
		h.deleteMessageRange(ctx, targetChatID, low, high)
	}

	deleted, err := h.deleteUser.Execute(ctx, targetTG)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return dispatchResult{text: fmt.Sprintf("Пользователь %s не найден в БД.", html.EscapeString(label))}, nil
		}
		return dispatchResult{}, err
	}

	if targetTG == actorTG {
		fresh, err := h.ensureUser.Execute(ctx, identityapp.EnsureUserInput{
			TelegramID:  actorTG,
			DisplayName: deleted.DisplayName,
		})
		if err != nil {
			return dispatchResult{}, fmt.Errorf("recreate user: %w", err)
		}
		_ = h.sessions.Reset(ctx, fresh.ID, chatID)
		out, err := h.runAction(ctx, fresh.ID, ActionHome)
		if err != nil {
			return dispatchResult{}, err
		}
		out.text = "🗑 Аккаунт удалён. Данные сброшены — можно начинать с нуля.\n\n" + out.text
		return out, nil
	}

	payload := h.basePayload(ctx, actor)
	delete(payload, PayloadPendingDeleteTG)
	delete(payload, PayloadPendingDeleteName)
	_ = h.sessions.SetState(ctx, actor, tginfra.StateIdle, payload)
	out, err := h.runAction(ctx, actor, ActionHome)
	if err != nil {
		return dispatchResult{}, err
	}
	out.text = fmt.Sprintf("🗑 Пользователь %s (tg:%d) удалён из БД, переписка сброшена.\n\n%s",
		html.EscapeString(label), targetTG, out.text)
	return out, nil
}

func payloadInt64(payload map[string]any, key string) (int64, bool) {
	if payload == nil {
		return 0, false
	}
	v, ok := payload[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case string:
		parsed, err := strconv.ParseInt(n, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
