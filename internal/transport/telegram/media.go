package telegram

import (
	"context"
	"fmt"
	"path"
	"strings"
)

const (
	maxVoiceBytes     = 20 << 20 // 20 MiB Telegram bot download soft cap
	maxVoiceDurationS = 120
)

type mediaSource struct {
	Kind     string // voice | audio | video_note | video | photo | document | text
	FileID   string
	Filename string
	Duration int
	MimeType string
	FileSize int64
}

// resolveIncomingText turns a Telegram message into text for the same free-text /
// agent path. Voice, audio, and video notes go through SpeechToText when configured.
// Photos/documents without caption ask for a caption (vision deferred).
func (h *MessageHandler) resolveIncomingText(ctx context.Context, chatID int64, msg *Message) (string, error) {
	if msg == nil {
		return "", nil
	}
	if text := strings.TrimSpace(msg.Text); text != "" {
		return text, nil
	}
	caption := strings.TrimSpace(msg.Caption)
	src := detectMedia(msg)
	if src.Kind == "" {
		return "", nil
	}

	switch src.Kind {
	case "photo", "document":
		if caption != "" {
			return caption, nil
		}
		return "", fmt.Errorf("media_needs_caption")
	case "voice", "audio", "video_note", "video":
		if caption != "" && h.stt == nil {
			return caption, nil
		}
		if h.stt == nil {
			return "", fmt.Errorf("stt_not_configured")
		}
		if src.Duration > maxVoiceDurationS {
			return "", fmt.Errorf("media_too_long")
		}
		if src.FileSize > 0 && src.FileSize > maxVoiceBytes {
			return "", fmt.Errorf("media_too_large")
		}
		_ = h.client.SendChatAction(ctx, chatID, "typing")
		text, err := h.transcribeMedia(ctx, src)
		if err != nil {
			return "", err
		}
		if caption != "" {
			return strings.TrimSpace(caption + "\n" + text), nil
		}
		return text, nil
	default:
		return caption, nil
	}
}

func detectMedia(msg *Message) mediaSource {
	switch {
	case msg.Voice != nil && msg.Voice.FileID != "":
		return mediaSource{
			Kind: "voice", FileID: msg.Voice.FileID, Filename: "voice.ogg",
			Duration: msg.Voice.Duration, MimeType: msg.Voice.MimeType, FileSize: msg.Voice.FileSize,
		}
	case msg.VideoNote != nil && msg.VideoNote.FileID != "":
		return mediaSource{
			Kind: "video_note", FileID: msg.VideoNote.FileID, Filename: "circle.mp4",
			Duration: msg.VideoNote.Duration, FileSize: msg.VideoNote.FileSize, MimeType: "video/mp4",
		}
	case msg.Audio != nil && msg.Audio.FileID != "":
		name := msg.Audio.FileName
		if name == "" {
			name = "audio.ogg"
		}
		return mediaSource{
			Kind: "audio", FileID: msg.Audio.FileID, Filename: name,
			Duration: msg.Audio.Duration, MimeType: msg.Audio.MimeType, FileSize: msg.Audio.FileSize,
		}
	case msg.Video != nil && msg.Video.FileID != "":
		name := msg.Video.FileName
		if name == "" {
			name = "video.mp4"
		}
		return mediaSource{
			Kind: "video", FileID: msg.Video.FileID, Filename: name,
			Duration: msg.Video.Duration, MimeType: msg.Video.MimeType, FileSize: msg.Video.FileSize,
		}
	case len(msg.Photo) > 0:
		best := msg.Photo[len(msg.Photo)-1]
		return mediaSource{Kind: "photo", FileID: best.FileID, Filename: "photo.jpg", FileSize: best.FileSize}
	case msg.Document != nil && msg.Document.FileID != "":
		name := msg.Document.FileName
		if name == "" {
			name = "file.bin"
		}
		return mediaSource{
			Kind: "document", FileID: msg.Document.FileID, Filename: name,
			MimeType: msg.Document.MimeType, FileSize: msg.Document.FileSize,
		}
	default:
		return mediaSource{}
	}
}

func (h *MessageHandler) transcribeMedia(ctx context.Context, src mediaSource) (string, error) {
	file, err := h.client.GetFile(ctx, src.FileID)
	if err != nil {
		return "", fmt.Errorf("get file: %w", err)
	}
	data, err := h.client.DownloadFile(ctx, file.FilePath, maxVoiceBytes)
	if err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}
	filename := src.Filename
	if ext := path.Ext(file.FilePath); ext != "" && path.Ext(filename) == "" {
		filename += ext
	}
	text, err := h.stt.Transcribe(ctx, data, filename, "ru")
	if err != nil {
		return "", fmt.Errorf("transcribe: %w", err)
	}
	return strings.TrimSpace(text), nil
}

func formatMediaResolveError(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case strings.Contains(err.Error(), "stt_not_configured"):
		return "Голосовые и кружочки пока недоступны: включи STT (LIFEOS_STT_ENABLED + ключ Whisper/Groq) или напиши текстом."
	case strings.Contains(err.Error(), "media_needs_caption"):
		return "Пока читаю только подпись к фото/файлу. Добавь caption или опиши текстом, что сделать."
	case strings.Contains(err.Error(), "media_too_long"):
		return "Слишком длинное аудио (лимит ~2 мин). Сократи или напиши текстом."
	case strings.Contains(err.Error(), "media_too_large"):
		return "Файл слишком большой для распознавания. Сократи запись или напиши текстом."
	default:
		return "Не смог распознать голос/медиа. Попробуй ещё раз или напиши текстом."
	}
}
