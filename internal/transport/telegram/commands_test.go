package telegram

import "testing"

func TestNormalizeBotCommand(t *testing.T) {
	t.Parallel()
	if got := normalizeBotCommand("/start"); got != "/start" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeBotCommand("/start@urban_assist_bot"); got != "/start" {
		t.Fatalf("got %q", got)
	}
}

func TestCommandToAction(t *testing.T) {
	t.Parallel()
	action, ok := commandToAction("/start")
	if !ok || action != ActionHome {
		t.Fatalf("got %q ok=%v", action, ok)
	}
	if _, ok := commandToAction("/clear"); ok {
		t.Fatal("/clear must not map via commandToAction; handled separately")
	}
	if got := normalizeBotCommand("/clear@lifeos_bot"); got != "/clear" {
		t.Fatalf("got %q", got)
	}
	cases := map[string]string{
		"/triage":   ActionTriage,
		"/today":    ActionTasksToday,
		"/tasks":    ActionTasksToday,
		"/habits":   ActionHabits,
		"/projects": ActionProjects,
		"/calendar": ActionCalendar,
	}
	for cmd, want := range cases {
		if got, ok := commandToAction(cmd); !ok || got != want {
			t.Fatalf("%s: got %q ok=%v, want %q", cmd, got, ok, want)
		}
	}
}

func TestDefaultBotCommandsCoverActions(t *testing.T) {
	t.Parallel()
	cmds := DefaultBotCommands()
	if len(cmds) == 0 {
		t.Fatal("empty command list")
	}
	seen := map[string]bool{}
	for _, c := range cmds {
		if seen[c.Command] {
			t.Fatalf("duplicate command %q", c.Command)
		}
		seen[c.Command] = true
		if c.Description == "" {
			t.Fatalf("command %q has empty description", c.Command)
		}
		if c.Command == "triage" {
			continue
		}
	}
	if !seen["triage"] {
		t.Fatal("/triage must be advertised (MA-C6)")
	}
	// Every advertised command except /clear and /cancel must resolve to an action.
	for _, c := range cmds {
		switch c.Command {
		case "clear", "cancel":
			continue
		}
		if _, ok := commandToAction("/" + c.Command); !ok {
			t.Fatalf("advertised command /%s has no action mapping", c.Command)
		}
	}
}

func TestIsCancelText(t *testing.T) {
	t.Parallel()
	if !isCancelText("отмена") || !isCancelText("/cancel") {
		t.Fatal("expected cancel")
	}
}
