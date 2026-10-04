package app_test

import (
	"encoding/json"
	"testing"
	"time"

	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestScheduleReminderInputTaskIDInPayloadShape(t *testing.T) {
	t.Parallel()
	in := notifapp.ScheduleReminderInput{
		UserID:  ids.NewUserID(),
		Message: "🔔 call mom",
		FireAt:  time.Now().UTC().Add(time.Hour),
		TaskID:  "019f8027-4b8c-7819-82b5-00f0577ab6f9",
	}
	payload, err := json.Marshal(map[string]string{
		"message": in.Message,
		"task_id": in.TaskID,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatal(err)
	}
	if got["task_id"] != in.TaskID || got["message"] != in.Message {
		t.Fatalf("payload=%v", got)
	}
}
