// Mock Ollama /api/chat for local e2e when real llama.cpp segfaults.
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
		system := ""
		for _, m := range req.Messages {
			switch m.Role {
			case "user":
				user = m.Content
			case "system":
				system = m.Content
			}
		}
		content := respond(system, user)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": map[string]string{"role": "assistant", "content": content},
		})
	})
	addr := ":11434"
	log.Printf("mock ollama listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func respond(system, user string) string {
	u := strings.ToLower(user)
	agentMode := strings.Contains(system, `"type"`) || strings.Contains(system, "ask") && strings.Contains(system, "tool")

	// If history already contains a tool result, finish with reply.
	if strings.Contains(u, "tool ") && strings.Contains(u, "result=") {
		return `{"type":"reply","text":"Готово — сделал по вашему запросу.","tool":"","args":{}}`
	}

	if agentMode {
		switch {
		case strings.Contains(u, "напом") || strings.Contains(u, "remind"):
			if !strings.Contains(u, "позвон") && !strings.Contains(u, "message") {
				return `{"type":"ask","text":"Что именно напомнить и на когда?","tool":"","args":{}}`
			}
			return `{"type":"tool","text":"","tool":"reminder.create","args":{"message":"позвонить","time_text":"вечером"}}`
		case strings.Contains(u, "потрат") || strings.Contains(u, "expense") || strings.Contains(u, "руб"):
			return `{"type":"tool","text":"","tool":"finance.expense","args":{"amount_rubles":500,"category":"еда"}}`
		case strings.Contains(u, "задач") && (strings.Contains(u, "сегодня") || strings.Contains(u, "список")):
			return `{"type":"tool","text":"","tool":"task.list_today","args":{}}`
		case strings.Contains(u, "запомни") || strings.Contains(u, "memory"):
			return `{"type":"tool","text":"","tool":"memory.save","args":{"kind":"preference","key":"coffee","value":"без сахара"}}`
		case strings.Contains(u, "привет") || strings.Contains(u, "как дела"):
			return `{"type":"reply","text":"Привет! Я LifeOS — могу задачи, финансы, напоминания и привычки. Что сделаем?","tool":"","args":{}}`
		default:
			return `{"type":"tool","text":"","tool":"task.create","args":{"title":"разобрать входящие"}}`
		}
	}

	// Legacy intent classifier JSON
	switch {
	case strings.Contains(u, "напом") || strings.Contains(u, "remind"):
		return `{"intent":"reminder.create","title":"","message":"позвонить маме","time_text":"вечером","target":"","unit":"","amount_rubles":null,"hour":null,"minute":null,"confidence":0.9}`
	default:
		return `{"intent":"task.create","title":"разобрать почту","message":"","time_text":"","target":"","unit":"","amount_rubles":null,"hour":null,"minute":null,"confidence":0.82}`
	}
}
