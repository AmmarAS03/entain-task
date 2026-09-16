// Package modeltest provides shared model.Event fixtures for tests that need
// a fully populated event, so the completeness of that fixture is guarded in
// one place rather than duplicated across packages.
package modeltest

import "git.neds.sh/technology/pricekinetics/tools/codetest/model"

// PopulatedSelection returns a Selection with every field set.
func PopulatedSelection(id string) *model.Selection {
	return &model.Selection{
		ID:            id,
		Name:          &model.OptionalString{Value: "Home Team", Deleted: true},
		BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen, Deleted: true},
		Price:         &model.OptionalDouble{Value: 1.80, Deleted: true},
	}
}

// PopulatedMarket returns a Market with every field set.
func PopulatedMarket(id string) *model.Market {
	return &model.Market{
		ID:            id,
		Name:          &model.OptionalString{Value: "Head to Head", Deleted: true},
		StartTime:     &model.OptionalInt64{Value: 1758244443000000000, Deleted: true},
		BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen, Deleted: true},
		Selections:    []*model.Selection{PopulatedSelection("sel-1")},
		ClosedAt:      &model.OptionalInt64{Value: 1758244443000000000, Deleted: true},
	}
}

// PopulatedSportEvent returns a SportEvent with every field set.
func PopulatedSportEvent() *model.SportEvent {
	return &model.SportEvent{
		Name:   &model.OptionalString{Value: "Rugby League", Deleted: true},
		Region: &model.OptionalString{Value: "AU", Deleted: true},
		League: &model.OptionalString{Value: "NRL", Deleted: true},
		Round:  &model.OptionalString{Value: "12", Deleted: true},
	}
}

// PopulatedRunner returns a Runner with every field set.
func PopulatedRunner(id string) *model.Runner {
	return &model.Runner{
		ID:             id,
		Number:         &model.OptionalInt64{Value: 3, Deleted: true},
		Name:           &model.OptionalString{Value: "Winx", Deleted: true},
		Barrier:        &model.OptionalInt64{Value: 4, Deleted: true},
		Weight:         &model.OptionalDouble{Value: 58.5, Deleted: true},
		Jockey:         &model.OptionalString{Value: "H Bowman", Deleted: true},
		Trainer:        &model.OptionalString{Value: "C Waller", Deleted: true},
		Scratched:      &model.OptionalBool{Value: true, Deleted: true},
		Silks:          &model.OptionalString{Value: "navy, white star", Deleted: true},
		FinishPosition: &model.OptionalInt64{Value: 1, Deleted: true},
	}
}

// PopulatedRacingEvent returns a RacingEvent with every field set.
func PopulatedRacingEvent() *model.RacingEvent {
	return &model.RacingEvent{
		Name:           &model.OptionalString{Value: "Horse Racing", Deleted: true},
		Region:         &model.OptionalString{Value: "AU", Deleted: true},
		TrackName:      &model.OptionalString{Value: "Randwick", Deleted: true},
		RaceNumber:     &model.OptionalInt64{Value: 5, Deleted: true},
		DistanceMetres: &model.OptionalInt64{Value: 1600, Deleted: true},
		TrackCondition: &model.OptionalString{Value: "Good 4", Deleted: true},
		Weather:        &model.OptionalString{Value: "Fine", Deleted: true},
		RaceClass:      &model.OptionalString{Value: "Group 1", Deleted: true},
		RaceStatus:     &model.OptionalRaceStatus{Value: model.RaceStatus_RaceScheduled, Deleted: true},
		FieldSize:      &model.OptionalInt64{Value: 12, Deleted: true},
		Runners:        []*model.Runner{PopulatedRunner("1")},
	}
}

// PopulatedEvent returns an Event with every field on every nested message
// set, for tests that need to guard against a field silently failing to
// merge, persist, or convert.
func PopulatedEvent() *model.Event {
	return &model.Event{
		ID:            "evt-1",
		Name:          &model.OptionalString{Value: "Test Event", Deleted: true},
		StartTime:     &model.OptionalInt64{Value: 1758244443000000000, Deleted: true},
		BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen, Deleted: true},
		Markets:       []*model.Market{PopulatedMarket("mkt-1")},
		EventTypeID:   &model.OptionalString{Value: "rugby_league", Deleted: true},
		SportData:     PopulatedSportEvent(),
		Display:       &model.OptionalBool{Value: true, Deleted: true},
		RacingData:    PopulatedRacingEvent(),
	}
}
