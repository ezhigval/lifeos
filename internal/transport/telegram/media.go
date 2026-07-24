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
	maxImageBytes     = 10 << 20 // 10 MiB for vision
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
// agent path. Voice/audio/video notes go through SpeechToText; photos (and image
// documents) without caption go through Vision when configured.
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
	case "photo":
		if caption != "" {
			return caption, nil
		}
		return h.withAck(ctx, chatID, "Смотрю фото…", func() (string, error) {
			return h.visionMedia(ctx, chatID, src)
		})
	case "document":
		if caption != "" {
			return caption, nil
		}
		if !isImageMedia(src) {
			return "", fmt.Errorf("media_needs_caption")
		}
		return h.withAck(ctx, chatID, "Смотрю фото…", func() (string, error) {
			return h.visionMedia(ctx, chatID, src)
		})
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
		return h.withAck(ctx, chatID, "Слушаю…", func() (string, error) {
			_ = h.client.SendChatAction(ctx, chatID, "typing")
			text, err := h.transcribeMedia(ctx, src)
			if err != nil {
				return "", err
			}
			if caption != "" {
				return strings.TrimSpace(caption + "\n" + text), nil
			}
			return text, nil
		})
	default:
		return caption, nil
	}
}

func (h *MessageHandler) withAck(ctx context.Context, chatID int64, ack string, fn func() (string, error)) (string, error) {
	var ackID int64
	if h.client != nil && strings.TrimSpace(ack) != "" {
		id, err := h.client.SendPlainMessage(ctx, chatID, ack)
		if err == nil {
			ackID = id
		}
	}
	defer func() {
		if ackID > 0 && h.client != nil {
			_ = h.client.DeleteMessage(ctx, chatID, ackID)
		}
	}()
	return fn()
}

func (h *MessageHandler) visionMedia(ctx context.Context, chatID int64, src mediaSource) (string, error) {
	if h.vision == nil {
		return "", fmt.Errorf("vision_not_configured")
	}
	if src.FileSize > 0 && src.FileSize > maxImageBytes {
		return "", fmt.Errorf("media_too_large")
	}
	_ = h.client.SendChatAction(ctx, chatID, "typing")
	file, err := h.client.GetFile(ctx, src.FileID)
	if err != nil {
		return "", fmt.Errorf("get file: %w", err)
	}
	data, err := h.client.DownloadFile(ctx, file.FilePath, maxImageBytes)
	if err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}
	mime := src.MimeType
	if mime == "" {
		mime = mimeFromFilename(src.Filename, file.FilePath)
	}
	text, err := h.vision.ImageToUserText(ctx, data, mime)
	if err != nil {
		return "", fmt.Errorf("vision: %w", err)
	}
	return strings.TrimSpace(text), nil
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
		return mediaSource{Kind: "photo", FileID: best.FileID, Filename: "photo.jpg", FileSize: best.FileSize, MimeType: "image/jpeg"}
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

func isImageMedia(src mediaSource) bool {
	mt := strings.ToLower(strings.TrimSpace(src.MimeType))
	if strings.HasPrefix(mt, "image/") {
		return true
	}
	ext := strings.ToLower(path.Ext(src.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func mimeFromFilename(names ...string) string {
	for _, name := range names {
		switch strings.ToLower(path.Ext(name)) {
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".webp":
			return "image/webp"
		case ".jpg", ".jpeg":
			return "image/jpeg"
		}
	}
	return "image/jpeg"
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
	case strings.Contains(err.Error(), "vision_not_configured"):
		return "Фото без подписи пока не читаю: включи Vision (LIFEOS_VISION_ENABLED + ключ) или добавь caption."
	case strings.Contains(err.Error(), "media_needs_caption"):
		return "К этому файлу нужна подпись (caption) или опиши текстом, что сделать."
	case strings.Contains(err.Error(), "media_too_long"):
		return "Слишком длинное аудио (лимит ~2 мин). Сократи или напиши текстом."
	case strings.Contains(err.Error(), "media_too_large"):
		return "Файл слишком большой для распознавания. Сократи запись или напиши текстом."
	default:
		return "Не смог распознать голос/медиа. Попробуй ещё раз или напиши текстом."
	}
}
