package decompression

import (
	"encoding/json"
	"reflect"
	"testing"

	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/model"
)

func testPlanAndSegments(t *testing.T) (model.DivePlan, model.DiverProfile, []model.ExposureSegment) {
	t.Helper()
	mix, err := EncodeGasMix(GasMix{O2: 0.21, N2: 0.79})
	if err != nil {
		t.Fatal(err)
	}
	plan := model.DivePlan{ID: 1, PlanCode: "TEST-1", WorksitePressureBar: 1, BreathingMixJSON: mix, PlanStatus: constants.PlanDraft, Version: 1}
	diver := model.DiverProfile{ID: 1, ProfileCode: "D-1", DefaultO2Fraction: 0.21, DefaultHeFraction: 0}
	segments := []model.ExposureSegment{
		{ID: 1, PlanID: 1, SequenceNo: 1, DepthM: 20, DurationMin: 2, GasMixJSON: mix, SegmentType: "descent"},
		{ID: 2, PlanID: 1, SequenceNo: 2, DepthM: 20, DurationMin: 18, GasMixJSON: mix, SegmentType: "bottom"},
		{ID: 3, PlanID: 1, SequenceNo: 3, DepthM: 0, DurationMin: 2, AscentRateMMin: 10, GasMixJSON: mix, SegmentType: "ascent"},
	}
	return plan, diver, segments
}

func TestRunIsDeterministic(t *testing.T) {
	plan, diver, segments := testPlanAndSegments(t)
	first, err := Run(plan, diver, segments, "training-v1", 24)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(plan, diver, segments, "training-v1", 24)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fixed input produced different model results")
	}
	if len(first.Curves) != 6 || len(first.Curves[0].Points) != 4 {
		t.Fatalf("unexpected curve shape: compartments=%d points=%d", len(first.Curves), len(first.Curves[0].Points))
	}
}

func TestCompartmentCalculation(t *testing.T) {
	plan, diver, segments := testPlanAndSegments(t)
	tests := []struct {
		name        string
		bottomDepth float64
		wantPoints  int
	}{
		{"twenty meter profile", 20, 4},
		{"thirty meter profile", 30, 4},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := append([]model.ExposureSegment(nil), segments...)
			input[0].DepthM = test.bottomDepth
			input[1].DepthM = test.bottomDepth
			input[2].AscentRateMMin = 18
			result, err := Run(plan, diver, input, "training-v1", 24)
			if err != nil {
				t.Fatal(err)
			}
			points := result.Curves[0].Points
			if len(points) != test.wantPoints {
				t.Fatalf("points=%d want=%d", len(points), test.wantPoints)
			}
			if points[2].TotalInertBar <= points[0].TotalInertBar {
				t.Fatalf("bottom load %.4f did not exceed baseline %.4f", points[2].TotalInertBar, points[0].TotalInertBar)
			}
		})
	}
}

func TestSnapshotReplayPayload(t *testing.T) {
	plan, diver, segments := testPlanAndSegments(t)
	tests := []struct {
		name    string
		version string
	}{
		{"current model", "training-v1"},
		{"explicit version", "training-v1.1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := BuildSnapshot(plan, diver, segments, DefaultCompartments(), test.version)
			encoded, err := json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			var decoded InputSnapshot
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.AlgorithmVersion != test.version || len(decoded.Segments) != len(segments) || decoded.SafetyBoundary == "" {
				t.Fatalf("snapshot lost replay evidence: %+v", decoded)
			}
		})
	}
}
