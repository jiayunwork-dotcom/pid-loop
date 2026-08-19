package pid

import "math"

// Tuning is a set of PID parameters suggested by a tuning rule.
type Tuning struct {
	Kp float64
	Ti float64
	Td float64
}

// ZieglerNichols suggests controller parameters from the open-loop
// process data (gain K, time constant tau, dead time theta) using the
// classic Ziegler-Nichols open-loop tuning table:
//
//	Kp = 1.2 tau / (K theta)
//	Ti = 2 theta
//	Td = 0.5 theta
//
// It is a heuristic starting point; the README documents that the metric
// contract must still be checked by simulation.
func ZieglerNichols(gain, tau, theta float64) (Tuning, error) {
	if gain == 0 || theta <= 0 {
		return Tuning{}, errDegenerate("Ziegler-Nichols needs non zero gain and positive dead time")
	}
	return Tuning{
		Kp: 1.2 * tau / (gain * theta),
		Ti: 2 * theta,
		Td: 0.5 * theta,
	}, nil
}

// LambdaTuning (lambda method) places the closed-loop pole at 1/lambda
// seconds for a FOPDT plant:
//
//	Kp = tau / (K (lambda + theta))
//	Ti = tau
//	Td = 0
//
// Larger lambda yields a more sluggish but more robust loop.
func LambdaTuning(gain, tau, theta, lambda float64) (Tuning, error) {
	if gain == 0 || lambda <= 0 {
		return Tuning{}, errDegenerate("lambda tuning needs non zero gain and positive lambda")
	}
	return Tuning{
		Kp: tau / (gain * (lambda + theta)),
		Ti: tau,
		Td: 0,
	}, nil
}

// TuningRules lists the available rule names for CLI help.
func TuningRules() []string {
	return []string{"ziegler-nichols", "lambda"}
}

// errDegenerate is a small helper for tuning failures.
func errDegenerate(msg string) error {
	return &degenerateError{msg: msg}
}

type degenerateError struct{ msg string }

func (e *degenerateError) Error() string { return e.msg }

// Degenerate reports whether a tuning result is unusable (non positive
// gain or negative integral time).
func (t Tuning) Degenerate() bool {
	return math.IsNaN(t.Kp) || t.Kp <= 0 || t.Ti <= 0 || t.Td < 0
}
