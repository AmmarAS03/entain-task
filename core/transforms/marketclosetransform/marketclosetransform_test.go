package marketclosetransform

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
)

func fixedClock(t int64) func() int64 {
	return func() int64 { return t }
}

func TestTransformEvent_StampsClosedMarket(t *testing.T) {
	client := &marketCloseTransformClient{now: fixedClock(100)}

	partial := &model.Event{ID: "evt-1", Markets: []*model.Market{{ID: "mkt-1"}}}
	fullModel := &model.Event{
		ID: "evt-1",
		Markets: []*model.Market{
			{ID: "mkt-1", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}},
		},
	}

	delta, err := client.TransformEvent(context.Background(), partial, fullModel)

	assert.NoError(t, err)
	assert.NotNil(t, delta)
	assert.Equal(t, "evt-1", delta.GetID())
	assert.Len(t, delta.GetMarkets(), 1)
	assert.Equal(t, "mkt-1", delta.GetMarkets()[0].GetID())
	assert.EqualValues(t, 100, delta.GetMarkets()[0].GetClosedAt().GetValue())
}

func TestTransformEvent_NoDeltaWhenPartialDidNotTouchMarkets(t *testing.T) {
	client := &marketCloseTransformClient{now: fixedClock(100)}

	partial := &model.Event{ID: "evt-1", Name: &model.OptionalString{Value: "renamed"}}
	fullModel := &model.Event{
		ID: "evt-1",
		Markets: []*model.Market{
			{ID: "mkt-1", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}},
		},
	}

	delta, err := client.TransformEvent(context.Background(), partial, fullModel)

	assert.NoError(t, err)
	assert.Nil(t, delta)
}

func TestTransformEvent_NoDeltaWhenMarketNotClosed(t *testing.T) {
	client := &marketCloseTransformClient{now: fixedClock(100)}

	for _, status := range []model.BettingStatus{model.BettingStatus_BettingOpen, model.BettingStatus_BettingSuspended} {
		partial := &model.Event{ID: "evt-1", Markets: []*model.Market{{ID: "mkt-1"}}}
		fullModel := &model.Event{
			ID:      "evt-1",
			Markets: []*model.Market{{ID: "mkt-1", BettingStatus: &model.OptionalBettingStatus{Value: status}}},
		}

		delta, err := client.TransformEvent(context.Background(), partial, fullModel)

		assert.NoError(t, err)
		assert.Nil(t, delta, "status %v should not produce a delta", status)
	}
}

func TestTransformEvent_NoDeltaWhenAlreadyStamped(t *testing.T) {
	client := &marketCloseTransformClient{now: fixedClock(200)}

	partial := &model.Event{ID: "evt-1", Markets: []*model.Market{{ID: "mkt-1"}}}
	fullModel := &model.Event{
		ID: "evt-1",
		Markets: []*model.Market{
			{
				ID:            "mkt-1",
				BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed},
				ClosedAt:      &model.OptionalInt64{Value: 100},
			},
		},
	}

	delta, err := client.TransformEvent(context.Background(), partial, fullModel)

	assert.NoError(t, err)
	assert.Nil(t, delta)
}

// TestTransformEvent_FrozenAcrossReopenAndReclose drives the client through a
// genuine close -> reopen -> re-close sequence, merging each delta back in the
// way Service.Update does, rather than asserting on a single hand-built
// fullModel - that would just re-test the already-stamped guard.
func TestTransformEvent_FrozenAcrossReopenAndReclose(t *testing.T) {
	client := &marketCloseTransformClient{now: fixedClock(100)}

	fullModel := &model.Event{
		ID:      "evt-1",
		Markets: []*model.Market{{ID: "mkt-1", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}}},
	}
	partial := &model.Event{ID: "evt-1", Markets: []*model.Market{{ID: "mkt-1"}}}

	delta, err := client.TransformEvent(context.Background(), partial, fullModel)
	assert.NoError(t, err)
	assert.NotNil(t, delta, "first close should stamp ClosedAt")
	firstClose := delta.GetMarkets()[0].GetClosedAt().GetValue()
	assert.EqualValues(t, 100, firstClose)

	fullModel.Markets[0].ClosedAt = &model.OptionalInt64{Value: firstClose}
	fullModel.Markets[0].BettingStatus = &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen}

	client.now = fixedClock(999)
	fullModel.Markets[0].BettingStatus = &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}

	delta, err = client.TransformEvent(context.Background(), partial, fullModel)
	assert.NoError(t, err)
	assert.Nil(t, delta, "re-closing after a reopen must not restamp ClosedAt")
}

// TestTransformEvent_ClientSuppliedEmptyClosedAtDoesNotPoisonGuard guards
// against a client on the public Update RPC setting an empty ClosedAt
// directly (which unmarshals to a non-nil, zero-value message): the guard
// must key off Value, not pointer nil-ness, or the market can never be
// stamped again.
func TestTransformEvent_ClientSuppliedEmptyClosedAtDoesNotPoisonGuard(t *testing.T) {
	client := &marketCloseTransformClient{now: fixedClock(100)}

	partial := &model.Event{ID: "evt-1", Markets: []*model.Market{{ID: "mkt-1"}}}
	fullModel := &model.Event{
		ID: "evt-1",
		Markets: []*model.Market{
			{
				ID:            "mkt-1",
				BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed},
				ClosedAt:      &model.OptionalInt64{}, // non-nil, zero Value - as if echoed back empty by a client
			},
		},
	}

	delta, err := client.TransformEvent(context.Background(), partial, fullModel)

	assert.NoError(t, err)
	assert.NotNil(t, delta, "a zero-value ClosedAt must still be treated as unstamped")
	assert.EqualValues(t, 100, delta.GetMarkets()[0].GetClosedAt().GetValue())
}

func TestTransformEvent_OnlyClosedUnstampedMarketsInDelta(t *testing.T) {
	client := &marketCloseTransformClient{now: fixedClock(500)}

	partial := &model.Event{
		ID: "evt-1",
		Markets: []*model.Market{
			{ID: "mkt-open"}, {ID: "mkt-closed-new"}, {ID: "mkt-closed-stamped"}, {ID: "mkt-suspended"},
		},
	}
	fullModel := &model.Event{
		ID: "evt-1",
		Markets: []*model.Market{
			{ID: "mkt-open", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen}},
			{ID: "mkt-closed-new", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}},
			{
				ID:            "mkt-closed-stamped",
				BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed},
				ClosedAt:      &model.OptionalInt64{Value: 1},
			},
			{ID: "mkt-suspended", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingSuspended}},
		},
	}

	delta, err := client.TransformEvent(context.Background(), partial, fullModel)

	assert.NoError(t, err)
	assert.NotNil(t, delta)
	assert.Len(t, delta.GetMarkets(), 1)
	assert.Equal(t, "mkt-closed-new", delta.GetMarkets()[0].GetID())
	assert.EqualValues(t, 500, delta.GetMarkets()[0].GetClosedAt().GetValue())
}

func TestGetName(t *testing.T) {
	client := NewMarketCloseTransformClient()
	assert.Equal(t, "MarketCloseTransform", client.GetName())
}
