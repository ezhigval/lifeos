package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

const (
	StateIdle                = "idle"
	StateAwaitTaskTitle      = "await_task_title"
	StateAwaitTaskProjects   = "await_task_projects"
	StateAwaitProjectTitle   = "await_project_title"
	StateAwaitProjectSpheres = "await_project_spheres"
	StateAwaitSphereName     = "await_sphere_name"
	StateAwaitAgentTurn      = "await_agent_turn"
)

type Session struct {
	UserID             ids.UserID
	ChatID             int64
	DashboardMessageID int64
	State              string
	StatePayload       map[string]any
}

type Sessions struct {
	pool *pgxpool.Pool
}

func NewSessions(pool *pgxpool.Pool) *Sessions {
	return &Sessions{pool: pool}
}

func (s *Sessions) Get(ctx context.Context, userID ids.UserID) (Session, error) {
	row, err := db.New(s.pool).GetTelegramSession(ctx, pgconv.UserID(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{UserID: userID, State: StateIdle, StatePayload: map[string]any{}}, nil
	}
	if err != nil {
		return Session{}, fmt.Errorf("get session: %w", err)
	}
	payload := map[string]any{}
	if len(row.StatePayload) > 0 {
		_ = json.Unmarshal(row.StatePayload, &payload)
	}
	var dash int64
	if row.DashboardMessageID.Valid {
		dash = row.DashboardMessageID.Int64
	}
	return Session{
		UserID:             pgconv.FromUserID(row.UserID),
		ChatID:             row.ChatID,
		DashboardMessageID: dash,
		State:              row.State,
		StatePayload:       payload,
	}, nil
}

func (s *Sessions) Save(ctx context.Context, sess Session) error {
	payload, _ := json.Marshal(sess.StatePayload)
	dash := pgconv.Int8Nullable(sess.DashboardMessageID)
	return db.New(s.pool).UpsertTelegramSession(ctx, db.UpsertTelegramSessionParams{
		UserID:             pgconv.UserID(sess.UserID),
		ChatID:             sess.ChatID,
		DashboardMessageID: dash,
		State:              sess.State,
		StatePayload:       payload,
	})
}

func (s *Sessions) SetDashboard(ctx context.Context, userID ids.UserID, chatID, messageID int64) error {
	sess, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	sess.ChatID = chatID
	sess.DashboardMessageID = messageID
	return s.Save(ctx, sess)
}

func (s *Sessions) SetState(ctx context.Context, userID ids.UserID, state string, payload map[string]any) error {
	sess, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	sess.State = state
	if payload == nil {
		payload = map[string]any{}
	}
	sess.StatePayload = payload
	return s.Save(ctx, sess)
}

func (s *Sessions) SetStateOnly(ctx context.Context, userID ids.UserID, state string) error {
	sess, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	sess.State = state
	return s.Save(ctx, sess)
}

func (s *Sessions) UpdatePayload(ctx context.Context, userID ids.UserID, patch map[string]any) error {
	sess, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	if sess.StatePayload == nil {
		sess.StatePayload = map[string]any{}
	}
	for k, v := range patch {
		sess.StatePayload[k] = v
	}
	return s.Save(ctx, sess)
}

// Reset clears conversation UI state (drafts, dashboard pointer, keyboard flag).
// Domain data for the user is not touched.
func (s *Sessions) Reset(ctx context.Context, userID ids.UserID, chatID int64) error {
	return db.New(s.pool).ResetTelegramSession(ctx, db.ResetTelegramSessionParams{
		UserID: pgconv.UserID(userID),
		ChatID: chatID,
	})
}
