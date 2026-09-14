// Package racingtransform supplies a racingtransformClient
package racingtransform

import (
	"context"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

type racingtransformClient struct{}

// NewRacingTransformClient creates a new Racing transform client
func NewRacingTransformClient() transforms.TransformClient {
	return &racingtransformClient{}
}

var racingTypeMap = map[string]string{
	"horse_racing":     "Horse Racing",
	"greyhound_racing": "Greyhound Racing",
	"harness_racing":   "Harness Racing",
}

// TransformEvent performs racing specific transformation on the Event
func (t *racingtransformClient) TransformEvent(_ context.Context, partialUpdate, fullModel *model.Event) (*model.Event, error) {
	var delta *model.RacingEvent

	if name := racingName(partialUpdate, fullModel); name != "" {
		delta = &model.RacingEvent{Name: &model.OptionalString{Value: name}}
	}

	if fieldSize, changed := fieldSize(partialUpdate, fullModel); changed {
		if delta == nil {
			delta = &model.RacingEvent{}
		}
		delta.FieldSize = &model.OptionalInt64{Value: fieldSize}
	}

	if delta == nil {
		return nil, nil
	}

	return &model.Event{ID: partialUpdate.GetID(), RacingData: delta}, nil
}

// racingName derives the display name of the racing code. Set once, the same
// shape as sporttransform: once a name exists there is nothing to recompute.
func racingName(partialUpdate, fullModel *model.Event) string {
	if partialUpdate.GetEventTypeID() == nil {
		return "" // the racing code did not change on this update
	}

	if fullModel.GetRacingData().GetName() != nil {
		return "" // name already set
	}

	return racingTypeMap[fullModel.GetEventTypeID().GetValue()]
}

// fieldSize counts the runners still engaged in the race. Unlike a set once
// derivation this has to be recomputed whenever runners change, because a late
// scratching reduces it. It returns changed=false when the value would not
// move, so an unchanged race produces no delta and no write.
func fieldSize(partialUpdate, fullModel *model.Event) (int64, bool) {
	if len(partialUpdate.GetRacingData().GetRunners()) == 0 {
		return 0, false // this update did not touch the runners
	}

	var count int64
	for _, runner := range fullModel.GetRacingData().GetRunners() {
		if !runner.GetScratched().GetValue() {
			count++
		}
	}

	existing := fullModel.GetRacingData().GetFieldSize()
	if existing != nil && existing.GetValue() == count {
		return 0, false // already correct
	}

	return count, true
}

func (t *racingtransformClient) GetName() string {
	return "RacingTransform"
}
