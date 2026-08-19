package pid

import "fmt"

// Diagnostics summarises one controller run's windup behaviour.
type Diagnostics struct {
	StepsSaturated   int
	IntegralFinal    float64
	WindupIfUnfrozen float64
	PeakIntegral     float64
}

// Analyse runs a pure controller open-loop (no plant) against a constant
// error signal and reports how much the integral grows. With the freeze
// strategy the integral must stay bounded while the output is saturated;
// WindupIfUnfrozen shows what would happen without the freeze.
func Analyse(c *Controller, eConst float64, steps int) Diagnostics {
	ctrl, err := New(c) // New enables the freeze
	if err != nil {
		return Diagnostics{}
	}
	ctrl.Reset()
	d := Diagnostics{}
	unfrozen := 0.0
	for i := 0; i < steps; i++ {
		ctrl.Compute(0, -eConst) // sp=0, pv=-e -> e=eConst
		if ctrl.saturated {
			d.StepsSaturated++
			unfrozen += eConst * ctrl.Ts
		}
		if ctrl.integral > d.PeakIntegral {
			d.PeakIntegral = ctrl.integral
		}
	}
	d.IntegralFinal = ctrl.integral
	d.WindupIfUnfrozen = unfrozen
	return d
}

// Describe renders the diagnostics as text.
func (d Diagnostics) Describe() string {
	return fmt.Sprintf(
		"saturated %d steps | integral final %.4f | peak %.4f | windup if unfrozen %.4f",
		d.StepsSaturated, d.IntegralFinal, d.PeakIntegral, d.WindupIfUnfrozen)
}

// GainSchedule describes how the controller reacts to a gain change: it
// recomputes the output at a constant error for two kp values and reports
// both. It is the unit-level version of the Kp contract.
func GainSchedule(c *Controller, kpA, kpB, e float64) (ua, ub float64) {
	a := *c
	a.Kp = kpA
	b := *c
	b.Kp = kpB
	a.Reset()
	b.Reset()
	ua = a.Compute(0, -e)
	ub = b.Compute(0, -e)
	return ua, ub
}
