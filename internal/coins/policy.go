package coins

import (
	"errors"
	"slices"
)

var (
	InvalidChangeSum = errors.New("impossible to calculate the change for a negative sum")
)

type policy struct {
	denominations []int
}

func (c policy) IsValid(denomination int) bool {
	return slices.Contains(c.denominations, denomination)
}

func (c policy) Denominations() []int {
	return c.denominations
}

func (c policy) CalculateChange(changeSum int) ([]int, bool, error) {
	if changeSum < 0 {
		return nil, false, InvalidChangeSum
	} else if changeSum == 0 {
		return nil, true, nil
	}

	var coins []int
	var total int

	for i := range len(c.denominations) {
		for i := i; i < len(c.denominations); {
			denomination := c.denominations[i]
			total += denomination
			if total > changeSum {
				total -= denomination
				i++
				continue
			}

			coins = append(coins, denomination)
			if total == changeSum {
				return coins, true, nil
			}
		}
		coins = nil
		total = 0
	}

	return nil, false, nil
}

func (c policy) CalculateChangeDP(changeSum int) ([]int, bool, error) {
	if changeSum < 0 {
		return nil, false, InvalidChangeSum
	} else if changeSum == 0 {
		return nil, true, nil
	}

	dpLenth := changeSum + 1
	dp := make([]int, dpLenth)
	parents := make([]int, dpLenth)

	for _, denomination := range c.denominations {
		for sum := denomination; sum < dpLenth; sum++ {
			rest := sum - denomination
			totalCoins := 1 + dp[rest]
			if (dp[sum] >= totalCoins || dp[sum] == 0) && (dp[rest] != 0 || rest == 0) {
				dp[sum] = totalCoins
				parents[sum] = denomination
			}
		}

	}

	change := make([]int, dp[changeSum])

	for i := range change {
		change[i] = parents[changeSum]
		changeSum -= parents[changeSum]
	}

	return change, len(change) > 0, nil
}

func NewPolicy(denominations ...int) policy {
	slices.Sort(denominations)
	slices.Reverse(denominations)
	return policy{denominations: denominations}
}
