package sim

// Oscillation summarises the oscillatory behaviour of a response: how many
// times the PV crossed the setpoint and how the second peak compares to
// the first (decay ratio). A well-tuned loop settles with few crossings
// and a decay ratio well below one.
type Oscillation struct {
	Crossings  int
	DecayRatio float64
	HasDecay   bool
}

// AnalyseOscillation counts setpoint crossings and estimates the decay
// ratio from the first two positive overshoot peaks of the error envelope.
func AnalyseOscillation(series []Series, setpoint float64) Oscillation {
	xs := Crossings(series, setpoint)
	decay, has := decayRatio(series, setpoint)
	return Oscillation{
		Crossings:  len(xs),
		DecayRatio: decay,
		HasDecay:   has,
	}
}

// decayRatio finds the first two local maxima of |PV - setpoint| after
// the initial transient and returns their ratio. Fewer than two peaks
// means no measurable oscillation.
func decayRatio(series []Series, setpoint float64) (float64, bool) {
	var peaks []float64
	increasing := false
	var cur float64
	for _, s := range series {
		e := abs(s.PV - setpoint)
		if e > cur {
			increasing = true
			cur = e
		} else if increasing {
			peaks = append(peaks, cur)
			increasing = false
			cur = e
			if len(peaks) >= 3 {
				break
			}
		}
	}
	if len(peaks) < 2 || peaks[0] == 0 {
		return 0, false
	}
	return peaks[1] / peaks[0], true
}

// PeakCount returns the number of distinct positive overshoot peaks.
func PeakCount(series []Series, setpoint float64) int {
	count := 0
	increasing := false
	cur := 0.0
	for _, s := range series {
		e := abs(s.PV - setpoint)
		if e > cur {
			increasing = true
			cur = e
		} else if increasing {
			count++
			increasing = false
			cur = e
		}
	}
	return count
}
