package pid

// Anti-windup strategy (documented in the README): conditional
// integration, i.e. freeze the integral accumulator while the output is
// saturated. The freeze decision lives in Compute; this file only exposes
// the bookkeeping helpers so the strategy is testable in isolation.

// WindupState captures the integrator condition at one sample.
type WindupState struct {
	Integral  float64
	Saturated bool
	AtLower   bool
	AtUpper   bool
}

// State reports the current integrator bookkeeping after a Compute call.
func (c *Controller) State() WindupState {
	return WindupState{
		Integral:  c.integral,
		Saturated: c.saturated,
		AtLower:   c.lastU <= c.UMin+1e-15,
		AtUpper:   c.lastU >= c.UMax-1e-15,
	}
}

// AntiWindupMode is the selected strategy name, exposed for the CLI help.
const AntiWindupMode = "integral-freeze (conditional integration)"

// FreezeIntegralForced returns the integral that would result if the
// accumulator grew with e*Ts while saturated; used by tests to quantify
// how much windup the freeze prevents.
func FreezeIntegralForced(current, e, ts float64) float64 {
	return current + e*ts
}
