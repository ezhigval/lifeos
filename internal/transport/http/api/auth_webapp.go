package api

import (
	"net/http"
	"time"

	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/platform/auth"
)

type webappAuthRequest struct {
	InitData string `json:"init_data"`
}

func (rt *Router) authTelegramWebApp(w http.ResponseWriter, r *http.Request) {
	if rt.deps.Tokens == nil {
		writeError(w, http.StatusServiceUnavailable, "api auth not configured")
		return
	}
	if rt.deps.BotToken == "" {
		writeError(w, http.StatusServiceUnavailable, "telegram bot token not configured")
		return
	}
	if rt.deps.EnsureUser == nil {
		writeError(w, http.StatusServiceUnavailable, "user bootstrap not configured")
		return
	}

	var req webappAuthRequest
	if err := decodeJSON(r, &req); err != nil || req.InitData == "" {
		writeError(w, http.StatusBadRequest, "init_data is required")
		return
	}

	maxAge := rt.deps.WebAppAuthTTL
	if maxAge <= 0 {
		maxAge = auth.DefaultWebAppAuthTTL
	}
	parsed, err := auth.ValidateWebAppInitData(req.InitData, rt.deps.BotToken, maxAge, time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid init_data")
		return
	}

	user, err := rt.deps.EnsureUser.Execute(r.Context(), identityapp.EnsureUserInput{
		TelegramID:  parsed.User.ID,
		DisplayName: parsed.User.DisplayName(),
	})
	if err != nil {
		if rt.deps.Log != nil {
			rt.deps.Log.Error("ensure user from webapp failed", "error", err, "telegram_id", parsed.User.ID)
		}
		writeError(w, http.StatusInternalServerError, "user resolve failed")
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
	})
}
