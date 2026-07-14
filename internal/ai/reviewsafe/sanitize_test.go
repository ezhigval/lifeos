package reviewsafe

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSanitizeHTMLKeepsBold(t *testing.T) {
	t.Parallel()
	got := SanitizeHTML(`Focus on <b>топ-3</b> today`)
	want := `Focus on <b>топ-3</b> today`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSanitizeHTMLEscapesUnsafeTags(t *testing.T) {
	t.Parallel()
	got := SanitizeHTML(`Hi <script>alert(1)</script> and <a href="x">link</a>`)
	if got == "" {
		t.Fatal("empty")
	}
	for _, bad := range []string{"<script", "<a ", "</a>"} {
		if strings.Contains(got, bad) {
			t.Fatalf("unsafe tag leaked (%s): %q", bad, got)
		}
	}
	if !strings.Contains(got, "&lt;script") || !strings.Contains(got, "&lt;a ") {
		t.Fatalf("expected escaped tags, got %q", got)
	}
}

func TestSanitizeHTMLEscapesBoldWithAttrs(t *testing.T) {
	t.Parallel()
	got := SanitizeHTML(`<b onclick="x">bad</b>`)
	if strings.Contains(got, `<b`) || strings.Contains(got, `</b>`) {
		t.Fatalf("attribute bold should not become real tags: %q", got)
	}
	if strings.Contains(got, "onclick=\"") {
		t.Fatalf("raw onclick attribute leaked: %q", got)
	}
}

func TestSanitizeHTMLTruncates(t *testing.T) {
	t.Parallel()
	runes := make([]rune, MaxReviewBodyRunes+50)
	for i := range runes {
		runes[i] = 'я'
	}
	got := SanitizeHTML(string(runes))
	if utf8.RuneCountInString(got) != MaxReviewBodyRunes {
		t.Fatalf("len=%d want %d", utf8.RuneCountInString(got), MaxReviewBodyRunes)
	}
}

func TestEscapePlain(t *testing.T) {
	t.Parallel()
	got := EscapePlain(`task <b>x</b> & y`)
	want := `task &lt;b&gt;x&lt;/b&gt; &amp; y`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
