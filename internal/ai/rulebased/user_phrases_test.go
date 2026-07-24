package rulebased_test

import (
	"context"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
)

func TestResolveCommonUserPhrases(t *testing.T) {
	r := rulebased.NewResolver()
	cases := []struct {
		text string
		typ  ai.IntentType
	}{
		{"добавь задачу купить хлеб", ai.IntentTaskCreate},
		{"потратил 300 на кофе", ai.IntentFinanceExpense},
		{"задачи на сегодня", ai.IntentTaskListToday},
	}
	for _, c := range cases {
		got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: c.text, Language: "ru"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Type != c.typ {
			t.Fatalf("%q: got %s want %s", c.text, got.Type, c.typ)
		}
		if c.typ == ai.IntentTaskCreate && got.Title == "" {
			t.Fatalf("%q empty title", c.text)
		}
	}
}
