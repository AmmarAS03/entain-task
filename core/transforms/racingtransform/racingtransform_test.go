package racingtransform_test

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms/racingtransform"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
)

func runners(scratched ...bool) []*model.Runner {
	out := make([]*model.Runner, 0, len(scratched))
	for i, s := range scratched {
		r := &model.Runner{ID: string(rune('1' + i))}
		if s {
			r.Scratched = &model.OptionalBool{Value: true}
		}
		out = append(out, r)
	}
	return out
}

func TestTransformEvent_SetsRacingName(t *testing.T) {
	client := racingtransform.NewRacingTransformClient()

	partial := &model.Event{ID: "race-1", EventTypeID: &model.OptionalString{Value: "greyhound_racing"}}
	full := &model.Event{ID: "race-1", EventTypeID: &model.OptionalString{Value: "greyhound_racing"}}

	delta, err := client.TransformEvent(context.Background(), partial, full)

	assert.NoError(t, err)
	assert.Equal(t, "Greyhound Racing", delta.GetRacingData().GetName().GetValue())
}

func TestTransformEvent_LeavesSportEventsAlone(t *testing.T) {
	client := racingtransform.NewRacingTransformClient()

	partial := &model.Event{ID: "evt-1", EventTypeID: &model.OptionalString{Value: "rugby_league"}}
	full := &model.Event{ID: "evt-1", EventTypeID: &model.OptionalString{Value: "rugby_league"}}

	delta, err := client.TransformEvent(context.Background(), partial, full)

	assert.NoError(t, err)
	assert.Nil(t, delta, "a sport event should produce no racing delta")
}

func TestTransformEvent_RacingNameIsSetOnce(t *testing.T) {
	client := racingtransform.NewRacingTransformClient()

	partial := &model.Event{ID: "race-1", EventTypeID: &model.OptionalString{Value: "horse_racing"}}
	full := &model.Event{
		ID:          "race-1",
		EventTypeID: &model.OptionalString{Value: "horse_racing"},
		RacingData:  &model.RacingEvent{Name: &model.OptionalString{Value: "Already Named"}},
	}

	delta, err := client.TransformEvent(context.Background(), partial, full)

	assert.NoError(t, err)
	assert.Nil(t, delta, "an existing name should not be recomputed")
}

func TestTransformEvent_FieldSizeCountsUnscratchedRunners(t *testing.T) {
	client := racingtransform.NewRacingTransformClient()

	partial := &model.Event{ID: "race-1", RacingData: &model.RacingEvent{Runners: runners(false)}}
	full := &model.Event{ID: "race-1", RacingData: &model.RacingEvent{Runners: runners(false, false, true, false)}}

	delta, err := client.TransformEvent(context.Background(), partial, full)

	assert.NoError(t, err)
	assert.EqualValues(t, 3, delta.GetRacingData().GetFieldSize().GetValue())
}

// A late scratching has to move FieldSize, which is why this derivation cannot
// use the set once guard that the racing name uses.
func TestTransformEvent_FieldSizeRecomputesOnScratching(t *testing.T) {
	client := racingtransform.NewRacingTransformClient()

	partial := &model.Event{ID: "race-1", RacingData: &model.RacingEvent{Runners: runners(true)}}
	full := &model.Event{
		ID: "race-1",
		RacingData: &model.RacingEvent{
			FieldSize: &model.OptionalInt64{Value: 4},
			Runners:   runners(false, false, true, false),
		},
	}

	delta, err := client.TransformEvent(context.Background(), partial, full)

	assert.NoError(t, err)
	assert.EqualValues(t, 3, delta.GetRacingData().GetFieldSize().GetValue())
}

func TestTransformEvent_NoDeltaWhenFieldSizeUnchanged(t *testing.T) {
	client := racingtransform.NewRacingTransformClient()

	partial := &model.Event{ID: "race-1", RacingData: &model.RacingEvent{Runners: runners(false)}}
	full := &model.Event{
		ID: "race-1",
		RacingData: &model.RacingEvent{
			FieldSize: &model.OptionalInt64{Value: 3},
			Runners:   runners(false, false, true, false),
		},
	}

	delta, err := client.TransformEvent(context.Background(), partial, full)

	assert.NoError(t, err)
	assert.Nil(t, delta, "an unchanged field size should produce no delta")
}

func TestTransformEvent_IgnoresUpdatesThatDoNotTouchRunners(t *testing.T) {
	client := racingtransform.NewRacingTransformClient()

	partial := &model.Event{ID: "race-1", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}}
	full := &model.Event{ID: "race-1", RacingData: &model.RacingEvent{Runners: runners(false, false, false)}}

	delta, err := client.TransformEvent(context.Background(), partial, full)

	assert.NoError(t, err)
	assert.Nil(t, delta, "a price or status update should not recompute the field")
}
