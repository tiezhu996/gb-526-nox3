package decompression

import (
	"fmt"
	"math"
	"sort"

	"commercial-diving-decompression-control/backend/internal/constants"
	"commercial-diving-decompression-control/backend/internal/model"
)

type RiskFlag struct {
	Code     string             `json:"code"`
	Band     constants.RiskBand `json:"band"`
	Message  string             `json:"message"`
	Evidence string             `json:"evidence"`
}

func ValidateInput(plan model.DivePlan, segments []model.ExposureSegment, maxSegments int) ([]model.ExposureSegment, error) {
	if plan.WorksitePressureBar < 0.7 || plan.WorksitePressureBar > 1.5 {
		return nil, fmt.Errorf("worksite pressure %.2f bar is outside the training model boundary 0.7-1.5", plan.WorksitePressureBar)
	}
	if _, err := DecodeGasMix(plan.BreathingMixJSON); err != nil {
		return nil, fmt.Errorf("plan breathing mix: %w", err)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("at least one exposure segment is required")
	}
	if len(segments) > maxSegments {
		return nil, fmt.Errorf("segment count %d exceeds configured maximum %d", len(segments), maxSegments)
	}
	ordered := append([]model.ExposureSegment(nil), segments...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].SequenceNo < ordered[j].SequenceNo })
	previousDepth := 0.0
	for index, segment := range ordered {
		if segment.SequenceNo != index+1 {
			return nil, fmt.Errorf("segment sequence must be continuous from 1; expected %d got %d", index+1, segment.SequenceNo)
		}
		if segment.DepthM < 0 || segment.DepthM > 120 {
			return nil, fmt.Errorf("segment %d depth must be between 0 and 120 m", segment.SequenceNo)
		}
		if segment.DurationMin <= 0 || segment.DurationMin > 240 {
			return nil, fmt.Errorf("segment %d duration must be greater than 0 and at most 240 min", segment.SequenceNo)
		}
		if segment.AscentRateMMin < 0 || segment.AscentRateMMin > 18 {
			return nil, fmt.Errorf("segment %d ascent rate must be between 0 and 18 m/min", segment.SequenceNo)
		}
		if _, err := DecodeGasMix(segment.GasMixJSON); err != nil {
			return nil, fmt.Errorf("segment %d: %w", segment.SequenceNo, err)
		}
		delta := segment.DepthM - previousDepth
		travelRate := math.Abs(delta) / segment.DurationMin
		if travelRate > 18.0001 {
			return nil, fmt.Errorf("segment %d depth transition %.2f m/min exceeds continuity boundary", segment.SequenceNo, travelRate)
		}
		switch segment.SegmentType {
		case "descent":
			if delta < 0 {
				return nil, fmt.Errorf("segment %d descent cannot end shallower than prior segment", segment.SequenceNo)
			}
		case "bottom", "transit":
		case "ascent":
			if delta > 0 || segment.AscentRateMMin <= 0 {
				return nil, fmt.Errorf("segment %d ascent must end no deeper and declare a positive ascent rate", segment.SequenceNo)
			}
			if travelRate > segment.AscentRateMMin+0.001 {
				return nil, fmt.Errorf("segment %d transition rate %.2f exceeds declared ascent rate %.2f", segment.SequenceNo, travelRate, segment.AscentRateMMin)
			}
		case "surface":
			if segment.DepthM != 0 {
				return nil, fmt.Errorf("segment %d surface segment depth must be 0", segment.SequenceNo)
			}
		default:
			return nil, fmt.Errorf("segment %d has unsupported type %q", segment.SequenceNo, segment.SegmentType)
		}
		previousDepth = segment.DepthM
	}
	return ordered, nil
}

func EvaluateRisk(plan model.DivePlan, segments []model.ExposureSegment, curves []CompartmentCurve) []RiskFlag {
	flags := []RiskFlag{{Code: "TRAINING_ONLY", Band: constants.RiskInformational, Message: "Model output is limited to training comparison and human review.", Evidence: "No executable stop depths or durations are produced."}}
	maxDepth, totalDuration, maxO2Partial, maxTransition := 0.0, 0.0, 0.0, 0.0
	previousDepth := 0.0
	for _, segment := range segments {
		maxDepth = math.Max(maxDepth, segment.DepthM)
		totalDuration += segment.DurationMin
		mix, _ := DecodeGasMix(segment.GasMixJSON)
		maxO2Partial = math.Max(maxO2Partial, AmbientPressureBar(segment.DepthM, plan.WorksitePressureBar)*mix.O2)
		maxTransition = math.Max(maxTransition, math.Abs(segment.DepthM-previousDepth)/segment.DurationMin)
		previousDepth = segment.DepthM
	}
	if maxDepth > 50 {
		flags = append(flags, RiskFlag{Code: "DEPTH_ASSUMPTION_ELEVATED", Band: constants.RiskElevated, Message: "The modeled depth is outside the routine training comparison range.", Evidence: fmt.Sprintf("maximum modeled depth %.1f m", maxDepth)})
	} else if maxDepth > 30 {
		flags = append(flags, RiskFlag{Code: "DEPTH_ASSUMPTION_CAUTION", Band: constants.RiskCaution, Message: "Depth assumption requires explicit supervisor attention.", Evidence: fmt.Sprintf("maximum modeled depth %.1f m", maxDepth)})
	}
	if totalDuration > 90 {
		flags = append(flags, RiskFlag{Code: "DURATION_ASSUMPTION_CAUTION", Band: constants.RiskCaution, Message: "Long aggregate exposure assumption should be compared with an alternative.", Evidence: fmt.Sprintf("aggregate duration %.1f min", totalDuration)})
	}
	if maxO2Partial > 1.4 {
		flags = append(flags, RiskFlag{Code: "O2_PARTIAL_ASSUMPTION_ELEVATED", Band: constants.RiskElevated, Message: "Modeled oxygen partial-pressure assumption exceeds the project review threshold.", Evidence: fmt.Sprintf("maximum modeled oxygen partial pressure %.2f bar", maxO2Partial)})
	}
	if maxTransition > 12 {
		flags = append(flags, RiskFlag{Code: "TRANSITION_RATE_CAUTION", Band: constants.RiskCaution, Message: "A depth transition approaches the model input boundary.", Evidence: fmt.Sprintf("maximum inferred transition %.2f m/min", maxTransition)})
	}
	maxRelative := 0.0
	for _, curve := range curves {
		for _, point := range curve.Points {
			maxRelative = math.Max(maxRelative, point.RelativeChange)
		}
	}
	if maxRelative > 2.5 {
		flags = append(flags, RiskFlag{Code: "RELATIVE_LOAD_CHANGE_ELEVATED", Band: constants.RiskElevated, Message: "One or more modeled compartments show a large relative change from the configured baseline.", Evidence: fmt.Sprintf("maximum relative change %.2f", maxRelative)})
	} else if maxRelative > 1.25 {
		flags = append(flags, RiskFlag{Code: "RELATIVE_LOAD_CHANGE_CAUTION", Band: constants.RiskCaution, Message: "Modeled compartment change warrants comparison and supervisor review.", Evidence: fmt.Sprintf("maximum relative change %.2f", maxRelative)})
	}
	return flags
}

func ComparativeScore(flags []RiskFlag, curves []CompartmentCurve) float64 {
	score := 100.0
	for _, flag := range flags {
		switch flag.Band {
		case constants.RiskElevated:
			score -= 18
		case constants.RiskCaution:
			score -= 7
		case constants.RiskInvalid:
			score -= 35
		}
	}
	if score < 0 {
		score = 0
	}
	return round2(score)
}

func HighestRiskBand(flags []RiskFlag) constants.RiskBand {
	highest := constants.RiskInformational
	rank := map[constants.RiskBand]int{constants.RiskInformational: 0, constants.RiskCaution: 1, constants.RiskElevated: 2, constants.RiskInvalid: 3}
	for _, flag := range flags {
		if rank[flag.Band] > rank[highest] {
			highest = flag.Band
		}
	}
	return highest
}
