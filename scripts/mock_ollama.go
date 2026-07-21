// Mock Ollama /api/chat for local e2e when real llama.cpp segfaults.
// Speaks enough of the Ollama protocol for LifeOS intent + review paths.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]string{{"name": "llama3.2:1b", "model": "llama3.2:1b"}},
		})
	})
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
			Format string `json:"format"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		user := ""
		for _, m := range req.Messages {
			if m.Role == "user" {
				user = m.Content
			}
		}
		content := classify(user, req.Format == "json" || wantsJSON(req.Messages))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": map[string]string{"role": "assistant", "content": content},
		})
	})
	addr := ":11434"
	log.Printf("mock ollama listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func wantsJSON(msgs []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}) bool {
	for _, m := range msgs {
		if m.Role == "system" && strings.Contains(m.Content, "JSON") {
			return true
		}
	}
	return false
}

func classify(user string, asJSON bool) string {
	u := strings.ToLower(strings.TrimSpace(user))
	if !asJSON {
		return "<b>Обзор</b>: сегодня спокойный день, фокус на приоритетах."
	}
	switch {
	case strings.Contains(u, "напом") || strings.Contains(u, "remind"):
		return `{"intent":"reminder.create","title":"","message":"позвонить маме","time_text":"вечером","target":"","unit":"","amount_rubles":null,"hour":null,"minute":null,"confidence":0.9}`
	case strings.Contains(u, "долг") || strings.Contains(u, "debt"):
		return `{"intent":"finance.create_debt","title":"","message":"","time_text":"","target":"банку","unit":"","amount_rubles":100000,"hour":null,"minute":null,"confidence":0.88}`
	case strings.Contains(u, "потрат") || strings.Contains(u, "expense"):
		return `{"intent":"finance.expense","title":"еда","message":"","time_text":"","target":"","unit":"","amount_rubles":500,"hour":null,"minute":null,"confidence":0.9}`
	case strings.Contains(u, "привыч") || strings.Contains(u, "habit"):
		return `{"intent":"habit.create","title":"бег","message":"","time_text":"","target":"","unit":"","amount_rubles":null,"hour":null,"minute":null,"confidence":0.85}`
	default:
		// Phrase that rulebased usually misses — still a valid task.create for e2e.
		return `{"intent":"task.create","title":"разобрать почту","message":"","time_text":"","target":"","unit":"","amount_rubles":null,"hour":null,"minute":null,"confidence":0.82}`
	}
}
