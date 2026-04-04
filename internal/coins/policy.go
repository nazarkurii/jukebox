package coins

import "slices"

type policy struct {
	denominations []int
}

func (c policy) IsValid(denomination int) bool {
	return slices.Contains(c.denominations, denomination)
}

func (c policy) Denominations() []int {
	return c.denominations
}

func (c policy) CalculateChange(changeSum int) ([]int, bool) {
	if changeSum == 0 {
		return nil, true
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
				return coins, true
			}
		}
		coins = nil
		total = 0
	}

	return nil, false
}

func NewPolicy(denominations ...int) policy {
	slices.Sort(denominations)
	slices.Reverse(denominations)
	return policy{denominations: denominations}
}
