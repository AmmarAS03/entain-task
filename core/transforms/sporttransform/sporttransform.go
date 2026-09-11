// Package sporttransform supplies a sporttransformClient
package sporttransform

import (
	"context"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

type sporttransformClient struct{}

// NewSportTransformClient creates a new Sport transform client
func NewSportTransformClient() transforms.TransformClient {
	return &sporttransformClient{}
}

var sportTypeMap = map[string]string{
	"soccer":       "Soccer",
	"rugby_league": "Rugby League",
}

// TransformEvent performs sport specifc transformation on the Event
func (t *sporttransformClient) TransformEvent(_ context.Context, parialUpdate, fullModel *model.Event) (*model.Event, error) {
	var outDelta *model.Event
	if parialUpdate.EventTypeID == nil {
		return outDelta, nil // if the EventTypeID didnt update on this update skip processing the event
	}

	if fullModel.GetSportData().GetName() != nil {
		return outDelta, nil // no need to update the name if the name already set
	}

	sportName := sportTypeMap[fullModel.GetEventTypeID().GetValue()]
	if sportName == "" {
		return outDelta, nil // unknown sport type so dont change anything
	}

	outDelta = &model.Event{ID: parialUpdate.ID, SportData: &model.SportEvent{Name: &model.OptionalString{Value: sportName}}}

	return outDelta, nil
}

func (t *sporttransformClient) GetName() string {
	return "SportsTransform"
}
