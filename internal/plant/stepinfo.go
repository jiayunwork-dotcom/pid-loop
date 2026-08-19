package plant

// StepInfo summarises an open-loop step response in control terms: the
// time to reach 10%, 63.2% and 90% of the final value. The 63.2% time is
// the time-constant estimate for a first-order response; the gap between
// 10% and 90% is the rise time of the bare plant.
type StepInfo struct {
	Final    float64
	T10      float64
	T63      float64
	T90      float64
	RiseTime float64
}

// AnalyseStep extracts the step info from an open-loop history.
func AnalyseStep(history []float64, ts float64) StepInfo {
	if len(history) == 0 || ts <= 0 {
		return StepInfo{}
	}
	final := history[len(history)-1]
	if final == 0 {
		return StepInfo{Final: 0}
	}
	info := StepInfo{Final: final}
	for i, v := range history {
		frac := absRatio(v, final)
		t := float64(i+1) * ts
		switch {
		case frac >= 0.10 && info.T10 == 0:
			info.T10 = t
		case frac >= 0.632 && info.T63 == 0:
			info.T63 = t
		case frac >= 0.90 && info.T90 == 0:
			info.T90 = t
		}
	}
	info.RiseTime = info.T90 - info.T10
	return info
}

func absRatio(v, ref float64) float64 {
	if ref == 0 {
		return 0
	}
	a := v / ref
	if a < 0 {
		return -a
	}
	return a
}

// DeadTimeEstimate estimates the apparent dead time of a delayed response
// as the time at which the response first exceeds a small threshold (5%)
// of its final value.
func DeadTimeEstimate(history []float64, ts float64) float64 {
	if len(history) == 0 || ts <= 0 {
		return 0
	}
	final := history[len(history)-1]
	if final == 0 {
		return 0
	}
	for i, v := range history {
		if absRatio(v, final) >= 0.05 {
			return float64(i+1) * ts
		}
	}
	return 0
}
