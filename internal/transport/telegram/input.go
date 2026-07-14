package telegram

import "strings"

// InputKind classifies inbound chat text before routing.
type InputKind int

const (
	InputCommand InputKind = iota + 1
	InputKeyboard
	InputText
)

// NormalizedInput is the result of classifying a raw Telegram message text.
type NormalizedInput struct {
	Kind    InputKind
	Raw     string
	Command string // set when Kind == InputCommand (e.g. "/start")
	Action  string // set when Kind == InputKeyboard (e.g. ActionHome)
	Text    string // set when Kind == InputText (free-form body)
}

// classifyInput splits user input into command / reply-keyboard / free text.
// Order: bot command prefix -> reply-keyboard label -> everything else is text.
func classifyInput(raw string) NormalizedInput {
	raw = strings.TrimSpace(raw)
	out := NormalizedInput{Raw: raw}
	if raw == "" {
		out.Kind = InputText
		return out
	}

	if cmd := normalizeBotCommand(raw); cmd != "" {
		out.Kind = InputCommand
		out.Command = cmd
		return out
	}

	if action, ok := textToAction(raw); ok {
		out.Kind = InputKeyboard
		out.Action = action
		return out
	}

	out.Kind = InputText
	out.Text = raw
	return out
}

func (k InputKind) String() string {
	switch k {
	case InputCommand:
		return "command"
	case InputKeyboard:
		return "keyboard"
	case InputText:
		return "text"
	default:
		return "unknown"
	}
}

func (in NormalizedInput) IsCommand() bool  { return in.Kind == InputCommand }
func (in NormalizedInput) IsKeyboard() bool { return in.Kind == InputKeyboard }
func (in NormalizedInput) IsText() bool     { return in.Kind == InputText }
