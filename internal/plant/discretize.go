package plant

// tustinCoeffs derives the Tustin difference-equation coefficients for the
// continuous plant b0 / (s^2 + a1 s + a0):
//
//	G(z) = b0 c^2 (z+1)^2 / ((z-1)^2 + a1 c (z^2-1) + a0 c^2 (z+1)^2)
//	c = Ts/2
//
// It returns the numerator coefficients n2, n1, n0 and the denominator
// coefficients d1, d0, d2; the caller normalises by d2.
func tustinCoeffs(b0, a1, a0, ts float64) (n2, n1, n0, d1, d0, d2 float64) {
	c := ts / 2
	c2 := c * c
	// Denominator of the bilinear transform applied to s^2 + a1 s + a0.
	d2 = 1 + a1*c + a0*c2
	d1 = -2 + 2*a0*c2
	d0 = 1 - a1*c + a0*c2
	// Numerator: b0 c^2 (z+1)^2.
	n2 = b0 * c2
	n1 = 2 * b0 * c2
	n0 = b0 * c2
	return n2, n1, n0, d1, d0, d2
}

// DCGainSecond returns the steady-state gain K of the standard-form second
// order plant, used by metrics to compute the expected final PV.
func (m *Second) DCGain() float64 {
	// With the normalised coefficients, the DC gain from the difference
	// equation is (n2+n1+n0)/(1+d1+d0).
	num := m.n2 + m.n1 + m.n0
	den := 1 + m.d1 + m.d0
	if den == 0 {
		return 0
	}
	return num / den
}
