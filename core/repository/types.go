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
}
