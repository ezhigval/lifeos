package telegram

import (
	"testing"
)

func TestDetectMedia(t *testing.T) {
	t.Parallel()
	msg := &Message{Voice: &Voice{FileID: "v1", Duration: 3, MimeType: "audio/ogg"}}
	src := detectMedia(msg)
	if src.Kind != "voice" || src.FileID != "v1" {
		t.Fatalf("%+v", src)
	}

	msg = &Message{VideoNote: &VideoNote{FileID: "vn1", Duration: 5}}
	src = detectMedia(msg)
	if src.Kind != "video_note" || src.Filename != "circle.mp4" {
		t.Fatalf("%+v", src)
	}

	msg = &Message{
		Caption: "чек",
		Photo:   []PhotoSize{{FileID: "p1", Width: 100}, {FileID: "p2", Width: 800}},
	}
	src = detectMedia(msg)
	if src.Kind != "photo" || src.FileID != "p2" {
		t.Fatalf("want largest photo, got %+v", src)
	}
}

func TestFormatMediaResolveError(t *testing.T) {
	t.Parallel()
	cases := []string{"stt_not_configured", "vision_not_configured", "media_needs_caption", "media_too_long", "media_too_large", "other"}
	for _, c := range cases {
		if got := formatMediaResolveError(errString(c)); got == "" {
			t.Fatalf("empty for %s", c)
		}
	}
}

func TestIsImageMedia(t *testing.T) {
	t.Parallel()
	if !isImageMedia(mediaSource{MimeType: "image/png"}) {
		t.Fatal("mime")
	}
	if !isImageMedia(mediaSource{Filename: "scan.JPEG"}) {
		t.Fatal("ext")
	}
	if isImageMedia(mediaSource{Filename: "doc.pdf", MimeType: "application/pdf"}) {
		t.Fatal("pdf should not be image")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
