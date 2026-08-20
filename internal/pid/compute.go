package pid

// Clamp restricts v to the closed interval [lo, hi]. It is the single
// place where output limits are applied, so every caller (the controller,
// the anti-windup logic and the tests) observes identical clamping.
func Clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Compute produces the clamped manipulated variable for one sampling
// interval from the setpoint and the current process variable:
//
//	e = sp - pv
//	u = Kp*e + (Kp/Ti)*integral + (Kp*Td)*de/dt
//
// The derivative is a backward difference on the error (so a setpoint
// step produces a derivative kick, which the README documents as the
// chosen convention). The integral accumulator grows by e*Ts only while
// the unclamped output stays inside the limits; while the output is
// saturated the accumulator is frozen to prevent windup.
func (c *Controller) Compute(sp, pv float64) float64 {
	e := sp - pv
	de := 0.0
	if c.hasPrev {
		de = (e - c.ePrev) / c.Ts
	}
	proportional, integralTerm, derivative := applyTerms(c, e, de)
	uRaw := proportional + integralTerm + derivative
	u := Clamp(uRaw, c.UMin, c.UMax)

	sat := u != uRaw
	if !sat || !c.FreezeOnSaturation {
		c.integral += e * c.Ts
	}
	c.saturated = sat
	c.lastU = u
	c.ePrev = e
	c.hasPrev = true
	return u
}

// Outputs splits the last computed u into its three contributions,
// available for diagnostics and for tests that verify the anti-windup
// behaviour term by term.
func (c *Controller) Outputs(sp, pv float64) (p, i, d, u float64) {
	e := sp - pv
	de := 0.0
	if c.hasPrev {
		de = (e - c.ePrev) / c.Ts
	}
	p = c.Kp * e
	i = (c.Kp / c.Ti) * c.integral
	d = c.Kp * c.Td * de
	u = Clamp(p+i+d, c.UMin, c.UMax)
	return p, i, d, u
}
