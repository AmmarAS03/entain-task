package core_test

import (
	"fmt"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
)

// A field of 10+ runners sorts "10" before "2" under a lexicographic ID sort.
// The response must order by saddlecloth/box Number instead, since that is
// what a race card renders by.
func TestRacingEvent_ConvertFromModel_OrdersRunnersByNumber(t *testing.T) {
	// Insertion order deliberately matches neither Number order nor
	// lexicographic ID order, so the assertion below only passes if
	// ConvertFromModel actually sorts by Number.
	insertionOrder := []int64{7, 11, 2, 1, 10, 9, 3, 8, 4, 6, 5}
	var runners []*model.Runner
	for _, n := range insertionOrder {
		runners = append(runners, &model.Runner{
			ID:     fmt.Sprintf("%d", n),
			Number: &model.OptionalInt64{Value: n},
		})
	}

	event := &model.Event{
		ID:         "race-1",
		RacingData: &model.RacingEvent{Runners: runners},
	}

	out := &core.RacingEvent{}
	out.ConvertFromModel(event)

	assert.Len(t, out.Runners, 11)
	for i, runner := range out.Runners {
		assert.EqualValues(t, i+1, runner.Number, "runner at index %d should be number %d", i, i+1)
	}
}

// A scratched runner's price join must not look like a live, tradeable price -
// the response needs to carry the selection's BettingStatus alongside it.
func TestRacingEvent_ConvertFromModel_CarriesBettingStatusForScratchedRunner(t *testing.T) {
	event := &model.Event{
		ID: "race-1",
		RacingData: &model.RacingEvent{
			Runners: []*model.Runner{
				{ID: "1", Scratched: &model.OptionalBool{Value: true}},
			},
		},
		Markets: []*model.Market{
			{
				ID: core.WinMarketID,
				Selections: []*model.Selection{
					{
						ID:            "1",
						Price:         &model.OptionalDouble{Value: 12.0},
						BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingSuspended},
					},
				},
			},
		},
	}

	out := &core.RacingEvent{}
	out.ConvertFromModel(event)

	assert.Len(t, out.Runners, 1)
	assert.True(t, out.Runners[0].Scratched)
	assert.InDelta(t, 12.0, out.Runners[0].WinPrice, 0.001, "the stale price is still surfaced")
	assert.Equal(t, "BettingSuspended", out.Runners[0].WinBettingStatus, "a consumer can tell the price is not live")
}

// GetRacingEvent on a non-racing event should report every racing field as
// empty, not just most of them.
func TestRacingEvent_ConvertFromModel_RaceStatusEmptyOnSportEvent(t *testing.T) {
	event := &model.Event{
		ID:        "evt-1",
		SportData: &model.SportEvent{Name: &model.OptionalString{Value: "Soccer"}},
	}

	out := &core.RacingEvent{}
	out.ConvertFromModel(event)

	assert.Empty(t, out.RaceStatus)
}

// A fully populated event should render every EventSummary field as its flat
// consumer-facing type, not the model's Optional* wrapper.
func TestEventSummary_ConvertFromModel_FullyPopulated(t *testing.T) {
	event := &model.Event{
		ID:            "evt-1",
		Name:          &model.OptionalString{Value: "Test Event"},
		StartTime:     &model.OptionalInt64{Value: 1758244443000000000},
		BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
		EventTypeID:   &model.OptionalString{Value: "soccer"},
		Display:       &model.OptionalBool{Value: true},
	}

	out := &core.EventSummary{}
	out.ConvertFromModel(event)

	assert.Equal(t, "evt-1", out.ID)
	assert.Equal(t, "Test Event", out.Name)
	assert.Contains(t, out.StartTime, "2025-09-19")
	assert.Equal(t, "BettingOpen", out.BettingStatus)
	assert.Equal(t, "soccer", out.EventTypeID)
	assert.True(t, out.Display)
}

// A sparse event (nil Display, nil StartTime, nil BettingStatus) should
// convert to zero-value defaults rather than panicking - in particular,
// unset Display must default to false (hidden), matching Task 1's semantics
// for SportEvent and RacingEvent, and unset StartTime must render as an empty
// string rather than the 1970-01-01 formatStartTime used to produce (zero
// means "not set" for a timestamp in this codebase - see
// marketclosetransform's ClosedAt guard - and a 1970 date the event does not
// really have would disagree with what a SearchEvents date filter can match).
func TestEventSummary_ConvertFromModel_Sparse(t *testing.T) {
	event := &model.Event{ID: "evt-2"}

	out := &core.EventSummary{}
	out.ConvertFromModel(event)

	assert.Equal(t, "evt-2", out.ID)
	assert.Empty(t, out.Name)
	assert.Empty(t, out.StartTime, "unset StartTime must not render as 1970-01-01")
	assert.Equal(t, "BettingUnknown", out.BettingStatus, "unset BettingStatus renders its zero enum value, same as SportEvent/RacingEvent")
	assert.Empty(t, out.EventTypeID)
	assert.False(t, out.Display)
}

// SportEvent and RacingEvent share formatStartTime with EventSummary, so an
// unset StartTime must render as an empty string on all three views, not just
// the one that happened to get tested first.
func TestSportEvent_ConvertFromModel_UnsetStartTimeIsEmpty(t *testing.T) {
	event := &model.Event{ID: "evt-3", SportData: &model.SportEvent{Name: &model.OptionalString{Value: "Soccer"}}}

	out := &core.SportEvent{}
	out.ConvertFromModel(event)

	assert.Empty(t, out.StartTime, "unset StartTime must not render as 1970-01-01")
}

func TestRacingEvent_ConvertFromModel_UnsetStartTimeIsEmpty(t *testing.T) {
	event := &model.Event{ID: "evt-4", RacingData: &model.RacingEvent{TrackName: &model.OptionalString{Value: "Randwick"}}}

	out := &core.RacingEvent{}
	out.ConvertFromModel(event)

	assert.Empty(t, out.StartTime, "unset StartTime must not render as 1970-01-01")
}
