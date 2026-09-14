// Package core contains the proto definitions of the Core Service
package core

import (
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

//go:generate ./gen-proto.sh

// Market IDs that carry per-runner prices. Markets are keyed by a business ID
// in this system (H2H, WIN, PLACE), so these are the racing equivalents of the
// H2H convention already used by sport events.
const (
	WinMarketID   = "WIN"
	PlaceMarketID = "PLACE"
)

// formatStartTime renders an epoch nanosecond timestamp in the form the
// consumer facing messages use.
func formatStartTime(startTime *model.OptionalInt64) string {
	return time.Unix(0, startTime.GetValue()).Format(time.RFC3339)
}

// ConvertFromModel converts a model.Event to a core.SportEvent
func (to *SportEvent) ConvertFromModel(model *model.Event) {
	to.ID = model.ID
	to.Name = model.GetName().GetValue()
	to.StartTime = formatStartTime(model.GetStartTime())
	to.BettingStatus = model.GetBettingStatus().GetValue().String()
	to.SportTypeID = model.GetEventTypeID().GetValue()
	to.Markets = model.Markets
	to.League = model.GetSportData().GetLeague().GetValue()
	to.SportName = model.GetSportData().GetName().GetValue()
	to.Round = model.GetSportData().GetRound().GetValue()
	to.Region = model.GetSportData().GetRegion().GetValue()

	// Unset Display defaults to false (hidden).
	to.Display = model.GetDisplay().GetValue()
}

// ConvertFromModel converts a model.Event to a core.RacingEvent, joining each
// Runner to its current Win and Place price so consumers can render a race card
// without walking the market structure themselves.
func (to *RacingEvent) ConvertFromModel(event *model.Event) {
	to.ID = event.GetID()
	to.Name = event.GetName().GetValue()
	to.StartTime = formatStartTime(event.GetStartTime())
	to.BettingStatus = event.GetBettingStatus().GetValue().String()
	to.RaceTypeID = event.GetEventTypeID().GetValue()
	to.Display = event.GetDisplay().GetValue()

	// Exotics (Exacta, Trifecta and friends) do not fit the per-runner join, so
	// the raw markets stay available alongside it.
	to.Markets = event.GetMarkets()

	racing := event.GetRacingData()
	to.RacingName = racing.GetName().GetValue()
	to.Region = racing.GetRegion().GetValue()
	to.TrackName = racing.GetTrackName().GetValue()
	to.RaceNumber = racing.GetRaceNumber().GetValue()
	to.DistanceMetres = racing.GetDistanceMetres().GetValue()
	to.TrackCondition = racing.GetTrackCondition().GetValue()
	to.Weather = racing.GetWeather().GetValue()
	to.RaceClass = racing.GetRaceClass().GetValue()
	to.RaceStatus = racing.GetRaceStatus().GetValue().String()
	to.FieldSize = racing.GetFieldSize().GetValue()

	winPrices, placePrices := runnerPrices(event.GetMarkets())

	runners := racing.GetRunners()
	to.Runners = make([]*RacingRunner, 0, len(runners))
	for _, runner := range runners {
		to.Runners = append(to.Runners, &RacingRunner{
			ID:             runner.GetID(),
			Number:         runner.GetNumber().GetValue(),
			Name:           runner.GetName().GetValue(),
			Barrier:        runner.GetBarrier().GetValue(),
			Weight:         runner.GetWeight().GetValue(),
			Jockey:         runner.GetJockey().GetValue(),
			Trainer:        runner.GetTrainer().GetValue(),
			Scratched:      runner.GetScratched().GetValue(),
			Silks:          runner.GetSilks().GetValue(),
			FinishPosition: runner.GetFinishPosition().GetValue(),
			WinPrice:       winPrices[runner.GetID()],
			PlacePrice:     placePrices[runner.GetID()],
		})
	}
}

// runnerPrices indexes Win and Place prices by Selection ID in a single pass
// over the markets. Selection IDs are runner IDs for racing events.
func runnerPrices(markets []*model.Market) (win, place map[string]float64) {
	win = map[string]float64{}
	place = map[string]float64{}

	for _, market := range markets {
		var target map[string]float64
		switch market.GetID() {
		case WinMarketID:
			target = win
		case PlaceMarketID:
			target = place
		default:
			continue
		}

		for _, selection := range market.GetSelections() {
			target[selection.GetID()] = selection.GetPrice().GetValue()
		}
	}

	return win, place
}
