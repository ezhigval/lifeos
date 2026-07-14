package telegram

import (
	"errors"
	"fmt"
	"testing"

	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

func TestFormatTaskLifecycleCallbackError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		err    error
		cancel bool
		want   string
	}{
		{"not found cancel", fmt.Errorf("cancel task: %w", tasksapp.ErrTaskNotFound), true, "Задача не найдена."},
		{"not found complete", fmt.Errorf("complete task: %w", tasksapp.ErrTaskNotFound), false, "Задача не найдена."},
		{"already cancelled", fmt.Errorf("cancel task: %w", taskdomain.ErrAlreadyCancelled), true, "Задача уже отменена."},
		{"cannot cancel done", fmt.Errorf("cancel task: %w", taskdomain.ErrCannotCancelDone), true, "Выполненную задачу нельзя отменить."},
		{"cannot complete", fmt.Errorf("complete task: %w", taskdomain.ErrCannotComplete), false, "Отменённую задачу нельзя выполнить."},
		{"unknown", errors.New("boom"), true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatTaskLifecycleCallbackError(tc.err, tc.cancel)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
