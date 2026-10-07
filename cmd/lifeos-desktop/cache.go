package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type cacheEntry struct {
	body        []byte
	contentType string
	storedAt    time.Time
}

// Cache stores successful GET responses on disk. The key is a hash, so the
// bearer token is not written into the database.
type Cache struct {
	db *sql.DB
}

func OpenCache(path string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("cache dir: %w", err)
	}
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: "_pragma=busy_timeout(5000)&mode=rwc",
	}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open cache: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS http_cache (
			cache_key TEXT PRIMARY KEY,
			body BLOB NOT NULL,
			content_type TEXT NOT NULL,
			stored_at INTEGER NOT NULL
		)`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("cache schema: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = db.Close()
		return nil, fmt.Errorf("cache mode: %w", err)
	}
	return &Cache{db: db}, nil
}

func (c *Cache) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}

func cacheKey(authorization, pathAndQuery string) string {
	sum := sha256.Sum256([]byte(authorization + "\n" + pathAndQuery))
	return hex.EncodeToString(sum[:])
}

func (c *Cache) Put(key, contentType string, body []byte, at time.Time) error {
	_, err := c.db.Exec(
		`INSERT INTO http_cache (cache_key, body, content_type, stored_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(cache_key) DO UPDATE SET
		   body = excluded.body,
		   content_type = excluded.content_type,
		   stored_at = excluded.stored_at`,
		key, body, contentType, at.Unix(),
	)
	return err
}

func (c *Cache) Get(key string, maxAge time.Duration, now time.Time) (cacheEntry, bool, error) {
	var body []byte
	var contentType string
	var stored int64
	err := c.db.QueryRow(
		`SELECT body, content_type, stored_at FROM http_cache WHERE cache_key = ?`,
		key,
	).Scan(&body, &contentType, &stored)
	if errors.Is(err, sql.ErrNoRows) {
		return cacheEntry{}, false, nil
	}
	if err != nil {
		return cacheEntry{}, false, err
	}
	at := time.Unix(stored, 0)
	if now.Sub(at) > maxAge {
		return cacheEntry{}, false, nil
	}
	return cacheEntry{body: body, contentType: contentType, storedAt: at}, true, nil
}
