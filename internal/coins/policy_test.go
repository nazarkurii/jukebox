package coins

import (
	"errors"
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
		name        string
		policy      policy
		sum         int
		expected    []int
		expectedOk  bool
		expectedErr error
	}{
		{
			name:        "returns change when possible (greedy)",
			policy:      policy{denominations: []int{10, 5, 1}},
			sum:         16,
			expected:    []int{10, 5, 1},
			expectedOk:  true,
			expectedErr: nil,
		},
		{
			name:        "returns change when possible (not greedy)",
			policy:      policy{denominations: []int{10, 6}},
			sum:         12,
			expected:    []int{6, 6},
			expectedOk:  true,
			expectedErr: nil,
		},
		{
			name:        "returns false when impossible to calculate change",
			policy:      policy{denominations: []int{10, 6}},
			sum:         7,
			expected:    nil,
			expectedOk:  false,
			expectedErr: nil,
		},
		{
			name:        "returns true when passed sum is 0",
			policy:      policy{denominations: []int{10, 6}},
			sum:         0,
			expected:    nil,
			expectedOk:  true,
			expectedErr: nil,
		},
		{
			name:        "returns sentinel error when passed sum is negative",
			policy:      policy{denominations: []int{10, 5, 1}},
			sum:         -5,
			expected:    nil,
			expectedOk:  false,
			expectedErr: InvalidChangeSum,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, ok, err := tc.policy.CalculateChange(tc.sum)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
				if ok != false {
					t.Errorf("error case: expected ok to be false, got %v", ok)
				}
				if result != nil {
					t.Errorf("error case: expected result to be nil, got %v", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if ok != tc.expectedOk {
				t.Fatalf("expected ok=%v, got %v", tc.expectedOk, ok)
			}
			if !slices.Equal(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}

}

func Test_policy_CalculateChangeDP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		policy      policy
		sum         int
		expected    []int
		expectedOk  bool
		expectedErr error
	}{
		{
			name:        "returns optimal 2 coins for {1, 3, 4} sum 6 (Greedy would fail this)",
			policy:      policy{denominations: []int{4, 3, 1}},
			sum:         6,
			expected:    []int{3, 3},
			expectedOk:  true,
			expectedErr: nil,
		},
		{
			name:        "returns change for large sum with standard denominations",
			policy:      policy{denominations: []int{25, 10, 5, 1}},
			sum:         41,
			expected:    []int{1, 5, 10, 25},
			expectedOk:  true,
			expectedErr: nil,
		},
		{
			name:        "returns true for sum 0",
			policy:      policy{denominations: []int{10, 5, 1}},
			sum:         0,
			expected:    nil,
			expectedOk:  true,
			expectedErr: nil,
		},
		{
			name:        "returns false for impossible sum (no 1s available)",
			policy:      policy{denominations: []int{10, 5}},
			sum:         7,
			expected:    nil,
			expectedOk:  false,
			expectedErr: nil,
		},
		{
			name:        "returns InvalidChangeSum sentinel for negative values",
			policy:      policy{denominations: []int{1, 5, 10}},
			sum:         -10,
			expected:    nil,
			expectedOk:  false,
			expectedErr: InvalidChangeSum,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, ok, err := tc.policy.CalculateChangeDP(tc.sum)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}

				if ok != false {
					t.Errorf("error case: expected ok to be false, got %v", ok)
				}
				if result != nil {
					t.Errorf("error case: expected result to be nil, got %v", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if ok != tc.expectedOk {
				t.Fatalf("expected ok=%v, got %v", tc.expectedOk, ok)
			}
			if ok && !slices.Equal(result, tc.expected) {
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
