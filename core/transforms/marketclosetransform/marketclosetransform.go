// Package marketclosetransform supplies a marketCloseTransformClient
package marketclosetransform

import (
	"context"
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

type marketCloseTransformClient struct {
	now func() int64
}

// NewMarketCloseTransformClient creates a new MarketClose transform client
func NewMarketCloseTransformClient() transforms.TransformClient {
	return &marketCloseTransformClient{now: func() int64 { return time.Now().UnixNano() }}
}

// TransformEvent stamps ClosedAt on any market that is BettingClosed and does
// not have one yet. It applies to every event type - market closure is not
// sport- or racing-specific.
func (t *marketCloseTransformClient) TransformEvent(_ context.Context, partialUpdate, fullModel *model.Event) (*model.Event, error) {
	if len(partialUpdate.GetMarkets()) == 0 {
		return nil, nil // this update did not touch any market
	}

	var closed []*model.Market
	for _, market := range fullModel.GetMarkets() {
		if market.GetBettingStatus().GetValue() != model.BettingStatus_BettingClosed {
			continue
		}
		if market.GetClosedAt().GetValue() != 0 {
			continue // already stamped; first close wins. Value, not nil-ness, so a client-supplied empty ClosedAt can't poison this guard.
		}
		closed = append(closed, &model.Market{ID: market.GetID(), ClosedAt: &model.OptionalInt64{Value: t.now()}})
	}

	if len(closed) == 0 {
		return nil, nil
	}

	return &model.Event{ID: partialUpdate.GetID(), Markets: closed}, nil
}

func (t *marketCloseTransformClient) GetName() string {
	return "MarketCloseTransform"
}
