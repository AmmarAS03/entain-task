package repository_test

import (
	"context"
	"testing"
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/repository"
	"git.neds.sh/technology/pricekinetics/tools/codetest/internal/modeltest"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_mongoRepo_UpdateEvent(t *testing.T) {
	repo, err := repository.NewMongoRepository(context.Background(), "mongodb://localhost:27017", "codetest")
	require.NoError(t, err)
	defer repo.Close(context.Background())

	in := modeltest.PopulatedEvent()
	defer repo.DeleteEventByID(context.Background(), in.ID)
	require.NoError(t, repo.UpdateEvent(context.Background(), in))

	out, err := repo.GetEventByID(context.Background(), in.ID)
	require.NoError(t, err)
	assert.True(t, proto.Equal(in, out), "round trip lost a field:\n want %v\n got  %v", in, out)
}

func Test_mongoRepo_GetEventByID_NotFound(t *testing.T) {
	repo, err := repository.NewMongoRepository(context.Background(), "mongodb://localhost:27017", "codetest")
	require.NoError(t, err)
	defer repo.Close(context.Background())

	out, err := repo.GetEventByID(context.Background(), "does-not-exist")
	assert.NoError(t, err)
	assert.Nil(t, out)
}

// Search tests use their own database rather than the shared "codetest" dev
// database. Unlike every other test in this file, which fetches by a unique
// ID, SearchEvents scans the whole collection - a stray document left behind
// by ./run_local.sh would make an assertion like "this filter returns 2
// events" fail intermittently on whoever happens to run it next.
const searchTestDatabase = "codetest_search_test"

func newSearchTestRepo(t *testing.T) repository.Repository {
	t.Helper()
	repo, err := repository.NewMongoRepository(context.Background(), "mongodb://localhost:27017", searchTestDatabase)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close(context.Background()) })
	return repo
}

func seedSearchEvent(t *testing.T, repo repository.Repository, event *model.Event) {
	t.Helper()
	require.NoError(t, repo.UpdateEvent(context.Background(), event))
	t.Cleanup(func() { repo.DeleteEventByID(context.Background(), event.ID) })
}

// searchFixture returns four events spread across four start times an hour
// apart, with different betting statuses and display states, so a single
// seeded set can exercise every filter combination below.
//
//	ID     StartTime  BettingStatus     Display
//	evt-A  T+0h        BettingOpen      true
//	evt-B  T+1h        BettingClosed    false (explicit)
//	evt-C  T+2h        BettingOpen      unset (never set)
//	evt-D  T+3h        BettingSuspended true
func searchFixture(baseTime int64) []*model.Event {
	hour := int64(time.Hour)
	return []*model.Event{
		{
			ID:            "evt-A",
			StartTime:     &model.OptionalInt64{Value: baseTime},
			BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
			Display:       &model.OptionalBool{Value: true},
		},
		{
			ID:            "evt-B",
			StartTime:     &model.OptionalInt64{Value: baseTime + hour},
			BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingClosed},
			Display:       &model.OptionalBool{Value: false},
		},
		{
			ID:            "evt-C",
			StartTime:     &model.OptionalInt64{Value: baseTime + 2*hour},
			BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
			// Display deliberately left unset.
		},
		{
			ID:            "evt-D",
			StartTime:     &model.OptionalInt64{Value: baseTime + 3*hour},
			BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingSuspended},
			Display:       &model.OptionalBool{Value: true},
		},
	}
}

func searchIDs(events []*model.Event) []string {
	ids := make([]string, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.GetID())
	}
	return ids
}

func Test_mongoRepo_SearchEvents(t *testing.T) {
	const baseTime = int64(1758244443000000000) // Friday, September 19, 2025 11:14:03 AM GMT+10:00
	hour := int64(time.Hour)

	// Seeded out of chronological order (D, A, C, B) so the ordering
	// assertions below only pass if SearchEvents actually sorts, rather than
	// returning Mongo's natural (insertion) order.
	seedOrder := []int{3, 0, 2, 1}

	seed := func(t *testing.T) repository.Repository {
		t.Helper()
		repo := newSearchTestRepo(t)
		fixture := searchFixture(baseTime)
		for _, i := range seedOrder {
			seedSearchEvent(t, repo, fixture[i])
		}
		return repo
	}

	t.Run("no filters returns every seeded event ascending by start time", func(t *testing.T) {
		repo := seed(t)
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{})
		require.NoError(t, err)
		assert.Equal(t, []string{"evt-A", "evt-B", "evt-C", "evt-D"}, searchIDs(got))
	})

	t.Run("StartTimeFrom only is inclusive of the boundary event", func(t *testing.T) {
		repo := seed(t)
		from := baseTime + hour
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{StartTimeFrom: &from})
		require.NoError(t, err)
		assert.Equal(t, []string{"evt-B", "evt-C", "evt-D"}, searchIDs(got))
	})

	t.Run("StartTimeTo only is exclusive of the boundary event", func(t *testing.T) {
		repo := seed(t)
		to := baseTime + 2*hour
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{StartTimeTo: &to})
		require.NoError(t, err)
		assert.Equal(t, []string{"evt-A", "evt-B"}, searchIDs(got))
	})

	t.Run("both bounds return only events strictly inside the window", func(t *testing.T) {
		repo := seed(t)
		from := baseTime + hour
		to := baseTime + 3*hour
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{StartTimeFrom: &from, StartTimeTo: &to})
		require.NoError(t, err)
		assert.Equal(t, []string{"evt-B", "evt-C"}, searchIDs(got))
	})

	t.Run("BettingStatus only returns matching status", func(t *testing.T) {
		repo := seed(t)
		status := model.BettingStatus_BettingOpen
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{BettingStatus: &status})
		require.NoError(t, err)
		assert.Equal(t, []string{"evt-A", "evt-C"}, searchIDs(got))
	})

	t.Run("Display true matches only the explicitly-true events", func(t *testing.T) {
		repo := seed(t)
		display := true
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{Display: &display})
		require.NoError(t, err)
		assert.Equal(t, []string{"evt-A", "evt-D"}, searchIDs(got))
	})

	t.Run("Display false matches explicitly-false and never-set events", func(t *testing.T) {
		repo := seed(t)
		display := false
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{Display: &display})
		require.NoError(t, err)
		assert.Equal(t, []string{"evt-B", "evt-C"}, searchIDs(got))
	})

	t.Run("date, status and display filters combine with AND", func(t *testing.T) {
		repo := seed(t)
		from := baseTime
		to := baseTime + 3*hour
		status := model.BettingStatus_BettingOpen
		display := true
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{
			StartTimeFrom: &from,
			StartTimeTo:   &to,
			BettingStatus: &status,
			Display:       &display,
		})
		require.NoError(t, err)
		// evt-C matches the window and status but not display; evt-A matches all three.
		assert.Equal(t, []string{"evt-A"}, searchIDs(got))
	})

	t.Run("a filter matching nothing returns an empty slice, not an error", func(t *testing.T) {
		repo := seed(t)
		from := baseTime + 100*hour
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{StartTimeFrom: &from})
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

// Test_mongoRepo_SearchEvents_UnsetFields covers events that never set
// StartTime and/or BettingStatus - a separate fixture from
// Test_mongoRepo_SearchEvents rather than an addition to searchFixture, so
// none of that function's nine exact-ID-list assertions need to change.
func Test_mongoRepo_SearchEvents_UnsetFields(t *testing.T) {
	repo := newSearchTestRepo(t)

	// unset-A has no StartTime, no BettingStatus and no Display at all - what
	// a brand-new event looks like. unset-B has an explicit BettingStatus but
	// still no StartTime, so a BettingStatus-only search can tell the two apart.
	seedSearchEvent(t, repo, &model.Event{ID: "unset-A"})
	seedSearchEvent(t, repo, &model.Event{
		ID:            "unset-B",
		BettingStatus: &model.OptionalBettingStatus{Value: model.BettingStatus_BettingOpen},
	})

	t.Run("BettingStatus: BettingUnknown finds an event that never set a status", func(t *testing.T) {
		status := model.BettingStatus_BettingUnknown
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{BettingStatus: &status})
		require.NoError(t, err)
		assert.Equal(t, []string{"unset-A"}, searchIDs(got))
	})

	// The probe-C trap: widening a non-zero BettingStatus to match null would
	// wrongly return unset-A here too, since it has no status at all.
	t.Run("BettingStatus: BettingOpen does not also match an event with no status", func(t *testing.T) {
		status := model.BettingStatus_BettingOpen
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{BettingStatus: &status})
		require.NoError(t, err)
		assert.Equal(t, []string{"unset-B"}, searchIDs(got))
	})

	t.Run("a date window matches only events that have a StartTime", func(t *testing.T) {
		from := int64(0)
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{StartTimeFrom: &from})
		require.NoError(t, err)
		assert.Empty(t, got, "neither fixture event has a StartTime, so no window should match either")
	})

	// Both events sort to the same (null) starttime.value key, so this is the
	// only assertion in the suite that would fail if the _id tiebreak were
	// dropped from the sort in mongo.go.
	t.Run("no filters, both events sort by _id when StartTime is equally unset", func(t *testing.T) {
		got, err := repo.SearchEvents(context.Background(), repository.EventFilter{})
		require.NoError(t, err)
		assert.Equal(t, []string{"unset-A", "unset-B"}, searchIDs(got))
	})
}
