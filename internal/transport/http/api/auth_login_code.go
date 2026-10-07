package api

import (
	"errors"
	"net/http"
	"time"

	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
)

type loginCodeRequest struct {
	Username string `json:"username"`
}

type loginCodeVerifyRequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
}

func (rt *Router) requestTelegramLoginCode(w http.ResponseWriter, r *http.Request) {
	login := rt.deps.TelegramLogin
	if login == nil {
		writeError(w, http.StatusServiceUnavailable, "Вход по коду не настроен")
		return
	}
	var req loginCodeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Введи ник Telegram")
		return
	}
	if err := login.Request(r.Context(), req.Username); err != nil {
		writeLoginCodeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"ttl_seconds": int(5 * time.Minute / time.Second),
	})
}

func (rt *Router) verifyTelegramLoginCode(w http.ResponseWriter, r *http.Request) {
	login := rt.deps.TelegramLogin
	if login == nil || rt.deps.Tokens == nil {
		writeError(w, http.StatusServiceUnavailable, "Вход по коду не настроен")
		return
	}
	var req loginCodeVerifyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Нужны ник и код")
		return
	}
	user, err := login.Verify(r.Context(), req.Username, req.Code)
	if err != nil {
		writeLoginCodeError(w, err)
		return
	}
	token, exp, err := rt.deps.Tokens.Issue(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token issue failed")
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: token,
		ExpiresIn:   int64(timeUntil(exp)),
		TokenType:   "Bearer",
		TelegramID:  user.TelegramID,
	})
}

func writeLoginCodeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identityapp.ErrLoginBadUsername):
		writeError(w, http.StatusBadRequest, "Введи ник Telegram: 5–32 символа, латиница, без ссылки")
	case errors.Is(err, identityapp.ErrLoginUnknownUser):
		writeError(w, http.StatusNotFound, "Не нашёл этот ник. Открой бота и нажми /start, затем повтори")
	case errors.Is(err, identityapp.ErrLoginRateLimited):
		writeError(w, http.StatusTooManyRequests, "Код уже отправлен. Подожди минуту и проверь Telegram")
	case errors.Is(err, identityapp.ErrLoginDelivery):
		writeError(w, http.StatusBadGateway, "Не удалось отправить код. Напиши боту /start и проверь, что он не заблокирован")
	case errors.Is(err, identityapp.ErrLoginBadCode):
		writeError(w, http.StatusUnauthorized, "Неверный или просроченный код")
	default:
		writeError(w, http.StatusInternalServerError, "Не удалось войти")
	}
}
