package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const webAppDataKey = "WebAppData"

// DefaultWebAppAuthTTL rejects initData older than this when MaxAge is unset.
const DefaultWebAppAuthTTL = 24 * time.Hour

// WebAppUser is the subset of Telegram.WebAppUser used for identity.
type WebAppUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsPremium    bool   `json:"is_premium"`
}

// WebAppInitData is a validated Telegram Mini App initData payload.
type WebAppInitData struct {
	User      WebAppUser
	AuthDate  time.Time
	QueryID   string
	RawFields map[string]string
}

// ValidateWebAppInitData verifies Telegram WebApp initData per
// https://core.telegram.org/bots/webapps#validating-data-received-via-the-web-app
//
// secret_key = HMAC_SHA256(bot_token, key=WebAppData)
// hash       = hex(HMAC_SHA256(data_check_string, key=secret_key))
func ValidateWebAppInitData(initData, botToken string, maxAge time.Duration, now time.Time) (WebAppInitData, error) {
	if strings.TrimSpace(initData) == "" {
		return WebAppInitData{}, fmt.Errorf("init_data is required")
	}
	if strings.TrimSpace(botToken) == "" {
		return WebAppInitData{}, fmt.Errorf("bot token is required")
	}
	if maxAge <= 0 {
		maxAge = DefaultWebAppAuthTTL
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return WebAppInitData{}, fmt.Errorf("parse init_data: %w", err)
	}
	hash := values.Get("hash")
	if hash == "" {
		return WebAppInitData{}, fmt.Errorf("missing hash")
	}

	pairs := make([]string, 0, len(values))
	fields := make(map[string]string, len(values))
	for key, vals := range values {
		if key == "hash" || key == "signature" {
			continue
		}
		if len(vals) == 0 {
			continue
		}
		// ParseQuery already URL-decodes values.
		fields[key] = vals[0]
		pairs = append(pairs, key+"="+vals[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secret := hmacSHA256([]byte(webAppDataKey), []byte(botToken))
	computed := hex.EncodeToString(hmacSHA256(secret, []byte(dataCheckString)))
	if !hmac.Equal([]byte(computed), []byte(hash)) {
		return WebAppInitData{}, fmt.Errorf("invalid init_data hash")
	}

	authRaw, ok := fields["auth_date"]
	if !ok || authRaw == "" {
		return WebAppInitData{}, fmt.Errorf("missing auth_date")
	}
	authUnix, err := strconv.ParseInt(authRaw, 10, 64)
	if err != nil || authUnix <= 0 {
		return WebAppInitData{}, fmt.Errorf("invalid auth_date")
	}
	authDate := time.Unix(authUnix, 0).UTC()
	if now.Sub(authDate) > maxAge {
		return WebAppInitData{}, fmt.Errorf("init_data expired")
	}
	if authDate.After(now.Add(5 * time.Minute)) {
		return WebAppInitData{}, fmt.Errorf("auth_date is in the future")
	}

	userRaw, ok := fields["user"]
	if !ok || userRaw == "" {
		return WebAppInitData{}, fmt.Errorf("missing user")
	}
	var user WebAppUser
	if err := json.Unmarshal([]byte(userRaw), &user); err != nil {
		return WebAppInitData{}, fmt.Errorf("invalid user json: %w", err)
	}
	if user.ID <= 0 {
		return WebAppInitData{}, fmt.Errorf("invalid user id")
	}

	return WebAppInitData{
		User:      user,
		AuthDate:  authDate,
		QueryID:   fields["query_id"],
		RawFields: fields,
	}, nil
}

// DisplayName builds a stable display name for EnsureUser.
func (u WebAppUser) DisplayName() string {
	switch {
	case u.FirstName != "" && u.LastName != "":
		return u.FirstName + " " + u.LastName
	case u.FirstName != "":
		return u.FirstName
	case u.Username != "":
		return u.Username
	default:
		return fmt.Sprintf("User %d", u.ID)
	}
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

// SignWebAppInitData builds a valid initData string for tests.
func SignWebAppInitData(fields map[string]string, botToken string) string {
	pairs := make([]string, 0, len(fields))
	q := url.Values{}
	for k, v := range fields {
		if k == "hash" || k == "signature" {
			continue
		}
		pairs = append(pairs, k+"="+v)
		q.Set(k, v)
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")
	secret := hmacSHA256([]byte(webAppDataKey), []byte(botToken))
	hash := hex.EncodeToString(hmacSHA256(secret, []byte(dataCheckString)))
	q.Set("hash", hash)
	return q.Encode()
}
