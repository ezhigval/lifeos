package rulebased_test

import (
	"context"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
)

func TestResolverIntents(t *testing.T) {
	t.Parallel()
	r := rulebased.NewResolver()
	cases := []struct {
		in   string
		want ai.IntentType
	}{
		{"Добавь задачу купить фильтр", ai.IntentTaskCreate},
		{"задачи на сегодня", ai.IntentTaskListToday},
		{"выполни задачу купить фильтр", ai.IntentTaskComplete},
		{"Добавь проект свадьба в сфере Личная жизнь", ai.IntentProjectCreate},
		{"прогресс проекта", ai.IntentProjectProgress},
		{"что сейчас важное", ai.IntentQueryPriorities},
		{"напомни вечером позвонить", ai.IntentReminderCreate},
		{"отмени напоминание позвонить", ai.IntentReminderCancel},
		{"работаю до 15", ai.IntentPlanSetAvailability},
		{"сегодня полный завал", ai.IntentPlanTriage},
		{"перенеси задачи", ai.IntentTaskReschedule},
		{"привет", ai.IntentUnknown},
		{"заплатил 10 тысяч банку", ai.IntentFinancePayDebt},
		{"архивируй проект свадьба", ai.IntentProjectArchive},
		{"запиши заметку идея для Jarvis", ai.IntentNoteCreate},
		{"мои заметки", ai.IntentNoteList},
		{"заметки с тегом work", ai.IntentNoteList},
		{"найди заметку Jarvis", ai.IntentNoteSearch},
		{"удали заметку jarvis", ai.IntentNoteDelete},
		{"вес 78.5", ai.IntentHealthRecordWeight},
		{"мой вес", ai.IntentHealthLatestWeight},
		{"шаги 8000", ai.IntentHealthRecordSteps},
		{"мои шаги", ai.IntentHealthLatestSteps},
		{"спал 7 часов", ai.IntentHealthRecordSleep},
		{"мой сон", ai.IntentHealthLatestSleep},
		{"добавь контакт Иван — Яндекс", ai.IntentCareerContactCreate},
		{"контакты", ai.IntentCareerContactList},
		{"найди контакт Иван", ai.IntentCareerContactSearch},
		{"навык Go senior", ai.IntentCareerSkillCreate},
		{"навыки", ai.IntentCareerSkillList},
	}
	for _, tc := range cases {
		got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: tc.in})
		if err != nil {
			t.Fatalf("Resolve(%q): %v", tc.in, err)
		}
		if got.Type != tc.want {
			t.Fatalf("Resolve(%q) = %s, want %s", tc.in, got.Type, tc.want)
		}
	}
}
