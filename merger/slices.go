package merger

import (
	"context"
	"math"
	"sort"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
)

// MergeMarketSlice merges two slices of markets
func MergeMarketSlice(ctx context.Context, left, right []*model.Market) []*model.Market {
	// Trivial cases
	if len(left) == 0 && len(right) == 0 {
		return nil
	} else if len(left) == 0 {
		return right
	} else if len(right) == 0 {
		return left
	}

	// Sort to canonical orders
	leftMax := len(left)
	rightMax := len(right)
	sort.Slice(left, func(i, j int) bool {
		return left[i].GetID() < left[j].GetID()
	})
	sort.Slice(right, func(i, j int) bool {
		return right[i].GetID() < right[j].GetID()
	})

	// Work forward through the slices
	leftPosition := 0
	rightPosition := 0
	sortTarget := int(math.Max(float64(leftMax), float64(rightMax)))
	output := make([]*model.Market, 0, sortTarget)
	for {
		if leftPosition >= leftMax && rightPosition >= rightMax {
			// If we're at the end of both lists, we're done
			break
		} else if leftPosition >= leftMax {
			// If we've finished the left list, keep eating the right
			output = append(output, right[rightPosition])
			rightPosition++
			continue
		} else if rightPosition >= rightMax {
			// If we've finished the r
			output = append(output, left[leftPosition])
			leftPosition++
			continue
		}

		// If we've got matching ID's, merge
		leftID := left[leftPosition].GetID()
		rightID := right[rightPosition].GetID()
		if leftID == rightID {
			output = append(output, MergeMarket(ctx, left[leftPosition], right[rightPosition]))
			leftPosition++
			rightPosition++
		} else if leftID < rightID {
			output = append(output, left[leftPosition])
			leftPosition++
		} else {
			output = append(output, right[rightPosition])
			rightPosition++
		}
	}

	// Sort to canonical order
	sort.Slice(output, func(i, j int) bool {
		return output[i].GetID() < output[j].GetID()
	})
	return output
}

// MergeSelectionSlice merges two slices of selections
func MergeSelectionSlice(ctx context.Context, left, right []*model.Selection) []*model.Selection {
	// Trivial cases
	if len(left) == 0 && len(right) == 0 {
		return nil
	} else if len(left) == 0 {
		return right
	} else if len(right) == 0 {
		return left
	}

	// Sort to canonical orders
	leftMax := len(left)
	rightMax := len(right)
	sort.Slice(left, func(i, j int) bool {
		return left[i].GetID() < left[j].GetID()
	})
	sort.Slice(right, func(i, j int) bool {
		return right[i].GetID() < right[j].GetID()
	})

	// Work forward through the slices
	leftPosition := 0
	rightPosition := 0
	sortTarget := int(math.Max(float64(leftMax), float64(rightMax)))
	output := make([]*model.Selection, 0, sortTarget)
	for {
		if leftPosition >= leftMax && rightPosition >= rightMax {
			// If we're at the end of both lists, we're done
			break
		} else if leftPosition >= leftMax {
			// If we've finished the left list, keep eating the right
			output = append(output, right[rightPosition])
			rightPosition++
			continue
		} else if rightPosition >= rightMax {
			// If we've finished the r
			output = append(output, left[leftPosition])
			leftPosition++
			continue
		}

		// If we've got matching ID's, merge
		leftID := left[leftPosition].GetID()
		rightID := right[rightPosition].GetID()
		if leftID == rightID {
			output = append(output, MergeSelection(ctx, left[leftPosition], right[rightPosition]))
			leftPosition++
			rightPosition++
		} else if leftID < rightID {
			output = append(output, left[leftPosition])
			leftPosition++
		} else {
			output = append(output, right[rightPosition])
			rightPosition++
		}
	}

	// Sort to canonical order
	sort.Slice(output, func(i, j int) bool {
		return output[i].GetID() < output[j].GetID()
	})
	return output
}
