// Package core contains the proto definitions of the Core Service
package core

import (
	"sort"
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
// consumer facing messages use. An unset StartTime renders as an empty string
// rather than 1970-01-01: zero means "not set" for a timestamp in this codebase
// (see marketclosetransform's ClosedAt guard), and reporting a date the event
// does not have makes the read RPCs disagree with a SearchEvents date filter,
// which can only match events that really have one.
func formatStartTime(startTime *model.OptionalInt64) string {
	if startTime.GetValue() == 0 {
		return ""
	}
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
	// Unset RaceStatus stays empty, matching how every other racing field on
	// this response behaves for a non-racing event.
	if raceStatus := racing.GetRaceStatus(); raceStatus != nil {
		to.RaceStatus = raceStatus.GetValue().String()
	}
	to.FieldSize = racing.GetFieldSize().GetValue()

	winSelections, placeSelections := runnerSelections(event.GetMarkets())

	runners := racing.GetRunners()
	to.Runners = make([]*RacingRunner, 0, len(runners))
	for _, runner := range runners {
		win := winSelections[runner.GetID()]
		place := placeSelections[runner.GetID()]
		to.Runners = append(to.Runners, &RacingRunner{
			ID:                 runner.GetID(),
			Number:             runner.GetNumber().GetValue(),
			Name:               runner.GetName().GetValue(),
			Barrier:            runner.GetBarrier().GetValue(),
			Weight:             runner.GetWeight().GetValue(),
			Jockey:             runner.GetJockey().GetValue(),
			Trainer:            runner.GetTrainer().GetValue(),
			Scratched:          runner.GetScratched().GetValue(),
			Silks:              runner.GetSilks().GetValue(),
			FinishPosition:     runner.GetFinishPosition().GetValue(),
			WinPrice:           win.GetPrice().GetValue(),
			WinBettingStatus:   win.GetBettingStatus().GetValue().String(),
			PlacePrice:         place.GetPrice().GetValue(),
			PlaceBettingStatus: place.GetBettingStatus().GetValue().String(),
		})
	}

	// Runners render in program order (saddlecloth/box number), not the
	// arbitrary storage order the repository and merger use internally.
	sort.Slice(to.Runners, func(i, j int) bool {
		return to.Runners[i].Number < to.Runners[j].Number
	})
}

// ConvertFromModel converts a model.Event to a core.EventSummary, the shape
// SearchEvents returns. It deliberately carries no sport or racing specific
// fields, so one result slice can hold both kinds of event.
func (to *EventSummary) ConvertFromModel(event *model.Event) {
	to.ID = event.GetID()
	to.Name = event.GetName().GetValue()
	to.StartTime = formatStartTime(event.GetStartTime())
	to.BettingStatus = event.GetBettingStatus().GetValue().String()
	to.EventTypeID = event.GetEventTypeID().GetValue()
	// Unset Display defaults to false (hidden), matching SportEvent and RacingEvent.
	to.Display = event.GetDisplay().GetValue()
}

// runnerSelections indexes the Win and Place selections by Selection ID in a
// single pass over the markets, so ConvertFromModel can join each runner to
// its price and betting status without walking the markets again. Selection
// IDs are runner IDs for racing events.
func runnerSelections(markets []*model.Market) (win, place map[string]*model.Selection) {
	win = map[string]*model.Selection{}
	place = map[string]*model.Selection{}

	for _, market := range markets {
		var target map[string]*model.Selection
		switch market.GetID() {
		case WinMarketID:
			target = win
		case PlaceMarketID:
			target = place
		default:
			continue
		}

		for _, selection := range market.GetSelections() {
			target[selection.GetID()] = selection
		}
	}

	return win, place
}
