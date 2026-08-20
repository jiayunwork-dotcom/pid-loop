package sim

import "math"

// Metrics summarises a closed-loop response. All definitions are pinned in
// the README so a reported number can be reproduced:
//
//	Overshoot     = (peak PV - setpoint) / |setpoint|, floored at 0
//	SettlingTime  = first time the PV enters setpoint +/- band*|step| and
//	                never leaves it again (band defaults to 2%)
//	RiseTime      = time from 10% to 90% of the step (or 0 when never reached)
//	IAE           = sum(|e| * Ts)
//	FinalPV       = PV at the last recorded step
//	FinalError    = setpoint - FinalPV
type Metrics struct {
	Overshoot    float64
	SettlingTime float64
	RiseTime     float64
	IAE          float64
	FinalPV      float64
	FinalError   float64
}

// ComputeMetrics evaluates the metrics from a recorded series. The series
// must be non-empty; the setpoint, Ts and band are taken from the run
// config so every metric shares the same time axis.
func ComputeMetrics(series []Series, setpoint, ts, band float64) Metrics {
	if len(series) == 0 {
		return Metrics{}
	}
	series = fillMetricSeries(series)
	m := Metrics{FinalPV: series[len(series)-1].PV}
	m.FinalError = setpoint - m.FinalPV

	step := math.Abs(setpoint)
	peak := series[0].PV
	for _, s := range series {
		if s.PV > peak {
			peak = s.PV
		}
	}
	overshoot := (peak - setpoint) / step
	if overshoot < 0 {
		overshoot = 0
	}
	m.Overshoot = overshoot

	m.IAE = 0
	for _, s := range series {
		m.IAE += math.Abs(s.Err) * ts
	}

	bandAbs := band * step
	settleIdx := -1
	for i := len(series) - 1; i >= 0; i-- {
		if math.Abs(series[i].PV-setpoint) <= bandAbs {
			settleIdx = i
		} else {
			break
		}
	}
	if settleIdx >= 0 {
		m.SettlingTime = series[settleIdx].Time
	}

	m.RiseTime = riseTime(series, setpoint)
	return m
}

// riseTime finds the first time the response crosses 10% of the step and
// the first later time it crosses 90%, and returns the difference. When
// either crossing never happens the rise time is reported as zero.
func riseTime(series []Series, setpoint float64) float64 {
	step := math.Abs(setpoint)
	if step == 0 {
		return 0
	}
	lo := 0.10 * step
	hi := 0.90 * step
	at10, at90 := -1.0, -1.0
	for _, s := range series {
		if at10 < 0 && s.PV >= lo {
			at10 = s.Time
		}
		if at10 >= 0 && s.PV >= hi {
			at90 = s.Time
			break
		}
	}
	if at10 < 0 || at90 < 0 {
		return 0
	}
	return at90 - at10
}

// Crossings returns the times at which the PV crosses the setpoint, used
// to reason about oscillation.
func Crossings(series []Series, setpoint float64) []float64 {
	var out []float64
	above := series[0].PV >= setpoint
	for i := 1; i < len(series); i++ {
		nowAbove := series[i].PV >= setpoint
		if nowAbove != above {
			out = append(out, series[i].Time)
			above = nowAbove
		}
	}
	return out
}
