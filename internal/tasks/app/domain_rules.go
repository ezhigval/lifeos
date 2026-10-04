package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	habitsdomain "github.com/valentinezhov/lifeos/internal/habits/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	spheresdomain "github.com/valentinezhov/lifeos/internal/spheres/domain"
)

// DomainRules — домен-правила TASK-011 п.8: связывает задачи со смежными доменами
// по сферам задачи (N:M из п.7). Идемпотентность: перед созданием сущности
// проверяется существование аналога (FindByName / ListPlanned), повтор не создаёт дубли.
type DomainRules struct {
	spheres sphereNameReader
	plans   plannedCashflowLister
	habits  habitStoreLike
	now     func() time.Time
}

// sphereNameReader — минимальный порт spheres (реализует spheres/infra.Repository).
type sphereNameReader interface {
	List(ctx context.Context, userID ids.UserID) ([]spheresdomain.Sphere, error)
}

// plannedCashflowLister — минимальный порт finance (реализует finance/infra.Repository).
type plannedCashflowLister interface {
	ListPlanned(ctx context.Context, userID ids.UserID) ([]financedomain.PlannedCashflow, error)
	SavePlanned(ctx context.Context, item financedomain.PlannedCashflow) error
}

// habitStoreLike — минимальный порт habits (реализует habits/infra.Repository).
type habitStoreLike interface {
	FindByName(ctx context.Context, userID ids.UserID, name string) (habitsdomain.Habit, error)
	Save(ctx context.Context, habit habitsdomain.Habit) error
}

func NewDomainRules(spheres sphereNameReader, plans plannedCashflowLister, habits habitStoreLike) *DomainRules {
	return &DomainRules{
		spheres: spheres, plans: plans, habits: habits,
		now: func() time.Time { return time.Now().UTC() },
	}
}

const (
	moneySphereName  = "Деньги"
	healthSphereName = "Здоровье"

	salaryTitleKeyword = "зарплат" // "зарплата", "зарплату"…
	salaryPlanTitle    = "Зарплата"
	salaryAmountCents  = int64(1) // домен требует положительную сумму; placeholder — пользователь отредактирует в плане
	wellnessHabitName  = "Спорт / здоровье (авто)"
)

// Apply выполняет правила для задачи после её создания/редактирования.
// Ошибки правил не должны ронять основную операцию — вызывающий логирует warn.
func (dr *DomainRules) Apply(ctx context.Context, userID ids.UserID, task TaskDTO) error {
	if dr == nil || len(task.SphereIDs) == 0 {
		return nil
	}
	names, err := dr.sphereNames(ctx, userID, task.SphereIDs)
	if err != nil {
		return fmt.Errorf("domain rules: load spheres: %w", err)
	}
	for _, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case strings.ToLower(moneySphereName):
			if err := dr.applyMoneyRule(ctx, userID, task); err != nil {
				return err
			}
		case strings.ToLower(healthSphereName):
			if err := dr.applyHealthRule(ctx, userID, task); err != nil {
				return err
			}
		}
	}
	return nil
}

func (dr *DomainRules) sphereNames(ctx context.Context, userID ids.UserID, sphereIDs []ids.SphereID) (map[ids.SphereID]string, error) {
	all, err := dr.spheres.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	byID := make(map[ids.SphereID]string, len(all))
	for _, s := range all {
		byID[s.ID] = s.Name
	}
	out := make(map[ids.SphereID]string, len(sphereIDs))
	for _, id := range sphereIDs {
		if n, ok := byID[id]; ok {
			out[id] = n
		}
	}
	return out, nil
}

// applyMoneyRule: задача в сфере «Деньги» с заголовком про зарплату →
// ежемесячный план дохода (если такого плана ещё нет).
func (dr *DomainRules) applyMoneyRule(ctx context.Context, userID ids.UserID, task TaskDTO) error {
	if !strings.Contains(strings.ToLower(task.Title), salaryTitleKeyword) {
		return nil
	}
	if dr.plans == nil {
		return nil
	}
	items, err := dr.plans.ListPlanned(ctx, userID)
	if err != nil {
		return fmt.Errorf("domain rules: list planned: %w", err)
	}
	for _, it := range items {
		if it.Kind == financedomain.PlanKindIncome && strings.EqualFold(it.Title, salaryPlanTitle) {
			return nil // уже есть — идемпотентно
		}
	}
	next := time.Now().UTC()
	if task.DueDate != nil {
		next = *task.DueDate
	} else {
		next = next.AddDate(0, 1, 0)
	}
	item, err := financedomain.NewPlannedCashflow(
		userID, financedomain.PlanKindIncome, salaryPlanTitle,
		salaryAmountCents, financedomain.PlanIntervalMonthly, next, dr.now(),
	)
	if err != nil {
		return fmt.Errorf("domain rules: build plan: %w", err)
	}
	if err := dr.plans.SavePlanned(ctx, item); err != nil {
		return fmt.Errorf("domain rules: save plan: %w", err)
	}
	return nil
}

// applyHealthRule: задача/проект в сфере «Здоровье» → habit-трекер (если ещё нет).
func (dr *DomainRules) applyHealthRule(ctx context.Context, userID ids.UserID, task TaskDTO) error {
	if dr.habits == nil {
		return nil
	}
	_, err := dr.habits.FindByName(ctx, userID, wellnessHabitName)
	if err == nil {
		return nil // уже есть — идемпотентно
	}
	if err != habitsdomain.ErrNotFound {
		return fmt.Errorf("domain rules: find habit: %w", err)
	}
	habit, err := habitsdomain.NewHabit(userID, wellnessHabitName, habitsdomain.FrequencyDaily, dr.now())
	if err != nil {
		return fmt.Errorf("domain rules: build habit: %w", err)
	}
	if err := dr.habits.Save(ctx, habit); err != nil {
		return fmt.Errorf("domain rules: save habit: %w", err)
	}
	return nil
}
