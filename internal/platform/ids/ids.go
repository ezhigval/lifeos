package ids

import "github.com/google/uuid"

type UserID uuid.UUID
type TaskID uuid.UUID
type GoalID uuid.UUID
type TransactionID uuid.UUID
type CategoryID uuid.UUID
type DebtID uuid.UUID
type PlannedCashflowID uuid.UUID
type HabitID uuid.UUID
type HabitLogID uuid.UUID
type ProjectID uuid.UUID
type EventID uuid.UUID
type NoteID uuid.UUID
type WeightLogID uuid.UUID
type StepLogID uuid.UUID
type SleepLogID uuid.UUID
type ContactID uuid.UUID
type SkillID uuid.UUID
type SphereID uuid.UUID

func NewUserID() UserID {
	return UserID(uuid.Must(uuid.NewV7()))
}

func ParseUserID(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, err
	}
	return UserID(id), nil
}

func (id UserID) String() string {
	return uuid.UUID(id).String()
}

func (id UserID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id UserID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewTaskID() TaskID {
	return TaskID(uuid.Must(uuid.NewV7()))
}

func ParseTaskID(s string) (TaskID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TaskID{}, err
	}
	return TaskID(id), nil
}

func (id TaskID) String() string {
	return uuid.UUID(id).String()
}

func (id TaskID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id TaskID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewGoalID() GoalID {
	return GoalID(uuid.Must(uuid.NewV7()))
}

func ParseGoalID(s string) (GoalID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return GoalID{}, err
	}
	return GoalID(id), nil
}

func (id GoalID) String() string {
	return uuid.UUID(id).String()
}

func (id GoalID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id GoalID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewTransactionID() TransactionID {
	return TransactionID(uuid.Must(uuid.NewV7()))
}

func ParseTransactionID(s string) (TransactionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TransactionID{}, err
	}
	return TransactionID(id), nil
}

func (id TransactionID) String() string {
	return uuid.UUID(id).String()
}

func (id TransactionID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id TransactionID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewCategoryID() CategoryID {
	return CategoryID(uuid.Must(uuid.NewV7()))
}

func (id CategoryID) String() string {
	return uuid.UUID(id).String()
}

func (id CategoryID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id CategoryID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewDebtID() DebtID {
	return DebtID(uuid.Must(uuid.NewV7()))
}

func (id DebtID) String() string {
	return uuid.UUID(id).String()
}

func (id DebtID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id DebtID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func ParseDebtID(s string) (DebtID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return DebtID{}, err
	}
	return DebtID(id), nil
}

func NewPlannedCashflowID() PlannedCashflowID {
	return PlannedCashflowID(uuid.Must(uuid.NewV7()))
}

func (id PlannedCashflowID) String() string {
	return uuid.UUID(id).String()
}

func (id PlannedCashflowID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id PlannedCashflowID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func ParsePlannedCashflowID(s string) (PlannedCashflowID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return PlannedCashflowID{}, err
	}
	return PlannedCashflowID(id), nil
}

func NewHabitID() HabitID {
	return HabitID(uuid.Must(uuid.NewV7()))
}

func ParseHabitID(s string) (HabitID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return HabitID{}, err
	}
	return HabitID(id), nil
}

func (id HabitID) String() string {
	return uuid.UUID(id).String()
}

func (id HabitID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id HabitID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewHabitLogID() HabitLogID {
	return HabitLogID(uuid.Must(uuid.NewV7()))
}

func (id HabitLogID) String() string {
	return uuid.UUID(id).String()
}

func (id HabitLogID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id HabitLogID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewProjectID() ProjectID {
	return ProjectID(uuid.Must(uuid.NewV7()))
}

func ParseProjectID(s string) (ProjectID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ProjectID{}, err
	}
	return ProjectID(id), nil
}

func (id ProjectID) String() string {
	return uuid.UUID(id).String()
}

func (id ProjectID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ProjectID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewEventID() EventID {
	return EventID(uuid.Must(uuid.NewV7()))
}

func (id EventID) String() string {
	return uuid.UUID(id).String()
}

func (id EventID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func NewNoteID() NoteID {
	return NoteID(uuid.Must(uuid.NewV7()))
}

func ParseNoteID(s string) (NoteID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return NoteID{}, err
	}
	return NoteID(id), nil
}

func (id NoteID) String() string {
	return uuid.UUID(id).String()
}

func (id NoteID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id NoteID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewWeightLogID() WeightLogID {
	return WeightLogID(uuid.Must(uuid.NewV7()))
}

func ParseWeightLogID(s string) (WeightLogID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return WeightLogID{}, err
	}
	return WeightLogID(id), nil
}

func (id WeightLogID) String() string {
	return uuid.UUID(id).String()
}

func (id WeightLogID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id WeightLogID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewStepLogID() StepLogID {
	return StepLogID(uuid.Must(uuid.NewV7()))
}

func (id StepLogID) String() string {
	return uuid.UUID(id).String()
}

func (id StepLogID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id StepLogID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewSleepLogID() SleepLogID {
	return SleepLogID(uuid.Must(uuid.NewV7()))
}

func (id SleepLogID) String() string {
	return uuid.UUID(id).String()
}

func (id SleepLogID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id SleepLogID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewContactID() ContactID {
	return ContactID(uuid.Must(uuid.NewV7()))
}

func ParseContactID(s string) (ContactID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ContactID{}, err
	}
	return ContactID(id), nil
}

func (id ContactID) String() string {
	return uuid.UUID(id).String()
}

func (id ContactID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ContactID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewSkillID() SkillID {
	return SkillID(uuid.Must(uuid.NewV7()))
}

func ParseSkillID(s string) (SkillID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SkillID{}, err
	}
	return SkillID(id), nil
}

func (id SkillID) String() string {
	return uuid.UUID(id).String()
}

func (id SkillID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id SkillID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewSphereID() SphereID {
	return SphereID(uuid.Must(uuid.NewV7()))
}

func ParseSphereID(s string) (SphereID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SphereID{}, err
	}
	return SphereID(id), nil
}

func (id SphereID) String() string {
	return uuid.UUID(id).String()
}

func (id SphereID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id SphereID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

type MemoryID uuid.UUID
type LearningEventID uuid.UUID

func NewMemoryID() MemoryID {
	return MemoryID(uuid.Must(uuid.NewV7()))
}

func ParseMemoryID(s string) (MemoryID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return MemoryID{}, err
	}
	return MemoryID(id), nil
}

func (id MemoryID) String() string {
	return uuid.UUID(id).String()
}

func (id MemoryID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id MemoryID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewLearningEventID() LearningEventID {
	return LearningEventID(uuid.Must(uuid.NewV7()))
}

func ParseLearningEventID(s string) (LearningEventID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return LearningEventID{}, err
	}
	return LearningEventID(id), nil
}

func (id LearningEventID) String() string {
	return uuid.UUID(id).String()
}

func (id LearningEventID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id LearningEventID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}
