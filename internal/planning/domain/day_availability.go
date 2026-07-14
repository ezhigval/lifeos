package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrInvalidDay = errors.New("day is required")

type DayAvailability struct {
	ID             ids.TaskID // reuse as generic uuid holder or add AvailabilityID
	UserID         ids.UserID
	Day            time.Time
	AvailableUntil time.Time
	Note           *string
}

func NewDayAvailability(userID ids.UserID, day time.Time, until time.Time) (DayAvailability, error) {
	if userID.IsZero() {
		return DayAvailability{}, errors.New("user id required")
	}
	return DayAvailability{
		ID:             ids.NewTaskID(),
		UserID:         userID,
		Day:            day,
		AvailableUntil: until,
	}, nil
}
