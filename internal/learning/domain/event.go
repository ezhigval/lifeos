package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// AnonSubject returns a stable HMAC-SHA256 hex digest for a user id.
// Never store or log the raw user id alongside learning events.
func AnonSubject(userID ids.UserID, salt string) string {
	mac := hmac.New(sha256.New, []byte(salt))
	_, _ = mac.Write([]byte(userID.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// Event is an anonymized learning signal. Meta must contain only coarse flags,
// never raw messages, names, or amounts as plaintext PII.
type Event struct {
	ID           ids.LearningEventID
	AnonSubject  string
	Type         string
	ToolOrIntent string
	Success      *bool
	AskRounds    int
	Model        string
	Meta         map[string]any
	CreatedAt    time.Time
}
