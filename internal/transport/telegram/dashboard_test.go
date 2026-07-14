package telegram_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	taskapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
	"github.com/valentinezhov/lifeos/internal/transport/telegram"
)

func TestFormatProjectTasksWithActionsSkipsTerminalAndOffersCancel(t *testing.T) {
	t.Parallel()
	doneID := ids.NewTaskID()
	cancelID := ids.NewTaskID()
	openID := ids.NewTaskID()
	_, inline := telegram.FormatProjectTasksWithActions("Demo", []taskapp.TaskDTO{
		{ID: doneID, Title: "done", Status: taskdomain.StatusDone},
		{ID: cancelID, Title: "cancelled", Status: taskdomain.StatusCancelled},
		{ID: openID, Title: "open", Status: taskdomain.StatusTodo},
	})
	if len(inline) != 1 {
		t.Fatalf("want 1 open-task row, got %#v", inline)
	}
	row := inline[0]
	if len(row) != 2 {
		t.Fatalf("want done+cancel buttons, got %#v", row)
	}
	if row[0].CallbackData != telegram.CBTaskDone+openID.String() {
		t.Fatalf("done cb: %s", row[0].CallbackData)
	}
	if row[1].CallbackData != telegram.CBTaskCancel+openID.String() {
		t.Fatalf("cancel cb: %s", row[1].CallbackData)
	}
}
