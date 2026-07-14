package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type TokenService struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenService(secret string, ttl time.Duration) (*TokenService, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
		now:    func() time.Time { return time.Now().UTC() },
	}, nil
}

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

func (s *TokenService) Issue(userID ids.UserID) (string, time.Time, error) {
	if userID.IsZero() {
		return "", time.Time{}, fmt.Errorf("user id is required")
	}
	now := s.now()
	exp := now.Add(s.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    "lifeos",
		},
	})
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

func (s *TokenService) Parse(token string) (ids.UserID, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return ids.UserID{}, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return ids.UserID{}, fmt.Errorf("invalid token claims")
	}
	return ids.ParseUserID(claims.UserID)
}
