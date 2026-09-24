package constants

type RiskBand string

const (
	RiskInformational RiskBand = "informational"
	RiskCaution       RiskBand = "caution"
	RiskElevated      RiskBand = "elevated"
	RiskInvalid       RiskBand = "invalid"
)

func ValidRiskBand(band RiskBand) bool {
	switch band {
	case RiskInformational, RiskCaution, RiskElevated, RiskInvalid:
		return true
	default:
		return false
	}
}
