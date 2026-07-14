package telegram

import "testing"

func TestClassifyInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw     string
		kind    InputKind
		command string
		action  string
		text    string
	}{
		{"/start", InputCommand, CmdStart, "", ""},
		{"/start@lifeos_bot", InputCommand, CmdStart, "", ""},
		{"/clear", InputCommand, CmdClear, "", ""},
		{"/delete @alice", InputCommand, CmdDelete, "", ""},
		{MenuHome, InputKeyboard, "", ActionHome, ""},
		{MenuTasksToday, InputKeyboard, "", ActionTasksToday, ""},
		{"купить молоко", InputText, "", "", "купить молоко"},
		{"  отмена  ", InputText, "", "", "отмена"},
		{"confirm", InputText, "", "", "confirm"},
	}

	for _, tc := range cases {
		got := classifyInput(tc.raw)
		if got.Kind != tc.kind {
			t.Fatalf("%q: kind=%s want=%s", tc.raw, got.Kind, tc.kind)
		}
		if got.Command != tc.command {
			t.Fatalf("%q: command=%q want=%q", tc.raw, got.Command, tc.command)
		}
		if got.Action != tc.action {
			t.Fatalf("%q: action=%q want=%q", tc.raw, got.Action, tc.action)
		}
		if got.Text != tc.text {
			t.Fatalf("%q: text=%q want=%q", tc.raw, got.Text, tc.text)
		}
		if got.IsCommand() != (tc.kind == InputCommand) ||
			got.IsKeyboard() != (tc.kind == InputKeyboard) ||
			got.IsText() != (tc.kind == InputText) {
			t.Fatalf("%q: predicate mismatch for kind %s", tc.raw, tc.kind)
		}
	}
}
