package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	freshCache = 2 * time.Minute
	staleCache = 7 * 24 * time.Hour
	maxCache   = 2 << 20
	maxBody    = 8 << 20
)

type desktopServer struct {
	ui       fs.FS
	origin   *url.URL
	cache    *Cache
	sessions *sessionStore
	client   *http.Client
	now      func() time.Time
}

func newDesktopServer(ui fs.FS, origin string, cache *Cache) (*desktopServer, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("api origin %q", origin)
	}
	return &desktopServer{
		ui:     ui,
		origin: u,
		cache:  cache,
		client: &http.Client{Timeout: 25 * time.Second},
		now:    time.Now,
	}, nil
}

func (s *desktopServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/desktop/session", s.localSession)
	mux.HandleFunc("/api/", s.proxyAPI)
	mux.HandleFunc("/", s.serveUI)
	return mux
}

func (s *desktopServer) proxyAPI(w http.ResponseWriter, r *http.Request) {
	pathAndQuery := r.URL.Path
	if r.URL.RawQuery != "" {
		pathAndQuery += "?" + r.URL.RawQuery
	}
	auth := r.Header.Get("Authorization")
	key := cacheKey(auth, pathAndQuery)

	if r.Method == http.MethodGet {
		if entry, ok, err := s.cache.Get(key, freshCache, s.now()); err != nil {
			log.Printf("cache get: %v", err)
		} else if ok {
			writeCached(w, entry)
			return
		}
	}

	resp, err := s.forward(r)
	if r.Method == http.MethodGet && (err != nil || resp.StatusCode >= 500) {
		if entry, ok, cacheErr := s.cache.Get(key, staleCache, s.now()); cacheErr == nil && ok {
			if resp != nil {
				resp.Body.Close()
			}
			writeCached(w, entry)
			return
		}
	}
	if err != nil {
		http.Error(w, "нет связи с сервером", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		http.Error(w, "нет связи с сервером", http.StatusBadGateway)
		return
	}
	if len(body) > maxBody {
		http.Error(w, "ответ слишком большой", http.StatusBadGateway)
		return
	}
	if r.Method == http.MethodGet && resp.StatusCode == http.StatusOK && len(body) <= maxCache {
		ctype := resp.Header.Get("Content-Type")
		if ctype == "" {
			ctype = "application/json"
		}
		if err := s.cache.Put(key, ctype, body, s.now()); err != nil {
			log.Printf("cache put: %v", err)
		}
	}
	if resp.StatusCode == http.StatusOK {
		s.rememberAuth(r, body)
	}
	copyResponse(w, resp, body)
}

func (s *desktopServer) rememberAuth(r *http.Request, body []byte) {
	if s.sessions == nil || r.Method != http.MethodPost {
		return
	}
	switch r.URL.Path {
	case "/api/v1/auth/telegram-login/verify", "/api/v1/auth/telegram-webapp", "/api/v1/auth/token":
	default:
		return
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TelegramID  int64  `json:"telegram_id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.AccessToken == "" {
		return
	}
	exp := s.now().Add(24 * time.Hour)
	if unix, ok := jwtExpiryUnix(payload.AccessToken); ok {
		exp = time.Unix(unix, 0)
	} else if payload.ExpiresIn > 0 {
		exp = s.now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}
	sess := storedSession{
		AccessToken: payload.AccessToken,
		ExpiresAt:   exp.UnixMilli(),
		TelegramID:  payload.TelegramID,
	}
	if !validSession(sess, s.now()) {
		return
	}
	if err := s.sessions.Save(sess); err != nil {
		log.Printf("session save failed")
	}
}

func (s *desktopServer) localSession(w http.ResponseWriter, r *http.Request) {
	if s.sessions == nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		sess, ok := s.sessions.Load(s.now())
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(sess)
	case http.MethodPut:
		var sess storedSession
		if err := json.NewDecoder(io.LimitReader(r.Body, 8192)).Decode(&sess); err != nil || !validSession(sess, s.now()) {
			http.Error(w, "bad session", http.StatusBadRequest)
			return
		}
		if err := s.sessions.Save(sess); err != nil {
			log.Printf("session save failed")
			http.Error(w, "session", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		if err := s.sessions.Clear(); err != nil {
			log.Printf("session clear failed")
			http.Error(w, "session", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		http.Error(w, "method", http.StatusMethodNotAllowed)
	}
}

func (s *desktopServer) forward(r *http.Request) (*http.Response, error) {
	target := s.origin.ResolveReference(&url.URL{Path: r.URL.Path, RawQuery: r.URL.RawQuery})
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header.Clone()
	req.Header.Del("Accept-Encoding")
	req.Host = target.Host
	return s.client.Do(req)
}

func writeCached(w http.ResponseWriter, entry cacheEntry) {
	if entry.contentType != "" {
		w.Header().Set("Content-Type", entry.contentType)
	}
	w.Header().Set("X-LifeOS-Cache", "hit")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(entry.body)
}

func copyResponse(w http.ResponseWriter, resp *http.Response, body []byte) {
	for k, values := range resp.Header {
		if hopByHop(k) || strings.EqualFold(k, "Content-Length") || strings.EqualFold(k, "Content-Encoding") {
			continue
		}
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func hopByHop(header string) bool {
	switch strings.ToLower(header) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func (s *desktopServer) serveUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if name == "" || name == "." {
		name = "index.html"
	}
	body, err := fs.ReadFile(s.ui, name)
	if err != nil {
		if path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		body, err = fs.ReadFile(s.ui, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		name = "index.html"
	}
	ctype := mime.TypeByExtension(path.Ext(name))
	if ctype == "" {
		ctype = "text/html; charset=utf-8"
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	_, _ = io.Copy(w, bytes.NewReader(body))
}
