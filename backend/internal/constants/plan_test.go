package constants

import "testing"

func TestPlanTransitions(t *testing.T) {
	tests := []struct {
		name string
		from PlanStatus
		to   PlanStatus
		want bool
	}{
		{"model draft", PlanDraft, PlanModeled, true},
		{"submit modeled", PlanModeled, PlanPendingReview, true},
		{"approve review", PlanPendingReview, PlanApprovedTraining, true},
		{"archive approval", PlanApprovedTraining, PlanArchived, true},
		{"cannot skip review", PlanModeled, PlanApprovedTraining, false},
		{"archive terminal", PlanArchived, PlanDraft, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CanTransitionPlan(test.from, test.to); got != test.want {
				t.Fatalf("transition %s -> %s = %t, want %t", test.from, test.to, got, test.want)
			}
		})
	}
}
