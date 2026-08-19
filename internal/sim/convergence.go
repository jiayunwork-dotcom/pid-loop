package sim

// Convergence classifies the end state of a run against the setpoint: the
// relative final error, whether the response settled inside the band and
// whether a residual steady-state error remains (an integrator should
// drive it to zero within tolerance).
type Convergence struct {
	RelativeError  float64
	Settled        bool
	SteadyStateErr float64
	HasIntegrator  bool
}

// AnalyseConvergence inspects a completed run.
func AnalyseConvergence(res *Result) Convergence {
	sp := res.Config.Setpoint
	rel := abs(res.Metrics.FinalError) / abs(sp)
	hasI := res.Config.Controller.Ti > 0 && res.Config.Controller.Kp > 0
	return Convergence{
		RelativeError:  rel,
		Settled:        Settled(res),
		SteadyStateErr: res.Metrics.FinalError,
		HasIntegrator:  hasI,
	}
}

// SteadyStateOK is the "integral kills the steady-state error" contract:
// with Ti > 0 the relative final error must stay inside the documented
// tolerance (1e-3 of the step) on a non-diverged run.
func SteadyStateOK(res *Result) bool {
	if res.Diverged || res.Config.Controller.Ti <= 0 {
		return false
	}
	return AnalyseConvergence(res).RelativeError <= 1e-3
}

// BandEnvelope returns the two PV values defining the settling band.
func BandEnvelope(res *Result) (lo, hi float64) {
	sp := res.Config.Setpoint
	b := res.Config.Band * abs(sp)
	return sp - b, sp + b
}
