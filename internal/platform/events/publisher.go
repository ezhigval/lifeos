package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
)

type Source string

const (
	SourceTelegram  Source = "telegram"
	SourceScheduler Source = "scheduler"
	SourceCLI       Source = "cli"
	SourceHTTP      Source = "http"
)

type Record struct {
	UserID        ids.UserID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       any
	Source        Source
	OccurredAt    time.Time
}

type Publisher struct {
	pool *pgxpool.Pool
}

func NewPublisher(pool *pgxpool.Pool) *Publisher {
	return &Publisher{pool: pool}
}

func (p *Publisher) Append(ctx context.Context, rec Record) error {
	payload, err := json.Marshal(rec.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	return p.queries(ctx).InsertDomainEvent(ctx, db.InsertDomainEventParams{
		ID:            pgconv.UUID(uuid.Must(uuid.NewV7())),
		UserID:        pgconv.UserID(rec.UserID),
		AggregateType: rec.AggregateType,
		AggregateID:   pgconv.UUID(rec.AggregateID),
		EventType:     rec.EventType,
		Payload:       payload,
		Source:        string(rec.Source),
		OccurredAt:    pgconv.Timestamptz(&rec.OccurredAt),
	})
}

func (p *Publisher) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(p.pool)
}
