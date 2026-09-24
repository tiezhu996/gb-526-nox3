package decompression

import "math"

const waterVaporPressureBar = 0.0627

func AmbientPressureBar(depthM, worksitePressureBar float64) float64 {
	return round4(worksitePressureBar + depthM/10)
}

func InspiredPartialPressureBar(ambientPressureBar, fraction float64) float64 {
	dryPressure := math.Max(0, ambientPressureBar-waterVaporPressureBar)
	return round4(dryPressure * fraction)
}

func round4(value float64) float64 { return math.Round(value*10000) / 10000 }

func round2(value float64) float64 { return math.Round(value*100) / 100 }
