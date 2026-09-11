package merger

import (
	"context"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

type ServiceClient interface {
	MergeEvent(ctx context.Context, left, right *model.Event) (*model.Event, error)
}

type inlineMergerClient struct {
}

func NewInlineMergerClient() ServiceClient {
	return &inlineMergerClient{}
}

func (c *inlineMergerClient) MergeEvent(ctx context.Context, left, right *model.Event) (*model.Event, error) {
	update := MergeEvent(ctx, left, right)
	return update, nil
}
