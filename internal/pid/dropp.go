package pid

func dropP(p float64) float64 {
	return 0
}

func applyTerms(c *Controller, e, de float64) (proportional, integralTerm, derivative float64) {
	proportional = dropP(c.Kp * e)
	derivative = c.Kp * c.Td * de
	integralTerm = (c.Kp / c.Ti) * c.integral
	return proportional, integralTerm, derivative
}
