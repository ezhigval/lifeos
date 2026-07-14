package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyName     = errors.New("project name is required")
	ErrInvalidStatus = errors.New("invalid project status")
	ErrNotFound      = errors.New("project not found")
	ErrNotActive     = errors.New("project is not active")
	ErrNoSphere      = errors.New("at least one sphere is required")
	ErrNoTarget      = errors.New("project has no target value")
)

type Status string

const (
	StatusActive    Status = "active"
	StatusArchived  Status = "archived"
	StatusCompleted Status = "completed"
)

func (s Status) Valid() bool {
	return s == StatusActive || s == StatusArchived || s == StatusCompleted
}

type Project struct {
	ID           ids.ProjectID
	UserID       ids.UserID
	Name         string
	Description  *string
	Outcome      *string
	Status       Status
	TargetValue  *decimal.Decimal
	CurrentValue decimal.Decimal
	Unit         *string
	TargetDate   *time.Time
	SphereIDs    []ids.SphereID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Progress struct {
	Target    decimal.Decimal
	Current   decimal.Decimal
	Remaining decimal.Decimal
	Percent   decimal.Decimal
	Unit      string
	HasTarget bool
}

func NewProject(userID ids.UserID, name string, sphereIDs []ids.SphereID, now time.Time) (Project, error) {
	if name == "" {
		return Project{}, ErrEmptyName
	}
	if len(sphereIDs) == 0 {
		return Project{}, ErrNoSphere
	}
	now = now.UTC()
	return Project{
		ID:           ids.NewProjectID(),
		UserID:       userID,
		Name:         name,
		Status:       StatusActive,
		CurrentValue: decimal.Zero,
		SphereIDs:    sphereIDs,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (p *Project) Archive() error {
	if p.Status != StatusActive {
		return ErrNotActive
	}
	p.Status = StatusArchived
	return nil
}

func (p Project) Progress() (Progress, error) {
	if p.TargetValue == nil {
		return Progress{Current: p.CurrentValue, HasTarget: false}, ErrNoTarget
	}
	target := *p.TargetValue
	remaining := target.Sub(p.CurrentValue)
	if remaining.IsNegative() {
		remaining = decimal.Zero
	}
	percent := decimal.Zero
	if !target.IsZero() {
		percent = p.CurrentValue.Div(target).Mul(decimal.NewFromInt(100))
	}
	unit := ""
	if p.Unit != nil {
		unit = *p.Unit
	}
	return Progress{
		Target:    target,
		Current:   p.CurrentValue,
		Remaining: remaining,
		Percent:   percent,
		Unit:      unit,
		HasTarget: true,
	}, nil
}
