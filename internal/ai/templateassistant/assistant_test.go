package templateassistant

import (
	"context"
	"strings"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
)

func TestSummarizeEscapesTaskTitles(t *testing.T) {
	t.Parallel()
	got, err := New().Summarize(context.Background(), ai.SummaryRequest{
		Tasks:    []string{`<script>x</script>`, `ok & fine`},
		Projects: []string{`<b>proj</b>`},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "<script") || strings.Contains(got, "<b>proj</b>") {
		t.Fatalf("unsafe HTML leaked: %q", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") || !strings.Contains(got, "&amp;") {
		t.Fatalf("expected escapes: %q", got)
	}
	if !strings.Contains(got, "&lt;b&gt;proj&lt;/b&gt;") {
		t.Fatalf("project not escaped: %q", got)
	}
}
