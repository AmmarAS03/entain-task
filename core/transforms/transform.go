// Package transforms defines interfaces and common code for transforms
package transforms

import (
	"context"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// TransformClient is an interface representing a service that can perform transformations ion a model
type TransformClient interface {
	TransformEvent(ctx context.Context, parialUpdate, fullModel *model.Event) (*model.Event, error)
	GetName() string
}
