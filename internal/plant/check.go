package plant

// Check is the result of a plant sanity audit: its open-loop step
// response, DC gain estimate and the resolved delay stepping.
type Check struct {
	Kind          string
	GainEstimate  float64
	TimeConstantS float64
	DelaySteps    int
	Plausible     bool
	Note          string
}

// Audit runs an open-loop step for n steps and returns the plant check.
func Audit(p *Plant, n int) (Check, error) {
	if err := Validate(p); err != nil {
		return Check{}, err
	}
	m, err := Build(p)
	if err != nil {
		return Check{}, err
	}
	hist := OpenLoopStep(m, 1.0, n)
	ck := Check{
		Kind:          m.Kind(),
		GainEstimate:  SteadyGain(hist, 1.0),
		TimeConstantS: TimeConstant63(hist, p.Ts, 1.0),
		DelaySteps:    p.DelaySteps,
	}
	ck.Plausible = ck.GainEstimate != 0 && ck.GainEstimate == p.Gain
	if ck.Plausible {
		ck.Note = "open-loop gain matches the configured K"
	} else {
		ck.Note = "open-loop gain deviates from K; check discretisation"
	}
	return ck, nil
}

// DelayTable lists the resolved delay steps for a range of dead times.
type DelayTable struct {
	Ts     float64
	Delays []float64
	Steps  []int
}

// BuildDelayTable resolves each dead time against the sampling period.
func BuildDelayTable(ts float64, delays []float64) DelayTable {
	t := DelayTable{Ts: ts, Delays: delays}
	for _, d := range delays {
		t.Steps = append(t.Steps, StepsForDelay(d, ts))
	}
	return t
}

// Format renders the delay table as text lines.
func (t DelayTable) Format() []string {
	out := make([]string, 0, len(t.Delays))
	for i, d := range t.Delays {
		out = append(out, formatDelayRow(d, t.Steps[i]))
	}
	return out
}
