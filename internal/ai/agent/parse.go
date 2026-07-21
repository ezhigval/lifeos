package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

type actionType string

const (
	actionAsk   actionType = "ask"
	actionReply actionType = "reply"
	actionTool  actionType = "tool"
)

type modelAction struct {
	Type actionType      `json:"type"`
	Text string          `json:"text"`
	Tool string          `json:"tool"`
	Args map[string]any  `json:"args"`
}

func parseAction(raw string) (modelAction, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return modelAction{}, fmt.Errorf("empty model response")
	}

	// Strip optional markdown fences.
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSpace(raw)
		if strings.HasPrefix(strings.ToLower(raw), "json") {
			raw = strings.TrimSpace(raw[4:])
		}
		if i := strings.LastIndex(raw, "```"); i >= 0 {
			raw = strings.TrimSpace(raw[:i])
		}
	}

	var a modelAction
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		// Try to extract a JSON object substring.
		start := strings.Index(raw, "{")
		end := strings.LastIndex(raw, "}")
		if start >= 0 && end > start {
			if err2 := json.Unmarshal([]byte(raw[start:end+1]), &a); err2 != nil {
				return modelAction{}, fmt.Errorf("parse action json: %w", err)
			}
		} else {
			return modelAction{}, fmt.Errorf("parse action json: %w", err)
		}
	}

	a.Type = actionType(strings.ToLower(strings.TrimSpace(string(a.Type))))
	a.Text = strings.TrimSpace(a.Text)
	a.Tool = strings.TrimSpace(a.Tool)
	switch a.Type {
	case actionAsk, actionReply:
		if a.Text == "" {
			return modelAction{}, fmt.Errorf("action %s requires text", a.Type)
		}
	case actionTool:
		if a.Tool == "" {
			return modelAction{}, fmt.Errorf("action tool requires tool name")
		}
		if a.Args == nil {
			a.Args = map[string]any{}
		}
	default:
		return modelAction{}, fmt.Errorf("unknown action type %q", a.Type)
	}
	return a, nil
}
