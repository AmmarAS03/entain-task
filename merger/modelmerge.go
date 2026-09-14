package merger

// Package merging merges instance of the codetest models
//lint:file-ignore SA4006 Generated code, potentially empty
//lint:file-ignore SA9003 Generated code, potentially empty
//lint:file-ignore SA1019 Deprecated code

import (
	"context"
	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// MergeBettingStatus generates a merged value between two members of the enumeration BettingStatus
func MergeBettingStatus(ctx context.Context, left, right model.BettingStatus) model.BettingStatus {
	// For enumerated types, we simply return the right operand
	return right
}

// MergeOptionalBettingStatus generates a new instance of the OptionalBettingStatus type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeOptionalBettingStatus(ctx context.Context, left, right *model.OptionalBettingStatus) *model.OptionalBettingStatus {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.OptionalBettingStatus{}

	result.Value = MergeBettingStatus(ctx, left.Value, right.Value)
	result.Deleted = right.Deleted // Copy primitive value from right, as non-pointers.
	return result
}

// MergeRaceStatus generates a merged value between two members of the enumeration RaceStatus
func MergeRaceStatus(ctx context.Context, left, right model.RaceStatus) model.RaceStatus {
	// For enumerated types, we simply return the right operand
	return right
}

// MergeOptionalRaceStatus generates a new instance of the OptionalRaceStatus type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeOptionalRaceStatus(ctx context.Context, left, right *model.OptionalRaceStatus) *model.OptionalRaceStatus {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.OptionalRaceStatus{}

	result.Value = MergeRaceStatus(ctx, left.Value, right.Value)
	result.Deleted = right.Deleted // Copy primitive value from right, as non-pointers.
	return result
}

// MergeEvent generates a new instance of the Event type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeEvent(ctx context.Context, left, right *model.Event) *model.Event {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.Event{}

	result.ID = right.ID // Copy primitive value from right, as non-pointers.
	result.Name = MergeOptionalString(ctx, left.Name, right.Name)
	result.StartTime = MergeOptionalInt64(ctx, left.StartTime, right.StartTime)
	result.BettingStatus = MergeOptionalBettingStatus(ctx, left.BettingStatus, right.BettingStatus)
	result.SportData = MergeSportEvent(ctx, left.SportData, right.SportData)
	result.Display = MergeOptionalBool(ctx, left.Display, right.Display)
	result.RacingData = MergeRacingEvent(ctx, left.RacingData, right.RacingData)

	// Generate the difference for Markets with a slice of Market
	mergedMarkets := MergeMarketSlice(ctx, left.Markets, right.Markets)
	if len(mergedMarkets) > 0 {
		result.Markets = mergedMarkets
	}
	result.EventTypeID = MergeOptionalString(ctx, left.EventTypeID, right.EventTypeID)
	return result
}

// MergeSportEvent generates a new instance of the SportEvent type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeSportEvent(ctx context.Context, left, right *model.SportEvent) *model.SportEvent {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.SportEvent{}

	result.Name = MergeOptionalString(ctx, left.Name, right.Name)
	result.Region = MergeOptionalString(ctx, left.Region, right.Region)
	result.League = MergeOptionalString(ctx, left.League, right.League)
	result.Round = MergeOptionalString(ctx, left.Round, right.Round)
	return result
}

// MergeRacingEvent generates a new instance of the RacingEvent type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeRacingEvent(ctx context.Context, left, right *model.RacingEvent) *model.RacingEvent {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.RacingEvent{}

	result.Name = MergeOptionalString(ctx, left.Name, right.Name)
	result.Region = MergeOptionalString(ctx, left.Region, right.Region)
	result.TrackName = MergeOptionalString(ctx, left.TrackName, right.TrackName)
	result.RaceNumber = MergeOptionalInt64(ctx, left.RaceNumber, right.RaceNumber)
	result.DistanceMetres = MergeOptionalInt64(ctx, left.DistanceMetres, right.DistanceMetres)
	result.TrackCondition = MergeOptionalString(ctx, left.TrackCondition, right.TrackCondition)
	result.Weather = MergeOptionalString(ctx, left.Weather, right.Weather)
	result.RaceClass = MergeOptionalString(ctx, left.RaceClass, right.RaceClass)
	result.RaceStatus = MergeOptionalRaceStatus(ctx, left.RaceStatus, right.RaceStatus)
	result.FieldSize = MergeOptionalInt64(ctx, left.FieldSize, right.FieldSize)

	// Generate the difference for Runners with a slice of Runner
	mergedRunners := MergeRunnerSlice(ctx, left.Runners, right.Runners)
	if len(mergedRunners) > 0 {
		result.Runners = mergedRunners
	}
	return result
}

// MergeRunner generates a new instance of the Runner type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeRunner(ctx context.Context, left, right *model.Runner) *model.Runner {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.Runner{}

	result.ID = right.ID // Copy primitive value from right, as non-pointers.
	result.Number = MergeOptionalInt64(ctx, left.Number, right.Number)
	result.Name = MergeOptionalString(ctx, left.Name, right.Name)
	result.Barrier = MergeOptionalInt64(ctx, left.Barrier, right.Barrier)
	result.Weight = MergeOptionalDouble(ctx, left.Weight, right.Weight)
	result.Jockey = MergeOptionalString(ctx, left.Jockey, right.Jockey)
	result.Trainer = MergeOptionalString(ctx, left.Trainer, right.Trainer)
	result.Scratched = MergeOptionalBool(ctx, left.Scratched, right.Scratched)
	result.Silks = MergeOptionalString(ctx, left.Silks, right.Silks)
	result.FinishPosition = MergeOptionalInt64(ctx, left.FinishPosition, right.FinishPosition)
	return result
}

// MergeMarket generates a new instance of the Market type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeMarket(ctx context.Context, left, right *model.Market) *model.Market {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.Market{}

	result.ID = right.ID // Copy primitive value from right, as non-pointers.
	result.Name = MergeOptionalString(ctx, left.Name, right.Name)
	result.StartTime = MergeOptionalInt64(ctx, left.StartTime, right.StartTime)
	result.BettingStatus = MergeOptionalBettingStatus(ctx, left.BettingStatus, right.BettingStatus)

	// Generate the difference for Selections with a slice of Selection
	mergedSelections := MergeSelectionSlice(ctx, left.Selections, right.Selections)
	if len(mergedSelections) > 0 {
		result.Selections = mergedSelections
	}
	return result
}

// MergeSelection generates a new instance of the Selection type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeSelection(ctx context.Context, left, right *model.Selection) *model.Selection {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.Selection{}

	result.ID = right.ID // Copy primitive value from right, as non-pointers.
	result.Name = MergeOptionalString(ctx, left.Name, right.Name)
	result.BettingStatus = MergeOptionalBettingStatus(ctx, left.BettingStatus, right.BettingStatus)
	result.Price = MergeOptionalDouble(ctx, left.Price, right.Price)
	return result
}

// MergeOptionalString generates a new instance of the OptionalString type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeOptionalString(ctx context.Context, left, right *model.OptionalString) *model.OptionalString {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.OptionalString{}

	result.Value = right.Value     // Copy primitive value from right, as non-pointers.
	result.Deleted = right.Deleted // Copy primitive value from right, as non-pointers.
	return result
}

// MergeOptionalBool generates a new instance of the OptionalBool type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeOptionalBool(ctx context.Context, left, right *model.OptionalBool) *model.OptionalBool {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.OptionalBool{}

	result.Value = right.Value     // Copy primitive value from right, as non-pointers.
	result.Deleted = right.Deleted // Copy primitive value from right, as non-pointers.
	return result
}

// MergeOptionalDouble generates a new instance of the OptionalDouble type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeOptionalDouble(ctx context.Context, left, right *model.OptionalDouble) *model.OptionalDouble {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.OptionalDouble{}

	result.Value = right.Value     // Copy primitive value from right, as non-pointers.
	result.Deleted = right.Deleted // Copy primitive value from right, as non-pointers.
	return result
}

// MergeOptionalInt64 generates a new instance of the OptionalInt64 type, where two input values are merged. Values on the left
// are overwritten with values from the right where they exist, recursively.
func MergeOptionalInt64(ctx context.Context, left, right *model.OptionalInt64) *model.OptionalInt64 {
	// Handle trivial cases
	if right == nil {
		return left
	}
	if left == nil {
		return right
	}

	// Create the new target
	result := &model.OptionalInt64{}

	result.Value = right.Value     // Copy primitive value from right, as non-pointers.
	result.Deleted = right.Deleted // Copy primitive value from right, as non-pointers.
	return result
}
