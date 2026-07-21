// Mock Ollama /api/chat for local e2e when real llama.cpp segfaults.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
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
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		user, system := "", ""
		for _, m := range req.Messages {
			switch m.Role {
			case "user":
				user = m.Content
			case "system":
				system = m.Content
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": map[string]string{"role": "assistant", "content": respond(system, user)},
		})
	})
	log.Printf("mock ollama listening on :11434")
	log.Fatal(http.ListenAndServe(":11434", mux))
}

func respond(system, user string) string {
	msg := extractUserMessage(user)
	u := strings.ToLower(msg)
	agentMode := strings.Contains(system, `"type":"ask"`) ||
		(strings.Contains(system, "ask") && strings.Contains(system, "tool"))

	if strings.Contains(strings.ToLower(user), "tool ") && strings.Contains(strings.ToLower(user), "result=") {
		return agentJSON("reply", "Готово — сделал по вашему запросу.", "", nil)
	}

	if agentMode {
		return agentRespond(msg, u)
	}
	return intentRespond(msg, u)
}

func agentRespond(msg, u string) string {
	switch {
	case strings.Contains(u, "привет") || strings.Contains(u, "как дела") || strings.Contains(u, "что умеешь"):
		return agentJSON("reply", "Привет! Могу задачи, расходы/доходы, напоминания и привычки. Напиши, что нужно.", "", nil)

	case strings.Contains(u, "напом"):
		rem := extractAfter(u, "напомни", "напомнить", "напоминание")
		when := "вечером"
		for _, w := range []string{"вечером", "утром", "через час", "завтра"} {
			if strings.Contains(u, w) {
				when = w
				break
			}
		}
		if rem == "" || rem == u {
			return agentJSON("ask", "Что именно напомнить и на когда?", "", nil)
		}
		return agentJSON("tool", "", "reminder.create", map[string]any{"message": rem, "time_text": when})

	case hasExpense(u):
		amount := extractAmount(u)
		cat := "прочее"
		for _, c := range []string{"еда", "кофе", "транспорт", "такси", "продукты"} {
			if strings.Contains(u, c) {
				cat = c
				break
			}
		}
		if amount <= 0 {
			return agentJSON("ask", "Какую сумму записать в расход?", "", nil)
		}
		return agentJSON("tool", "", "finance.expense", map[string]any{"amount_rubles": amount, "category": cat})

	case hasIncome(u):
		amount := extractAmount(u)
		if amount <= 0 {
			return agentJSON("ask", "Какую сумму дохода записать?", "", nil)
		}
		return agentJSON("tool", "", "finance.income", map[string]any{"amount_rubles": amount, "description": "доход"})

	case (strings.Contains(u, "задач") && (strings.Contains(u, "сегодня") || strings.Contains(u, "список") || strings.Contains(u, "что у меня"))) ||
		strings.Contains(u, "что на сегодня"):
		return agentJSON("tool", "", "task.list_today", map[string]any{})

	case strings.Contains(u, "запомни") || strings.Contains(u, "помни что"):
		val := extractAfter(u, "запомни", "запомни:", "помни что")
		if val == "" {
			return agentJSON("ask", "Что именно запомнить?", "", nil)
		}
		return agentJSON("tool", "", "memory.save", map[string]any{"kind": "fact", "key": "note", "value": val})

	case strings.Contains(u, "привычк") && (strings.Contains(u, "отметь") || strings.Contains(u, "сделал") || strings.Contains(u, "трек")):
		name := extractAfter(u, "привычку", "привычка", "отметь")
		if name == "" {
			return agentJSON("ask", "Какую привычку отметить?", "", nil)
		}
		return agentJSON("tool", "", "habit.track", map[string]any{"name": name})

	case wantsTask(u):
		title := extractTaskTitle(msg, u)
		if title == "" {
			return agentJSON("ask", "Как назвать задачу?", "", nil)
		}
		return agentJSON("tool", "", "task.create", map[string]any{"title": title})

	default:
		// Do NOT invent a placeholder task — ask.
		return agentJSON("ask", "Уточни, пожалуйста: это задача, расход, напоминание или что-то ещё?", "", nil)
	}
}

func intentRespond(msg, u string) string {
	// Classifier path: prefer unknown over fake task.create.
	if wantsTask(u) {
		title := extractTaskTitle(msg, u)
		if title == "" {
			title = strings.TrimSpace(msg)
		}
		return intentJSON("task.create", title, "", "", 0.85)
	}
	if hasExpense(u) {
		return intentJSON("finance.expense", "", "", "", 0.85)
	}
	if strings.Contains(u, "напом") {
		return intentJSON("reminder.create", "", "напоминание", "вечером", 0.85)
	}
	return intentJSON("unknown", "", "", "", 0.2)
}

func wantsTask(u string) bool {
	keys := []string{
		"добавь задачу", "создай задачу", "новая задача", "задачу ",
		"добавь в задачи", "нужно ", "надо ", "сделай ", "купи", "позвон",
		"запланир", "todo",
	}
	for _, k := range keys {
		if strings.Contains(u, k) {
			return true
		}
	}
	return strings.HasPrefix(u, "задача ") || strings.HasPrefix(u, "задача:")
}

func extractTaskTitle(msg, u string) string {
	prefixes := []string{
		"добавь задачу", "создай задачу", "новая задача", "задачу",
		"добавь в задачи", "задача:", "задача ", "нужно", "надо", "сделай",
	}
	title := msg
	lu := u
	for _, p := range prefixes {
		if i := strings.Index(lu, p); i >= 0 {
			title = strings.TrimSpace(msg[i+len(p):])
			title = strings.TrimLeft(title, " :—-")
			break
		}
	}
	title = strings.TrimSpace(title)
	if isPlaceholderTitle(title) {
		return ""
	}
	return title
}

func isPlaceholderTitle(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "", "разобрать входящие", "разобрать почту", "задача", "todo", "new task":
		return true
	default:
		return false
	}
}

func hasExpense(u string) bool {
	for _, k := range []string{"потрат", "расход", "купил", "оплатил", "списал"} {
		if strings.Contains(u, k) {
			return true
		}
	}
	return false
}

func hasIncome(u string) bool {
	for _, k := range []string{"пришёл", "пришел", "доход", "получил", "зарплат", "аванс"} {
		if strings.Contains(u, k) {
			return true
		}
	}
	return false
}

var amountRe = regexp.MustCompile(`(?i)(\d+[.,]?\d*)\s*(тыс|тысяч|к|руб|₽)?`)

func extractAmount(u string) float64 {
	m := amountRe.FindStringSubmatch(u)
	if len(m) < 2 {
		return 0
	}
	raw := strings.ReplaceAll(m[1], ",", ".")
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	if len(m) > 2 {
		switch strings.ToLower(m[2]) {
		case "тыс", "тысяч", "к":
			n *= 1000
		}
	}
	return n
}

func extractAfter(u string, markers ...string) string {
	for _, m := range markers {
		if i := strings.Index(u, m); i >= 0 {
			rest := strings.TrimSpace(u[i+len(m):])
			rest = strings.TrimLeft(rest, " :—-")
			return rest
		}
	}
	return ""
}

func extractUserMessage(blob string) string {
	const marker = "Сообщение пользователя:\n"
	if i := strings.Index(blob, marker); i >= 0 {
		return strings.TrimSpace(blob[i+len(marker):])
	}
	// History form: last "user: ..." line
	lines := strings.Split(blob, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "user: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "user: "))
		}
	}
	return strings.TrimSpace(blob)
}

func agentJSON(typ, text, tool string, args map[string]any) string {
	if args == nil {
		args = map[string]any{}
	}
	b, _ := json.Marshal(map[string]any{
		"type": typ, "text": text, "tool": tool, "args": args,
	})
	return string(b)
}

func intentJSON(intent, title, message, timeText string, conf float64) string {
	b, _ := json.Marshal(map[string]any{
		"intent": intent, "title": title, "message": message, "time_text": timeText,
		"target": "", "unit": "", "amount_rubles": nil, "hour": nil, "minute": nil, "confidence": conf,
	})
	return string(b)
}
