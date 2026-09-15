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
