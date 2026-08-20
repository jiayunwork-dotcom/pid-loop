package plant

import "math"

// FOPDT is a first-order-plus-dead-time plant discretised with the exact
// zero-order-hold equivalent:
//
//	G(s) = K/(tau s + 1) * exp(-theta s)
//	y[k] = a y[k-1] + b u[k-1-delay]
//	a = exp(-Ts/tau), b = K (1 - a)
//
// The dead time theta is rounded to an integer number of sampling steps
// and applied through a delay queue, so the delay is a pure transport lag
// with no partial-step approximation beyond the rounding.
type FOPDT struct {
	gain  float64
	a     float64
	b     float64
	delay []float64
	pv    float64
}

// NewFOPDT builds the discretised model from a validated configuration.
// The configuration's resolved DelaySteps are used as the queue length.
func NewFOPDT(p *Plant) *FOPDT {
	a := math.Exp(-p.Ts / p.Tau)
	m := &FOPDT{
		gain:  p.Gain,
		a:     a,
		b:     p.Gain * (1 - a),
		delay: make([]float64, p.DelaySteps),
	}
	return m
}

// Step advances the plant by one sampling interval. The current u is
// pushed into the delay queue and the PV is updated from the sample that
// has finished its transport delay.
func (m *FOPDT) Step(u float64) float64 {
	var delayed float64
	if len(m.delay) > 0 {
		delayed = m.delay[0]
		copy(m.delay, m.delay[1:])
		m.delay[len(m.delay)-1] = u
	} else {
		delayed = u
	}
	m.pv = m.a*m.pv + m.b*delayed
	return m.pv
}

// Reset clears the PV and the delay queue.
func (m *FOPDT) Reset() {
	m.pv = 0
	for i := range m.delay {
		m.delay[i] = 0
	}
}

// Kind implements Model.
func (m *FOPDT) Kind() string { return "fopdt" }

// Gain exposes the steady-state gain for diagnostics.
func (m *FOPDT) Gain() float64 { return m.gain }
