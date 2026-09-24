package decompression

import (
	"encoding/json"
	"fmt"
	"math"
)

type GasMix struct {
	O2 float64 `json:"o2"`
	He float64 `json:"he"`
	N2 float64 `json:"n2"`
}

func (g GasMix) Sum() float64 { return g.O2 + g.He + g.N2 }

func (g GasMix) Validate() error {
	values := map[string]float64{"o2": g.O2, "he": g.He, "n2": g.N2}
	for name, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return fmt.Errorf("gas component %s must be between 0 and 1", name)
		}
	}
	if math.Abs(g.Sum()-1) > 0.001 {
		return fmt.Errorf("gas components must sum to 1, got %.4f", g.Sum())
	}
	if g.O2 < 0.16 {
		return fmt.Errorf("o2 fraction %.3f is outside this training model input boundary", g.O2)
	}
	return nil
}

func DecodeGasMix(raw string) (GasMix, error) {
	var mix GasMix
	if err := json.Unmarshal([]byte(raw), &mix); err != nil {
		return GasMix{}, fmt.Errorf("decode gas mix: %w", err)
	}
	if err := mix.Validate(); err != nil {
		return GasMix{}, err
	}
	return mix, nil
}

func EncodeGasMix(mix GasMix) (string, error) {
	if err := mix.Validate(); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(mix)
	if err != nil {
		return "", fmt.Errorf("encode gas mix: %w", err)
	}
	return string(encoded), nil
}
