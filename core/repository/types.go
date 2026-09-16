package repository

import (
	"context"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// Repository is an interface for something that can Retrieve, Update and remove Events from a persistance layer
type Repository interface {
	HealthCheck(ctx context.Context) bool
	GetEventByID(ctx context.Context, id string) (*model.Event, error)
	UpdateEvent(ctx context.Context, event *model.Event) error
	DeleteEventByID(ctx context.Context, id string) error
	Close(ctx context.Context) error
	SearchEvents(ctx context.Context, filter EventFilter) ([]*model.Event, error)
}

// EventFilter describes optional criteria for SearchEvents. A nil field means
// "do not filter on this". Non-nil fields combine with AND.
type EventFilter struct {
	StartTimeFrom *int64 // inclusive, epoch nanoseconds
	StartTimeTo   *int64 // exclusive, epoch nanoseconds
	BettingStatus *model.BettingStatus
	Display       *bool
}
