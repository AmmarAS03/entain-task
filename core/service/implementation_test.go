package service_test

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/repository"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/service"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms/marketclosetransform"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms/racingtransform"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms/sporttransform"
	"git.neds.sh/technology/pricekinetics/tools/codetest/merger"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
)

func TestService_IntegrationTest_NewEvent(t *testing.T) {
	repo, err := repository.NewRedisRepository(context.Background(), "localhost:6379", "")
	assert.NoError(t, err)
	defer repo.DeleteEventByID(context.Background(), "integration-test-1")
	host := &service.Service{
		Upstreams: &service.Upstreams{
			MergerClient: merger.NewInlineMergerClient(),
			Repo:         repo,
			Transforms: []transforms.TransformClient{
				sporttransform.NewSportTransformClient(),
				marketclosetransform.NewMarketCloseTransformClient(),
			},
		},
	}

	output, err := host.Update(context.Background(), &core.UpdateRequest{
		Event: &model.Event{
			ID:          "integration-test-1",
			Name:        &model.OptionalString{Value: "Test event"},
			EventTypeID: &model.OptionalString{Value: "soccer"},
			StartTime:   &model.OptionalInt64{Value: 1758244443000000000}, // Friday, September 19, 2025 11:14:03 AM GMT+10:00
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "New Event born integration-test-1", output.Message)

	output, err = host.Update(context.Background(), &core.UpdateRequest{
		Event: &model.Event{
			ID: "integration-test-1",
			Markets: []*model.Market{
				{
					ID:   "mkt01",
					Name: &model.OptionalString{Value: "New Market"},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "Success", output.Message)

	final, err := host.GetSportEvent(context.Background(), &core.GetSportEventRequest{EventID: "integration-test-1"})
	assert.NoError(t, err)
	assert.Equal(t, "Test event", final.Event.Name)
	assert.Contains(t, final.Event.StartTime, "2025-09-19")
	assert.Equal(t, "soccer", final.Event.SportTypeID)
	assert.Equal(t, "Soccer", final.Event.SportName)
	assert.Equal(t, "New Market", final.Event.Markets[0].Name.Value)
}

// TestService_IntegrationTest_Display covers Display end to end: unset, hidden,
// unaffected by an unrelated update, then shown again.
func TestService_IntegrationTest_Display(t *testing.T) {
	repo, err := repository.NewRedisRepository(context.Background(), "localhost:6379", "")
	assert.NoError(t, err)
	defer repo.DeleteEventByID(context.Background(), "integration-test-display")

	host := &service.Service{
		Upstreams: &service.Upstreams{
			MergerClient: merger.NewInlineMergerClient(),
			Repo:         repo,
			Transforms: []transforms.TransformClient{
				sporttransform.NewSportTransformClient(),
				marketclosetransform.NewMarketCloseTransformClient(),
			},
		},
	}

	update := func(event *model.Event) {
		t.Helper()
		_, uErr := host.Update(context.Background(), &core.UpdateRequest{Event: event})
		assert.NoError(t, uErr)
	}

	displayOf := func() bool {
		t.Helper()
		got, gErr := host.GetSportEvent(context.Background(), &core.GetSportEventRequest{EventID: "integration-test-display"})
		assert.NoError(t, gErr)
		return got.Event.Display
	}

	// Unset defaults to hidden.
	update(&model.Event{
		ID:   "integration-test-display",
		Name: &model.OptionalString{Value: "Test event"},
	})
	assert.False(t, displayOf())

	// Shown.
	update(&model.Event{
		ID:      "integration-test-display",
		Display: &model.OptionalBool{Value: true},
	})
	assert.True(t, displayOf())

	// Unrelated update, Display unaffected.
	update(&model.Event{
		ID:      "integration-test-display",
		Markets: []*model.Market{{ID: "mkt01", Name: &model.OptionalString{Value: "New Market"}}},
	})
	assert.True(t, displayOf())

	// Hidden again.
	update(&model.Event{
		ID:      "integration-test-display",
		Display: &model.OptionalBool{Value: false},
	})
	assert.False(t, displayOf())
}

// TestService_IntegrationTest_RacingEvent covers the racing path end to end:
// runners joined to their Win and Place prices, FieldSize derived by the
// transform, and a late scratching moving it.
func TestService_IntegrationTest_RacingEvent(t *testing.T) {
	repo, err := repository.NewRedisRepository(context.Background(), "localhost:6379", "")
	assert.NoError(t, err)
	defer repo.DeleteEventByID(context.Background(), "integration-test-racing")

	host := &service.Service{
		Upstreams: &service.Upstreams{
			MergerClient: merger.NewInlineMergerClient(),
			Repo:         repo,
			Transforms: []transforms.TransformClient{
				sporttransform.NewSportTransformClient(),
				racingtransform.NewRacingTransformClient(),
				marketclosetransform.NewMarketCloseTransformClient(),
			},
		},
	}

	update := func(event *model.Event) {
		t.Helper()
		_, uErr := host.Update(context.Background(), &core.UpdateRequest{Event: event})
		assert.NoError(t, uErr)
	}

	racingEvent := func() *core.RacingEvent {
		t.Helper()
		got, gErr := host.GetRacingEvent(context.Background(), &core.GetRacingEventRequest{EventID: "integration-test-racing"})
		assert.NoError(t, gErr)
		return got.GetEvent()
	}

	update(&model.Event{
		ID:          "integration-test-racing",
		Name:        &model.OptionalString{Value: "Randwick Race 5"},
		StartTime:   &model.OptionalInt64{Value: 1758244443000000000},
		EventTypeID: &model.OptionalString{Value: "horse_racing"},
		RacingData: &model.RacingEvent{
			TrackName:      &model.OptionalString{Value: "Randwick"},
			RaceNumber:     &model.OptionalInt64{Value: 5},
			DistanceMetres: &model.OptionalInt64{Value: 1600},
			TrackCondition: &model.OptionalString{Value: "Good 4"},
			RaceStatus:     &model.OptionalRaceStatus{Value: model.RaceStatus_RaceScheduled},
			Runners: []*model.Runner{
				{ID: "1", Number: &model.OptionalInt64{Value: 1}, Name: &model.OptionalString{Value: "Winx"}, Barrier: &model.OptionalInt64{Value: 4}, Jockey: &model.OptionalString{Value: "H Bowman"}},
				{ID: "2", Number: &model.OptionalInt64{Value: 2}, Name: &model.OptionalString{Value: "Black Caviar"}, Barrier: &model.OptionalInt64{Value: 7}, Jockey: &model.OptionalString{Value: "L Nolen"}},
				{ID: "3", Number: &model.OptionalInt64{Value: 3}, Name: &model.OptionalString{Value: "Phar Lap"}, Barrier: &model.OptionalInt64{Value: 1}, Jockey: &model.OptionalString{Value: "J Pike"}},
			},
		},
		Markets: []*model.Market{
			{
				ID:   core.WinMarketID,
				Name: &model.OptionalString{Value: "Win"},
				Selections: []*model.Selection{
					{ID: "1", Price: &model.OptionalDouble{Value: 2.40}},
					{ID: "2", Price: &model.OptionalDouble{Value: 3.10}},
					{ID: "3", Price: &model.OptionalDouble{Value: 8.00}},
				},
			},
			{
				ID:   core.PlaceMarketID,
				Name: &model.OptionalString{Value: "Place"},
				Selections: []*model.Selection{
					{ID: "1", Price: &model.OptionalDouble{Value: 1.30}},
					{ID: "2", Price: &model.OptionalDouble{Value: 1.55}},
					{ID: "3", Price: &model.OptionalDouble{Value: 2.20}},
				},
			},
		},
	})

	got := racingEvent()
	assert.Equal(t, "Randwick", got.TrackName)
	assert.EqualValues(t, 5, got.RaceNumber)
	assert.EqualValues(t, 1600, got.DistanceMetres)
	assert.Equal(t, "Good 4", got.TrackCondition)
	assert.Equal(t, "RaceScheduled", got.RaceStatus)
	assert.Equal(t, "horse_racing", got.RaceTypeID)
	assert.Equal(t, "Horse Racing", got.RacingName, "racingtransform should derive the code name")
	assert.EqualValues(t, 3, got.FieldSize, "three runners, none scratched")

	assert.Len(t, got.Runners, 3)
	assert.Equal(t, "Winx", got.Runners[0].Name)
	assert.EqualValues(t, 4, got.Runners[0].Barrier)
	assert.Equal(t, "H Bowman", got.Runners[0].Jockey)
	assert.InDelta(t, 2.40, got.Runners[0].WinPrice, 0.001, "win price joined from the WIN market")
	assert.InDelta(t, 1.30, got.Runners[0].PlacePrice, 0.001, "place price joined from the PLACE market")

	// A price move touches only the market, never the runner metadata.
	update(&model.Event{
		ID: "integration-test-racing",
		Markets: []*model.Market{
			{ID: core.WinMarketID, Selections: []*model.Selection{{ID: "1", Price: &model.OptionalDouble{Value: 2.10}}}},
		},
	})
	got = racingEvent()
	assert.InDelta(t, 2.10, got.Runners[0].WinPrice, 0.001, "new win price")
	assert.Equal(t, "Winx", got.Runners[0].Name, "runner metadata survives a price update")
	assert.EqualValues(t, 3, got.FieldSize, "a price move must not change the field size")

	// A scratching names one runner and moves the field size.
	update(&model.Event{
		ID: "integration-test-racing",
		RacingData: &model.RacingEvent{
			Runners: []*model.Runner{{ID: "2", Scratched: &model.OptionalBool{Value: true}}},
		},
	})
	got = racingEvent()
	assert.EqualValues(t, 2, got.FieldSize, "scratching one of three leaves two")
	assert.Len(t, got.Runners, 3, "a scratched runner stays on the card")
	assert.True(t, got.Runners[1].Scratched)
	assert.Equal(t, "Black Caviar", got.Runners[1].Name, "scratching must not drop the name")
}

// A racing event fetched through the sport RPC must not report racing data as
// though it were sport data, and vice versa.
func TestService_IntegrationTest_RacingAndSportViewsAreSeparate(t *testing.T) {
	repo, err := repository.NewRedisRepository(context.Background(), "localhost:6379", "")
	assert.NoError(t, err)
	defer repo.DeleteEventByID(context.Background(), "integration-test-separation")

	host := &service.Service{
		Upstreams: &service.Upstreams{
			MergerClient: merger.NewInlineMergerClient(),
			Repo:         repo,
			Transforms: []transforms.TransformClient{
				sporttransform.NewSportTransformClient(),
				racingtransform.NewRacingTransformClient(),
				marketclosetransform.NewMarketCloseTransformClient(),
			},
		},
	}

	_, err = host.Update(context.Background(), &core.UpdateRequest{Event: &model.Event{
		ID:          "integration-test-separation",
		Name:        &model.OptionalString{Value: "Randwick Race 5"},
		EventTypeID: &model.OptionalString{Value: "horse_racing"},
		RacingData: &model.RacingEvent{
			TrackName: &model.OptionalString{Value: "Randwick"},
			Runners:   []*model.Runner{{ID: "1", Name: &model.OptionalString{Value: "Winx"}}},
		},
	}})
	assert.NoError(t, err)

	sportView, err := host.GetSportEvent(context.Background(), &core.GetSportEventRequest{EventID: "integration-test-separation"})
	assert.NoError(t, err)
	assert.Equal(t, "Randwick Race 5", sportView.Event.Name, "shared event fields still resolve")
	assert.Empty(t, sportView.Event.SportName, "racing data must not leak into the sport view")
	assert.Empty(t, sportView.Event.League)
	assert.Empty(t, sportView.Event.Round)

	racingView, err := host.GetRacingEvent(context.Background(), &core.GetRacingEventRequest{EventID: "integration-test-separation"})
	assert.NoError(t, err)
	assert.Equal(t, "Randwick", racingView.Event.TrackName)
	assert.Len(t, racingView.Event.Runners, 1)
}

// TestService_IntegrationTest_MarketClose covers ClosedAt end to end: stamped
// on first close, unaffected by an unrelated update, and frozen across a
// reopen and re-close.
func TestService_IntegrationTest_MarketClose(t *testing.T) {
	repo, err := repository.NewRedisRepository(context.Background(), "localhost:6379", "")
	assert.NoError(t, err)
	defer repo.DeleteEventByID(context.Background(), "integration-test-close")

	host := &service.Service{
		Upstreams: &service.Upstreams{
			MergerClient: merger.NewInlineMergerClient(),
			Repo:         repo,
			Transforms: []transforms.TransformClient{
				sporttransform.NewSportTransformClient(),
				marketclosetransform.NewMarketCloseTransformClient(),
			},
		},
	}

	update := func(event *model.Event) {
		t.Helper()
		_, uErr := host.Update(context.Background(), &core.UpdateRequest{Event: event})
		assert.NoError(t, uErr)
	}

	marketOf := func() *model.Market {
		t.Helper()
		got, gErr := host.GetSportEvent(context.Background(), &core.GetSportEventRequest{EventID: "integration-test-close"})
		assert.NoError(t, gErr)
		return got.Event.Markets[0]
	}

	update(&model.Event{
		ID:   "integration-test-close",
		Name: &model.OptionalString{Value: "Test event"},
		Markets: []*model.Market{
			{
				ID:            "mkt01",
				Name:          &model.OptionalString{Value: "Head to Head"},
				BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
				Selections: []*model.Selection{
					{ID: "home", Price: &model.OptionalDouble{Value: 1.80}},
					{ID: "away", Price: &model.OptionalDouble{Value: 2.10}},
				},
			},
		},
	})
	assert.Nil(t, marketOf().ClosedAt, "an open market has no ClosedAt")

	update(&model.Event{
		ID:      "integration-test-close",
		Markets: []*model.Market{{ID: "mkt01", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}}},
	})
	closedMarket := marketOf()
	firstClose := closedMarket.GetClosedAt().GetValue()
	assert.NotZero(t, firstClose, "closing the market stamps ClosedAt")
	assert.Equal(t, "Head to Head", closedMarket.GetName().GetValue(), "the close delta must not clobber Name")
	assert.Len(t, closedMarket.GetSelections(), 2, "the close delta must not clobber Selections")

	// An unrelated market update must not move the timestamp.
	update(&model.Event{
		ID:      "integration-test-close",
		Markets: []*model.Market{{ID: "mkt01", Name: &model.OptionalString{Value: "Head to Head Renamed"}}},
	})
	assert.Equal(t, firstClose, marketOf().GetClosedAt().GetValue(), "ClosedAt must not move on an unrelated market update")

	// Reopen then re-close: ClosedAt stays frozen at the first close.
	update(&model.Event{
		ID:      "integration-test-close",
		Markets: []*model.Market{{ID: "mkt01", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen}}},
	})
	update(&model.Event{
		ID:      "integration-test-close",
		Markets: []*model.Market{{ID: "mkt01", BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed}}},
	})
	assert.Equal(t, firstClose, marketOf().GetClosedAt().GetValue(), "ClosedAt freezes at the first close")
}
