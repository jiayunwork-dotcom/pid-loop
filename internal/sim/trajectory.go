package sim

// Trajectory describes the salient points of the PV history: the peak, its
// time, the first undershoot and the final value.
type Trajectory struct {
	Peak        float64
	PeakTime    float64
	Undershoot  float64
	UndershootT float64
	Final       float64
	HasPeak     bool
}

// AnalyseTrajectory scans the series for the overshoot peak and the first
// undershoot relative to the setpoint.
func AnalyseTrajectory(series []Series, setpoint float64) Trajectory {
	tr := Trajectory{Final: series[len(series)-1].PV}
	for _, s := range series {
		if s.PV > tr.Peak {
			tr.Peak = s.PV
			tr.PeakTime = s.Time
			tr.HasPeak = true
		}
		if s.PV < tr.Undershoot && tr.HasPeak {
			tr.Undershoot = s.PV
			tr.UndershootT = s.Time
		}
	}
	return tr
}

// OvershootPct returns the overshoot as a percentage of the step for
// display purposes.
func (t Trajectory) OvershootPct(setpoint float64) float64 {
	step := abs(setpoint)
	if step == 0 {
		return 0
	}
	os := (t.Peak - setpoint) / step * 100
	if os < 0 {
		return 0
	}
	return os
}

// SettleBandTimes returns the entry and exit times of the band crossing:
// first time inside the band, last time outside (before the final settle).
// It backs the settling-time definition with the raw crossings.
func SettleBandTimes(series []Series, setpoint, band float64) (enter, lastExit float64, ok bool) {
	bandAbs := band * abs(setpoint)
	enter = -1
	for _, s := range series {
		if abs(s.PV-setpoint) <= bandAbs {
			if enter < 0 {
				enter = s.Time
			}
			ok = true
		} else {
			lastExit = s.Time
		}
	}
	return enter, lastExit, ok
}
