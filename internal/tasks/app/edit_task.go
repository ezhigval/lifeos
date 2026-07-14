package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type EditTask struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	projects   ProjectChecker
	now        func() time.Time
}

func NewEditTask(store TaskStore, events EventLog, transactor Transactor, projects ProjectChecker) *EditTask {
	return &EditTask{
		store: store, events: events, transactor: transactor, projects: projects,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type EditTaskInput struct {
	UserID           ids.UserID
	TaskID           ids.TaskID
	Title            *string
	Description      *string
	ClearDescription bool
	Priority         *domain.Priority
	DueDate          *time.Time
	ClearDueDate     bool
	DurationMinutes  *int
	ClearDuration    bool
	Tags             *[]string
	Kind             *domain.Kind
	Address          *string
	ClearAddress     bool
	NoteID           *ids.NoteID
	ClearNoteID      bool
	ProjectIDs       *[]ids.ProjectID
	Source           events.Source
}

func (uc *EditTask) Execute(ctx context.Context, in EditTaskInput) (TaskDTO, error) {
	if in.UserID.IsZero() || in.TaskID.IsZero() {
		return TaskDTO{}, fmt.Errorf("user id and task id are required")
	}

	now := uc.now()
	var result TaskDTO
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := uc.store.GetByID(txCtx, in.UserID, in.TaskID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return ErrTaskNotFound
			}
			return err
		}
		if err := task.Edit(domain.EditFields{
			Title:            in.Title,
			Description:      in.Description,
			ClearDescription: in.ClearDescription,
			Priority:         in.Priority,
			DueDate:          in.DueDate,
			ClearDueDate:     in.ClearDueDate,
			DurationMinutes:  in.DurationMinutes,
			ClearDuration:    in.ClearDuration,
			Tags:             in.Tags,
			Kind:             in.Kind,
			Address:          in.Address,
			ClearAddress:     in.ClearAddress,
			NoteID:           in.NoteID,
			ClearNoteID:      in.ClearNoteID,
		}, now); err != nil {
			return err
		}

		if in.ProjectIDs != nil {
			idsCopy := append([]ids.ProjectID(nil), (*in.ProjectIDs)...)
			if len(idsCopy) > 0 && uc.projects != nil {
				ok, err := uc.projects.AllExist(txCtx, in.UserID, idsCopy)
				if err != nil {
					return fmt.Errorf("validate projects: %w", err)
				}
				if !ok {
					return fmt.Errorf("project not found")
				}
			}
			task.ProjectIDs = idsCopy
			if err := uc.store.SetProjects(txCtx, task.ID, task.ProjectIDs); err != nil {
				return err
			}
		}

		if err := uc.store.Update(txCtx, task); err != nil {
			return err
		}
		if err := uc.events.Append(txCtx, events.Record{
			UserID:        task.UserID,
			AggregateType: "task",
			AggregateID:   task.ID.UUID(),
			EventType:     "TaskUpdated",
			Payload: map[string]any{
				"title": task.Title, "priority": task.Priority, "due_date": task.DueDate,
				"duration_minutes": task.DurationMinutes, "tags": task.Tags, "project_ids": task.ProjectIDs,
			},
			Source:     in.Source,
			OccurredAt: now,
		}); err != nil {
			return err
		}
		result = ToDTO(task)
		return nil
	})
	if err != nil {
		return TaskDTO{}, fmt.Errorf("edit task: %w", err)
	}
	return result, nil
}
