package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/career/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrSkillNotFound = errors.New("skill not found")

type DeleteSkill struct {
	store      SkillStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewDeleteSkill(store SkillStore, events EventLog, transactor Transactor) *DeleteSkill {
	return &DeleteSkill{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type DeleteSkillInput struct {
	UserID  ids.UserID
	SkillID ids.SkillID
	Source  events.Source
}

func (uc *DeleteSkill) Execute(ctx context.Context, in DeleteSkillInput) (SkillDTO, error) {
	if in.UserID.IsZero() || in.SkillID.IsZero() {
		return SkillDTO{}, fmt.Errorf("user id and skill id are required")
	}
	now := uc.now()
	var deleted SkillDTO
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		skill, err := uc.store.DeleteSkill(txCtx, in.UserID, in.SkillID)
		if errors.Is(err, domain.ErrSkillNotFound) {
			return ErrSkillNotFound
		}
		if err != nil {
			return err
		}
		deleted = ToSkillDTO(skill)
		return uc.events.Append(txCtx, events.Record{
			UserID:        skill.UserID,
			AggregateType: "skill",
			AggregateID:   skill.ID.UUID(),
			EventType:     "SkillDeleted",
			Payload:       map[string]any{"name": skill.Name},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		if errors.Is(err, ErrSkillNotFound) {
			return SkillDTO{}, ErrSkillNotFound
		}
		return SkillDTO{}, fmt.Errorf("delete skill: %w", err)
	}
	return deleted, nil
}
