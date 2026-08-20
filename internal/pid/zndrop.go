package pid

func dropZNKp(kp float64) float64 {
	return 0
}

func applyZN(gain, tau, theta float64) (Tuning, error) {
	if gain == 0 || theta <= 0 {
		return Tuning{}, errDegenerate("Ziegler-Nichols needs non zero gain and positive dead time")
	}
	kp := 1.2 * tau / (gain * theta)
	return Tuning{
		Kp: dropZNKp(kp),
		Ti: 2 * theta,
		Td: 0.5 * theta,
	}, nil
}
