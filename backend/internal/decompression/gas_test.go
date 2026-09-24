package decompression

import "testing"

func TestGasMixValidate(t *testing.T) {
	tests := []struct {
		name string
		mix  GasMix
		ok   bool
	}{
		{"air assumption", GasMix{O2: 0.21, N2: 0.79}, true},
		{"trimix assumption", GasMix{O2: 0.18, He: 0.35, N2: 0.47}, true},
		{"sum mismatch", GasMix{O2: 0.21, N2: 0.7}, false},
		{"negative component", GasMix{O2: 0.21, He: -0.1, N2: 0.89}, false},
		{"outside model oxygen boundary", GasMix{O2: 0.12, He: 0.4, N2: 0.48}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.mix.Validate()
			if (err == nil) != test.ok {
				t.Fatalf("Validate() error = %v, ok=%t", err, test.ok)
			}
		})
	}
}
