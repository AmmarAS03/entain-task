package service_test

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/repository"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/service"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
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
