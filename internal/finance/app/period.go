package app

import (
	"context"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

func monthBounds(ctx context.Context, users UserReader, userID ids.UserID, now time.Time) (time.Time, time.Time, error) {
	tz, err := users.Timezone(ctx, userID)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return timeutil.MonthBounds(now, tz)
}
