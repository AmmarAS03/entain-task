// Package core contains the proto definitions of the Core Service
package core

import (
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

//go:generate ./gen-proto.sh

// ConvertFromModel converts a model.Event to a core.SportEvent
func (to *SportEvent) ConvertFromModel(model *model.Event) {
	to.ID = model.ID
	to.Name = model.GetName().GetValue()
	to.StartTime = time.Unix(0, model.StartTime.GetValue()).Format(time.RFC3339)
	to.BettingStatus = model.GetBettingStatus().GetValue().String()
	to.SportTypeID = model.GetEventTypeID().GetValue()
	to.Markets = model.Markets
	to.League = model.GetSportData().GetLeague().GetValue()
	to.SportName = model.GetSportData().GetName().GetValue()
	to.Round = model.GetSportData().GetRound().GetValue()
	to.Region = model.GetSportData().GetRegion().GetValue()
}
