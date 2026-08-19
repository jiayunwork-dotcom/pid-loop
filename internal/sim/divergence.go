package sim

// DivergenceThreshold returns the envelope beyond which a signal counts as
// divergent: DivergenceFactor times the step magnitude. A loop whose PV or
// OP leaves this envelope is reported as diverged, never as tuned.
func DivergenceThreshold(setpoint float64) float64 {
	return DivergenceFactor * abs(setpoint)
}

// CheckDivergence scans a recorded series and reports the first step whose
// PV or OP magnitude exceeds the divergence threshold. A negative index
// means the run stayed inside the envelope.
func CheckDivergence(series []Series, setpoint float64) int {
	limit := DivergenceThreshold(setpoint)
	for i, s := range series {
		if abs(s.PV) > limit || abs(s.OP) > limit {
			return i
		}
	}
	return -1
}

// PeakPV returns the maximum absolute process variable of the series.
func PeakPV(series []Series) float64 {
	peak := 0.0
	for _, s := range series {
		if abs(s.PV) > peak {
			peak = abs(s.PV)
		}
	}
	return peak
}
