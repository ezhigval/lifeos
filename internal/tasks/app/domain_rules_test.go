package app_test

import (
	"context"
	"testing"
	"time"

	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	habitsdomain "github.com/valentinezhov/lifeos/internal/habits/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	spheresdomain "github.com/valentinezhov/lifeos/internal/spheres/domain"
	"github.com/valentinezhov/lifeos/internal/tasks/app"
)

type rulesSphereRepo struct {
	spheres []spheresdomain.Sphere
}

func (r *rulesSphereRepo) List(context.Context, ids.UserID) ([]spheresdomain.Sphere, error) {
	return r.spheres, nil
}

type rulesPlanStore struct {
	items []financedomain.PlannedCashflow
}

func (s *rulesPlanStore) ListPlanned(context.Context, ids.UserID) ([]financedomain.PlannedCashflow, error) {
	return s.items, nil
}

func (s *rulesPlanStore) SavePlanned(_ context.Context, item financedomain.PlannedCashflow) error {
	s.items = append(s.items, item)
	return nil
}

type rulesHabitStore struct {
	habits []habitsdomain.Habit
}

func (h *rulesHabitStore) FindByName(_ context.Context, userID ids.UserID, name string) (habitsdomain.Habit, error) {
	for _, hab := range h.habits {
		if hab.UserID == userID && hab.Name == name {
			return hab, nil
		}
	}
	return habitsdomain.Habit{}, habitsdomain.ErrNotFound
}

func (h *rulesHabitStore) Save(_ context.Context, habit habitsdomain.Habit) error {
	h.habits = append(h.habits, habit)
	return nil
}

func sphereID(name string) (ids.SphereID, []spheresdomain.Sphere) {
	id := ids.NewSphereID()
	return id, []spheresdomain.Sphere{{ID: id, Name: name}}
}

func TestDomainRulesMoneyCreatesPlannedIncomeIdempotent(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	sid, spheres := sphereID("Деньги")
	plans := &rulesPlanStore{}
	rules := app.NewDomainRules(&rulesSphereRepo{spheres}, plans, &rulesHabitStore{})

	due := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	task := app.TaskDTO{ID: ids.NewTaskID(), Title: "перевести зарплату", SphereIDs: []ids.SphereID{sid}, DueDate: &due}
	if err := rules.Apply(context.Background(), userID, task); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(plans.items) != 1 {
		t.Fatalf("planned income not created: %+v", plans.items)
	}
	it := plans.items[0]
	if it.Kind != financedomain.PlanKindIncome || it.Title != "Зарплата" || it.Interval != financedomain.PlanIntervalMonthly {
		t.Fatalf("unexpected plan %+v", it)
	}
	// Идемпотентность: повтор не создаёт дубль.
	if err := rules.Apply(context.Background(), userID, task); err != nil {
		t.Fatalf("Apply second: %v", err)
	}
	if len(plans.items) != 1 {
		t.Fatalf("duplicate plan created: %+v", plans.items)
	}
}

func TestDomainRulesMoneyIgnoresOtherTitles(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	sid, spheres := sphereID("Деньги")
	plans := &rulesPlanStore{}
	rules := app.NewDomainRules(&rulesSphereRepo{spheres}, plans, &rulesHabitStore{})
	task := app.TaskDTO{Title: "купить молоко", SphereIDs: []ids.SphereID{sid}}
	if err := rules.Apply(context.Background(), userID, task); err != nil {
		t.Fatal(err)
	}
	if len(plans.items) != 0 {
		t.Fatalf("unexpected plan: %+v", plans.items)
	}
}

func TestDomainRulesHealthCreatesHabitOnce(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	sid, spheres := sphereID("Здоровье")
	habits := &rulesHabitStore{}
	rules := app.NewDomainRules(&rulesSphereRepo{spheres}, &rulesPlanStore{}, habits)

	task := app.TaskDTO{Title: "забег", SphereIDs: []ids.SphereID{sid}}
	if err := rules.Apply(context.Background(), userID, task); err != nil {
		t.Fatal(err)
	}
	if len(habits.habits) != 1 || habits.habits[0].Frequency != habitsdomain.FrequencyDaily {
		t.Fatalf("habit not created: %+v", habits.habits)
	}
	if err := rules.Apply(context.Background(), userID, task); err != nil {
		t.Fatal(err)
	}
	if len(habits.habits) != 1 {
		t.Fatalf("duplicate habit: %+v", habits.habits)
	}
}

func TestDomainRulesNoSpheresIsNoop(t *testing.T) {
	t.Parallel()
	rules := app.NewDomainRules(&rulesSphereRepo{}, &rulesPlanStore{}, &rulesHabitStore{})
	if err := rules.Apply(context.Background(), ids.NewUserID(), app.TaskDTO{Title: "зарплата"}); err != nil {
		t.Fatal(err)
	}
}
