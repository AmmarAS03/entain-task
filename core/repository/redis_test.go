package repository

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
)

func Test_redisRepo_UpdateEvent(t *testing.T) {
	repo, err := NewRedisRepository(context.Background(), "localhost:6379", "")
	assert.NoError(t, err)
	input := &model.Event{
		ID:            "e001",
		Name:          &model.OptionalString{Value: "Test Event"},
		StartTime:     &model.OptionalInt64{Value: 1758244443000000000}, // Friday, September 19, 2025 11:14:03 AM GMT+10:00
		BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
		EventTypeID:   &model.OptionalString{Value: "rugby_league"},
		Markets: []*model.Market{
			{
				ID:            "mkt-1",
				Name:          &model.OptionalString{Value: "Head to Head"},
				StartTime:     &model.OptionalInt64{Value: 1758244443000000000}, // Friday, September 19, 2025 11:14:03 AM GMT+10:00
				BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
				Selections: []*model.Selection{
					{
						ID:            "sel-1",
						Name:          &model.OptionalString{Value: "Home Team"},
						BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
						Price:         &model.OptionalDouble{Value: 1.80},
					},
					{
						ID:            "sel-2",
						Name:          &model.OptionalString{Value: "Away Team"},
						BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingSuspended},
						Price:         &model.OptionalDouble{Value: 1.80},
					},
				},
			},
		},
	}

	updErr := repo.UpdateEvent(context.Background(), input)
	assert.NoError(t, updErr)
	output, getErr := repo.GetEventByID(context.Background(), input.ID)
	assert.NoError(t, getErr)
	assert.Equal(t, input.Name.Value, output.Name.Value)
	assert.Equal(t, input.BettingStatus.Value, output.BettingStatus.Value)
	assert.Equal(t, input.Markets[0].Name.Value, output.Markets[0].Name.Value)
	assert.Equal(t, input.Markets[0].Selections[0].Name.Value, output.Markets[0].Selections[0].Name.Value)
	delErr := repo.DeleteEventByID(context.Background(), input.ID)
	assert.NoError(t, delErr)
}
