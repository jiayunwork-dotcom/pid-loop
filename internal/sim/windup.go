package sim

import "fmt"

// WindupImpact quantifies how much the anti-windup strategy changes the
// response: it compares the IAE of the loop with the integral freeze
// enabled and with it disabled. The two runs use identical parameters, so
// any difference in the metrics comes from the windup behaviour alone.
type WindupImpact struct {
	IAEWithFreeze    float64
	IAEWithoutFreeze float64
	Ratio            float64
	FinalWith        float64
	FinalWithout     float64
}

// CompareWindup runs the same config with and without the integral freeze.
// On a saturating run the no-freeze variant must accumulate extra
// integral, so its IAE is larger (or equal): that is the observable
// contract of the anti-windup design.
func CompareWindup(cfg *Config) (WindupImpact, error) {
	with, err := Simulate(cfg)
	if err != nil {
		return WindupImpact{}, err
	}
	without, err := SimulateNoFreeze(cfg)
	if err != nil {
		return WindupImpact{}, err
	}
	w := WindupImpact{
		IAEWithFreeze:    with.Metrics.IAE,
		IAEWithoutFreeze: without.Metrics.IAE,
		FinalWith:        with.Metrics.FinalPV,
		FinalWithout:     without.Metrics.FinalPV,
	}
	if w.IAEWithFreeze > 0 {
		w.Ratio = w.IAEWithoutFreeze / w.IAEWithFreeze
	}
	return w, nil
}

// Describe renders the impact as text.
func (w WindupImpact) Describe() string {
	return fmt.Sprintf("IAE freeze=%.6f no-freeze=%.6f ratio=%.3f",
		w.IAEWithFreeze, w.IAEWithoutFreeze, w.Ratio)
}
