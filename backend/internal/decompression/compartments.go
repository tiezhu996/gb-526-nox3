package decompression

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"commercial-diving-decompression-control/backend/internal/model"
)

type CompartmentSpec struct {
	Name          string  `json:"name"`
	N2HalfTimeMin float64 `json:"n2_half_time_min"`
	HeHalfTimeMin float64 `json:"he_half_time_min"`
}

type LoadPoint struct {
	SequenceNo     int     `json:"sequence_no"`
	DepthM         float64 `json:"depth_m"`
	ElapsedMin     float64 `json:"elapsed_min"`
	AmbientBar     float64 `json:"ambient_bar"`
	InspiredN2Bar  float64 `json:"inspired_n2_bar"`
	InspiredHeBar  float64 `json:"inspired_he_bar"`
	N2LoadBar      float64 `json:"n2_load_bar"`
	HeLoadBar      float64 `json:"he_load_bar"`
	TotalInertBar  float64 `json:"total_inert_bar"`
	RelativeChange float64 `json:"relative_change"`
}

type CompartmentCurve struct {
	Name          string      `json:"name"`
	N2HalfTimeMin float64     `json:"n2_half_time_min"`
	HeHalfTimeMin float64     `json:"he_half_time_min"`
	Points        []LoadPoint `json:"points"`
}

type ModelAssumptions struct {
	Purpose                string            `json:"purpose"`
	PressureConversion     string            `json:"pressure_conversion"`
	WaterVaporBar          float64           `json:"water_vapor_bar"`
	Kinetics               string            `json:"kinetics"`
	ConfiguredCompartments []CompartmentSpec `json:"configured_compartments"`
	ProhibitedUse          []string          `json:"prohibited_use"`
}

type InputSnapshot struct {
	Plan             model.DivePlan          `json:"plan"`
	Diver            model.DiverProfile      `json:"diver"`
	Segments         []model.ExposureSegment `json:"segments"`
	Compartments     []CompartmentSpec       `json:"compartments"`
	AlgorithmVersion string                  `json:"algorithm_version"`
	SafetyBoundary   string                  `json:"safety_boundary"`
}

type Result struct {
	Snapshot         InputSnapshot      `json:"snapshot"`
	Curves           []CompartmentCurve `json:"curves"`
	RiskFlags        []RiskFlag         `json:"risk_flags"`
	ComparativeScore float64            `json:"comparative_score"`
	Assumptions      ModelAssumptions   `json:"assumptions"`
}

func DefaultCompartments() []CompartmentSpec {
	return []CompartmentSpec{
		{Name: "C05", N2HalfTimeMin: 5, HeHalfTimeMin: 2},
		{Name: "C10", N2HalfTimeMin: 10, HeHalfTimeMin: 4},
		{Name: "C20", N2HalfTimeMin: 20, HeHalfTimeMin: 8},
		{Name: "C40", N2HalfTimeMin: 40, HeHalfTimeMin: 16},
		{Name: "C80", N2HalfTimeMin: 80, HeHalfTimeMin: 32},
		{Name: "C120", N2HalfTimeMin: 120, HeHalfTimeMin: 48},
	}
}

func BuildSnapshot(plan model.DivePlan, diver model.DiverProfile, segments []model.ExposureSegment, specs []CompartmentSpec, version string) InputSnapshot {
	ordered := append([]model.ExposureSegment(nil), segments...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].SequenceNo < ordered[j].SequenceNo })
	return InputSnapshot{
		Plan: plan, Diver: diver, Segments: ordered, Compartments: append([]CompartmentSpec(nil), specs...),
		AlgorithmVersion: version,
		SafetyBoundary:   "Training comparison and decision support only; not an executable decompression schedule or safety clearance.",
	}
}

func Run(plan model.DivePlan, diver model.DiverProfile, segments []model.ExposureSegment, modelVersion string, maxSegments int) (Result, error) {
	ordered, err := ValidateInput(plan, segments, maxSegments)
	if err != nil {
		return Result{}, err
	}
	specs := DefaultCompartments()
	curves, err := calculateCurves(plan, diver, ordered, specs)
	if err != nil {
		return Result{}, err
	}
	flags := EvaluateRisk(plan, ordered, curves)
	return Result{
		Snapshot:         BuildSnapshot(plan, diver, ordered, specs, modelVersion),
		Curves:           curves,
		RiskFlags:        flags,
		ComparativeScore: ComparativeScore(flags, curves),
		Assumptions: ModelAssumptions{
			Purpose:                "Deterministic offline training comparison",
			PressureConversion:     "ambient_bar = worksite_pressure_bar + depth_m / 10",
			WaterVaporBar:          waterVaporPressureBar,
			Kinetics:               "Constant-depth exponential approach per segment; helium half-time is independently configured",
			ConfiguredCompartments: specs,
			ProhibitedUse:          []string{"live dive control", "medical diagnosis", "executable stop depth or duration", "equipment connection", "safety certification"},
		},
	}, nil
}

func calculateCurves(plan model.DivePlan, diver model.DiverProfile, segments []model.ExposureSegment, specs []CompartmentSpec) ([]CompartmentCurve, error) {
	baselineAmbient := AmbientPressureBar(0, plan.WorksitePressureBar)
	defaultMix := GasMix{O2: diver.DefaultO2Fraction, He: diver.DefaultHeFraction, N2: diver.DefaultN2Fraction()}
	if err := defaultMix.Validate(); err != nil {
		return nil, fmt.Errorf("default profile gas assumption: %w", err)
	}
	baselineN2 := InspiredPartialPressureBar(baselineAmbient, defaultMix.N2)
	baselineHe := InspiredPartialPressureBar(baselineAmbient, defaultMix.He)
	curves := make([]CompartmentCurve, 0, len(specs))
	for _, spec := range specs {
		n2Load, heLoad, elapsed := baselineN2, baselineHe, 0.0
		points := make([]LoadPoint, 0, len(segments)+1)
		points = append(points, LoadPoint{SequenceNo: 0, DepthM: 0, ElapsedMin: 0, AmbientBar: baselineAmbient, InspiredN2Bar: baselineN2, InspiredHeBar: baselineHe, N2LoadBar: n2Load, HeLoadBar: heLoad, TotalInertBar: round4(n2Load + heLoad)})
		for _, segment := range segments {
			mix, err := DecodeGasMix(segment.GasMixJSON)
			if err != nil {
				return nil, fmt.Errorf("segment %d gas: %w", segment.SequenceNo, err)
			}
			ambient := AmbientPressureBar(segment.DepthM, plan.WorksitePressureBar)
			inspiredN2 := InspiredPartialPressureBar(ambient, mix.N2)
			inspiredHe := InspiredPartialPressureBar(ambient, mix.He)
			n2Load = exponentialLoad(n2Load, inspiredN2, segment.DurationMin, spec.N2HalfTimeMin)
			heLoad = exponentialLoad(heLoad, inspiredHe, segment.DurationMin, spec.HeHalfTimeMin)
			elapsed += segment.DurationMin
			total := n2Load + heLoad
			baseline := math.Max(0.0001, baselineN2+baselineHe)
			points = append(points, LoadPoint{
				SequenceNo: segment.SequenceNo, DepthM: round2(segment.DepthM), ElapsedMin: round2(elapsed), AmbientBar: ambient,
				InspiredN2Bar: inspiredN2, InspiredHeBar: inspiredHe, N2LoadBar: round4(n2Load), HeLoadBar: round4(heLoad),
				TotalInertBar: round4(total), RelativeChange: round4((total - baseline) / baseline),
			})
		}
		curves = append(curves, CompartmentCurve{Name: spec.Name, N2HalfTimeMin: spec.N2HalfTimeMin, HeHalfTimeMin: spec.HeHalfTimeMin, Points: points})
	}
	return curves, nil
}

func exponentialLoad(initial, inspired, duration, halfTime float64) float64 {
	if duration <= 0 || halfTime <= 0 {
		return initial
	}
	fraction := 1 - math.Exp(-math.Ln2*duration/halfTime)
	return initial + (inspired-initial)*fraction
}

func MarshalResult(result Result) (snapshot, curves, flags, assumptions string, err error) {
	values := []any{result.Snapshot, result.Curves, result.RiskFlags, result.Assumptions}
	encoded := make([]string, len(values))
	for index, value := range values {
		data, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			return "", "", "", "", fmt.Errorf("marshal model result part %d: %w", index, marshalErr)
		}
		encoded[index] = string(data)
	}
	return encoded[0], encoded[1], encoded[2], encoded[3], nil
}
