package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/career/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type SkillStore interface {
	SaveSkill(ctx context.Context, skill domain.Skill) error
	ListRecentSkills(ctx context.Context, userID ids.UserID, limit int32) ([]domain.Skill, error)
	SearchSkills(ctx context.Context, userID ids.UserID, query string, limit int32) ([]domain.Skill, error)
	DeleteSkill(ctx context.Context, userID ids.UserID, skillID ids.SkillID) (domain.Skill, error)
}

type SkillDTO struct {
	ID        ids.SkillID
	Name      string
	Level     string
	CreatedAt time.Time
}

func ToSkillDTO(s domain.Skill) SkillDTO {
	return SkillDTO{ID: s.ID, Name: s.Name, Level: s.Level, CreatedAt: s.CreatedAt}
}

type CreateSkill struct {
	store      SkillStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateSkill(store SkillStore, events EventLog, transactor Transactor) *CreateSkill {
	return &CreateSkill{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateSkillInput struct {
	UserID ids.UserID
	Name   string
	Level  string
	Source events.Source
}

func (uc *CreateSkill) Execute(ctx context.Context, in CreateSkillInput) (SkillDTO, error) {
	if in.UserID.IsZero() {
		return SkillDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	skill, err := domain.NewSkill(in.UserID, in.Name, in.Level, now)
	if err != nil {
		return SkillDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.SaveSkill(txCtx, skill); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        skill.UserID,
			AggregateType: "skill",
			AggregateID:   skill.ID.UUID(),
			EventType:     "SkillCreated",
			Payload:       map[string]any{"name": skill.Name, "level": skill.Level},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return SkillDTO{}, fmt.Errorf("create skill: %w", err)
	}
	return ToSkillDTO(skill), nil
}

type ListSkills struct {
	store SkillStore
}

func NewListSkills(store SkillStore) *ListSkills {
	return &ListSkills{store: store}
}

func (uc *ListSkills) Execute(ctx context.Context, userID ids.UserID) ([]SkillDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	items, err := uc.store.ListRecentSkills(ctx, userID, defaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	out := make([]SkillDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToSkillDTO(item))
	}
	return out, nil
}

type SearchSkills struct {
	store SkillStore
}

func NewSearchSkills(store SkillStore) *SearchSkills {
	return &SearchSkills{store: store}
}

type SearchSkillsInput struct {
	UserID ids.UserID
	Query  string
}

func (uc *SearchSkills) Execute(ctx context.Context, in SearchSkillsInput) ([]SkillDTO, error) {
	if in.UserID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}
	items, err := uc.store.SearchSkills(ctx, in.UserID, query, defaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("search skills: %w", err)
	}
	out := make([]SkillDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToSkillDTO(item))
	}
	return out, nil
}
