package merger_test

import (
	"context"
	"fmt"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/internal/modeltest"
	"git.neds.sh/technology/pricekinetics/tools/codetest/merger"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// assertAllFieldsSet fails if any field on msg is unset, so an incomplete
// fixture cannot mask an incomplete Merge function.
func assertAllFieldsSet(t *testing.T, msg protoreflect.Message, path string) {
	t.Helper()

	desc := msg.Descriptor()
	fields := desc.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		fieldPath := fmt.Sprintf("%s.%s", path, fd.Name())

		if !msg.Has(fd) {
			t.Errorf("fixture %s is unset - populate it here, then check Merge%s copies it", fieldPath, desc.Name())
			continue
		}

		switch {
		case fd.IsList():
			list := msg.Get(fd).List()
			if fd.Kind() != protoreflect.MessageKind {
				continue
			}
			for j := 0; j < list.Len(); j++ {
				assertAllFieldsSet(t, list.Get(j).Message(), fmt.Sprintf("%s[%d]", fieldPath, j))
			}
		case fd.Kind() == protoreflect.MessageKind:
			assertAllFieldsSet(t, msg.Get(fd).Message(), fieldPath)
		}
	}
}

func TestMergeEvent_MergesEveryField(t *testing.T) {
	right := modeltest.PopulatedEvent()
	assertAllFieldsSet(t, right.ProtoReflect(), "Event")

	result := merger.MergeEvent(context.Background(), &model.Event{}, right)

	assert.True(t, proto.Equal(right, result), "MergeEvent dropped a field:\n want %v\n got  %v", right, result)
}

func TestMergeMarket_MergesEveryField(t *testing.T) {
	right := modeltest.PopulatedMarket("mkt-1")
	assertAllFieldsSet(t, right.ProtoReflect(), "Market")

	result := merger.MergeMarket(context.Background(), &model.Market{}, right)

	assert.True(t, proto.Equal(right, result), "MergeMarket dropped a field:\n want %v\n got  %v", right, result)
}

func TestMergeSelection_MergesEveryField(t *testing.T) {
	right := modeltest.PopulatedSelection("sel-1")
	assertAllFieldsSet(t, right.ProtoReflect(), "Selection")

	result := merger.MergeSelection(context.Background(), &model.Selection{}, right)

	assert.True(t, proto.Equal(right, result), "MergeSelection dropped a field:\n want %v\n got  %v", right, result)
}

func TestMergeSportEvent_MergesEveryField(t *testing.T) {
	right := modeltest.PopulatedSportEvent()
	assertAllFieldsSet(t, right.ProtoReflect(), "SportEvent")

	result := merger.MergeSportEvent(context.Background(), &model.SportEvent{}, right)

	assert.True(t, proto.Equal(right, result), "MergeSportEvent dropped a field:\n want %v\n got  %v", right, result)
}

func TestMergeRacingEvent_MergesEveryField(t *testing.T) {
	right := modeltest.PopulatedRacingEvent()
	assertAllFieldsSet(t, right.ProtoReflect(), "RacingEvent")

	result := merger.MergeRacingEvent(context.Background(), &model.RacingEvent{}, right)

	assert.True(t, proto.Equal(right, result), "MergeRacingEvent dropped a field:\n want %v\n got  %v", right, result)
}

func TestMergeRunner_MergesEveryField(t *testing.T) {
	right := modeltest.PopulatedRunner("1")
	assertAllFieldsSet(t, right.ProtoReflect(), "Runner")

	result := merger.MergeRunner(context.Background(), &model.Runner{}, right)

	assert.True(t, proto.Equal(right, result), "MergeRunner dropped a field:\n want %v\n got  %v", right, result)
}

// Runners merge by ID, so a scratching that names one runner must leave the
// rest of the field untouched.
func TestMergeRunnerSlice_MergesByID(t *testing.T) {
	existing := []*model.Runner{
		{ID: "1", Name: &model.OptionalString{Value: "Winx"}},
		{ID: "2", Name: &model.OptionalString{Value: "Black Caviar"}},
		{ID: "3", Name: &model.OptionalString{Value: "Phar Lap"}},
	}
	scratching := []*model.Runner{
		{ID: "2", Scratched: &model.OptionalBool{Value: true}},
	}

	result := merger.MergeRunnerSlice(context.Background(), existing, scratching)

	assert.Len(t, result, 3)
	byID := map[string]*model.Runner{}
	for _, r := range result {
		byID[r.GetID()] = r
	}
	assert.True(t, byID["2"].GetScratched().GetValue(), "runner 2 should be scratched")
	assert.Equal(t, "Black Caviar", byID["2"].GetName().GetValue(), "scratching must not drop the name")
	assert.False(t, byID["1"].GetScratched().GetValue(), "runner 1 must be untouched")
	assert.False(t, byID["3"].GetScratched().GetValue(), "runner 3 must be untouched")
}

// Merge semantics for the Display field: set, hide, re-show, and leave alone.
func TestMergeEvent_Display(t *testing.T) {
	tests := []struct {
		name  string
		left  *model.OptionalBool
		right *model.OptionalBool
		want  *model.OptionalBool
	}{
		{
			name: "absent on both sides stays absent",
		},
		{
			name:  "first producer to set it wins on a new event",
			right: &model.OptionalBool{Value: false},
			want:  &model.OptionalBool{Value: false},
		},
		{
			name:  "an incoming false hides a visible event",
			left:  &model.OptionalBool{Value: true},
			right: &model.OptionalBool{Value: false},
			want:  &model.OptionalBool{Value: false},
		},
		{
			name:  "an incoming true re-shows a hidden event",
			left:  &model.OptionalBool{Value: false},
			right: &model.OptionalBool{Value: true},
			want:  &model.OptionalBool{Value: true},
		},
		{
			name: "a partial update that omits Display keeps the existing value",
			left: &model.OptionalBool{Value: false},
			want: &model.OptionalBool{Value: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := &model.Event{ID: "evt-1", Display: tt.left}
			update := &model.Event{ID: "evt-1", Display: tt.right}

			result := merger.MergeEvent(context.Background(), existing, update)

			assert.True(t, proto.Equal(tt.want, result.GetDisplay()),
				"want %v, got %v", tt.want, result.GetDisplay())
		})
	}
}
