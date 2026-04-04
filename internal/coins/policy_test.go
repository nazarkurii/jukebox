package coins

import (
	"slices"
	"testing"
)

func Test_policy_IsValid(t *testing.T) {
	t.Parallel()

	policy := policy{denominations: []int{10, 5, 1}}

	t.Run("returns true when denominations contain amount", func(t *testing.T) {
		t.Parallel()

		if !policy.IsValid(5) {
			t.Error("expected coin to be valid")
		}
	})

	t.Run("returns false when denominations do not contain amount", func(t *testing.T) {
		t.Parallel()

		if policy.IsValid(2) {
			t.Error("expected coin to be invalid")
		}
	})
}

func Test_policy_CalculateChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		policy   policy
		sum      int
		expected []int
		ok       bool
	}{
		{
			name:     "returns change when possible (greedy)",
			policy:   policy{denominations: []int{10, 5, 1}},
			sum:      16,
			expected: []int{10, 5, 1},
			ok:       true,
		},
		{
			name:     "returns change when possible (not greedy)",
			policy:   policy{denominations: []int{10, 6}},
			sum:      12,
			expected: []int{6, 6},
			ok:       true,
		},
		{
			name:     "returns false when impossible to calculate change",
			policy:   policy{denominations: []int{10, 6}},
			sum:      7,
			expected: nil,
			ok:       false,
		},
		{
			name:     "returns true when passed sum is 0",
			policy:   policy{denominations: []int{10, 6}},
			sum:      0,
			expected: nil,
			ok:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, ok := tc.policy.CalculateChange(tc.sum)

			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}

			if !slices.Equal(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
func TestNewPolicy(t *testing.T) {
	t.Parallel()

	policy := NewPolicy(1, 10, 5)

	expected := []int{10, 5, 1}
	if !slices.Equal(policy.denominations, expected) {
		t.Errorf("expected %v, got %v", expected, policy.denominations)
	}
}
