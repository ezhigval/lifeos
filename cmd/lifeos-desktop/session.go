package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// sessionSkew matches the mini app: a token inside the last minute is already stale.
const sessionSkew = 60 * time.Second

// storedSession is the desktop copy of the browser login.
// The file lives next to the read cache and is not a replay log.
type storedSession struct {
	AccessToken string `json:"accessToken"`
	ExpiresAt   int64  `json:"expiresAt"`
	TelegramID  int64  `json:"telegramId,omitempty"`
}

type sessionStore struct {
	path string
	mu   sync.Mutex
}

func openSessionStore(path string) *sessionStore {
	return &sessionStore{path: path}
}

func (s *sessionStore) Load(now time.Time) (storedSession, bool) {
	if s == nil {
		return storedSession{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return storedSession{}, false
	}
	var sess storedSession
	if err := json.Unmarshal(raw, &sess); err != nil || !validSession(sess, now) {
		return storedSession{}, false
	}
	return sess, true
}

func (s *sessionStore) Save(sess storedSession) error {
	if s == nil {
		return errors.New("no session store")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *sessionStore) Clear() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func validSession(sess storedSession, now time.Time) bool {
	if sess.AccessToken == "" || len(sess.AccessToken) > 4096 {
		return false
	}
	if strings.ContainsAny(sess.AccessToken, " \t\r\n\x00") {
		return false
	}
	if sess.TelegramID < 0 {
		return false
	}
	return sess.ExpiresAt > now.Add(sessionSkew).UnixMilli()
}

func jwtExpiryUnix(token string) (int64, bool) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return 0, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp <= 0 {
		return 0, false
	}
	return claims.Exp, true
}
