package plant

import "math"

// OpenLoopStep drives a model with a constant input u0 for n steps and
// returns the recorded PV history. It is the canonical way to verify the
// steady-state gain and the time constant of a discretised model.
func OpenLoopStep(m Model, u0 float64, n int) []float64 {
	m.Reset()
	out := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, m.Step(u0))
	}
	return out
}

// SteadyGain estimates the DC gain from an open-loop step response as the
// ratio of the final PV to the input magnitude. The caller should run
// enough steps for the transient to settle (>= 8 time constants).
func SteadyGain(history []float64, u0 float64) float64 {
	if len(history) == 0 || u0 == 0 {
		return 0
	}
	return history[len(history)-1] / u0
}

// TimeConstant63 estimates the time constant as the first time the
// response reaches 63.2% of its final value, scaled by the sampling
// period. It is meaningful only for a monotone first-order response.
func TimeConstant63(history []float64, ts float64, u0 float64) float64 {
	if len(history) == 0 || ts <= 0 || u0 == 0 {
		return 0
	}
	final := history[len(history)-1]
	if final == 0 {
		return 0
	}
	target := 0.632 * final
	for i, v := range history {
		if math.Abs(v) >= math.Abs(target) {
			return float64(i+1) * ts
		}
	}
	return 0
}

// FOPDTStepClosed returns the expected PV after one step of the
// discretised FOPDT with a fresh input, used to validate the delay
// pipeline: with delay 0 the first PV is K(1-a)u.
func FOPDTStepClosed(gain, tau, ts, u float64) float64 {
	a := math.Exp(-ts / tau)
	return gain * (1 - a) * u
}

// GainKind reports whether a plant's open-loop gain is positive, negative
// or zero, which determines the sign convention of the controller.
func GainKind(g float64) string {
	switch {
	case g > 0:
		return "direct"
	case g < 0:
		return "reverse"
	default:
		return "zero"
	}
}
