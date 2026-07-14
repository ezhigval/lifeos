package pgconv

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func UUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func UserID(id ids.UserID) pgtype.UUID {
	return UUID(id.UUID())
}

func TaskID(id ids.TaskID) pgtype.UUID {
	return UUID(id.UUID())
}

func GoalID(id ids.GoalID) pgtype.UUID {
	return UUID(id.UUID())
}

func TransactionID(id ids.TransactionID) pgtype.UUID {
	return UUID(id.UUID())
}

func CategoryID(id ids.CategoryID) pgtype.UUID {
	return UUID(id.UUID())
}

func DebtID(id ids.DebtID) pgtype.UUID {
	return UUID(id.UUID())
}

func HabitID(id ids.HabitID) pgtype.UUID {
	return UUID(id.UUID())
}

func ProjectID(id ids.ProjectID) pgtype.UUID {
	return UUID(id.UUID())
}

func EventID(id ids.EventID) pgtype.UUID {
	return UUID(id.UUID())
}

func NoteID(id ids.NoteID) pgtype.UUID {
	return UUID(id.UUID())
}

func WeightLogID(id ids.WeightLogID) pgtype.UUID {
	return UUID(id.UUID())
}

func StepLogID(id ids.StepLogID) pgtype.UUID {
	return UUID(id.UUID())
}

func SleepLogID(id ids.SleepLogID) pgtype.UUID {
	return UUID(id.UUID())
}

func ContactID(id ids.ContactID) pgtype.UUID {
	return UUID(id.UUID())
}

func SkillID(id ids.SkillID) pgtype.UUID {
	return UUID(id.UUID())
}

func FromUUID(v pgtype.UUID) uuid.UUID {
	if !v.Valid {
		return uuid.Nil
	}
	return uuid.UUID(v.Bytes)
}

func FromUserID(v pgtype.UUID) ids.UserID {
	return ids.UserID(FromUUID(v))
}

func FromTaskID(v pgtype.UUID) ids.TaskID {
	return ids.TaskID(FromUUID(v))
}

func FromGoalID(v pgtype.UUID) ids.GoalID {
	return ids.GoalID(FromUUID(v))
}

func FromTransactionID(v pgtype.UUID) ids.TransactionID {
	return ids.TransactionID(FromUUID(v))
}

func FromCategoryID(v pgtype.UUID) ids.CategoryID {
	return ids.CategoryID(FromUUID(v))
}

func FromDebtID(v pgtype.UUID) ids.DebtID {
	return ids.DebtID(FromUUID(v))
}

func FromHabitID(v pgtype.UUID) ids.HabitID {
	return ids.HabitID(FromUUID(v))
}

func FromProjectID(v pgtype.UUID) ids.ProjectID {
	return ids.ProjectID(FromUUID(v))
}

func FromEventID(v pgtype.UUID) ids.EventID {
	return ids.EventID(FromUUID(v))
}

func FromNoteID(v pgtype.UUID) ids.NoteID {
	return ids.NoteID(FromUUID(v))
}

func FromWeightLogID(v pgtype.UUID) ids.WeightLogID {
	return ids.WeightLogID(FromUUID(v))
}

func FromStepLogID(v pgtype.UUID) ids.StepLogID {
	return ids.StepLogID(FromUUID(v))
}

func FromSleepLogID(v pgtype.UUID) ids.SleepLogID {
	return ids.SleepLogID(FromUUID(v))
}

func FromContactID(v pgtype.UUID) ids.ContactID {
	return ids.ContactID(FromUUID(v))
}

func FromSkillID(v pgtype.UUID) ids.SkillID {
	return ids.SkillID(FromUUID(v))
}

func SphereID(id ids.SphereID) pgtype.UUID {
	return UUID(id.UUID())
}

func FromSphereID(v pgtype.UUID) ids.SphereID {
	return ids.SphereID(FromUUID(v))
}

func Date(t time.Time) pgtype.Date {
	return pgtype.Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), Valid: true}
}

func DatePtr(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return Date(*t)
}

func FromDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	t := time.Date(d.Time.Year(), d.Time.Month(), d.Time.Day(), 0, 0, 0, 0, time.UTC)
	return &t
}

func Timestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func TimestamptzValue(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func FromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func Text(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func FromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func Int8Nullable(v int64) pgtype.Int8 {
	if v == 0 {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: v, Valid: true}
}

func FromInt8(v pgtype.Int8) int64 {
	if !v.Valid {
		return 0
	}
	return v.Int64
}
