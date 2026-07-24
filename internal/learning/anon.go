package learning

import (
	"github.com/valentinezhov/lifeos/internal/learning/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// AnonSubject returns a stable HMAC-SHA256 hex digest for a user id.
func AnonSubject(userID ids.UserID, salt string) string {
	return domain.AnonSubject(userID, salt)
}

// Event is an anonymized learning signal.
type Event = domain.Event
