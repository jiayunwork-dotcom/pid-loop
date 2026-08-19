package plant

// Second is a second-order plant in standard form
//
//	G(s) = K wn^2 / (s^2 + 2 zeta wn s + wn^2)
//
// discretised with the Tustin (bilinear) transform s = (2/Ts)(z-1)/(z+1).
// The resulting difference equation is
//
//	d2 y[k] = n2 u[k] + n1 u[k-1] + n0 u[k-2] - d1 y[k-1] - d0 y[k-2]
//
// with the coefficients derived in discretize.go. The transform keeps a
// stable continuous plant stable, so divergence in simulation means the
// closed loop is genuinely unstable, not the discretisation.
type Second struct {
	n2, n1, n0 float64
	d1, d0     float64
	u1, u2     float64
	y1, y2     float64
}

// NewSecond builds the discretised model from a validated configuration.
func NewSecond(p *Plant) *Second {
	// G(s) = b0 / (s^2 + a1 s + a0)
	b0 := p.Gain * p.Omega * p.Omega
	a1 := 2 * p.Zeta * p.Omega
	a0 := p.Omega * p.Omega
	n2, n1, n0, d1, d0, d2 := tustinCoeffs(b0, a1, a0, p.Ts)
	return &Second{
		n2: n2 / d2, n1: n1 / d2, n0: n0 / d2,
		d1: d1 / d2, d0: d0 / d2,
	}
}

// Step advances the plant by one sampling interval using the direct-form
// difference equation. The current input is available because the
// controller computes the OP before the plant advances (no physical
// computation delay is modelled beyond the FOPDT queue).
func (m *Second) Step(u float64) float64 {
	y := m.n2*u + m.n1*m.u1 + m.n0*m.u2 - m.d1*m.y1 - m.d0*m.y2
	m.u2, m.u1 = m.u1, u
	m.y2, m.y1 = m.y1, y
	return y
}

// Reset clears all history registers.
func (m *Second) Reset() {
	m.u1, m.u2 = 0, 0
	m.y1, m.y2 = 0, 0
}

// Kind implements Model.
func (m *Second) Kind() string { return "second" }
