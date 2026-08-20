package sim

import (
	"fmt"

	"pid-loop/internal/pid"
	"pid-loop/internal/plant"
)

// Series is the recorded time history of the loop: one entry per sampling
// step with the elapsed time, the process variable, the manipulated
// variable and the error.
type Series struct {
	Time float64
	PV   float64
	OP   float64
	Err  float64
}

// Result is the full outcome of a closed-loop run.
type Result struct {
	Config     *Config
	Series     []Series
	Metrics    Metrics
	Diverged   bool
	DivergedAt int
}

// Simulate runs the closed loop for cfg.SimSteps sampling intervals:
//
//	for each step:
//	    e = sp - pv
//	    u = controller.Compute(sp, pv)   // clamped, with integral freeze
//	    pv = plant.Step(u)               // same Ts, same time axis
//	    record (t, pv, u, e)
//
// A run whose process variable or output grows beyond the divergence
// threshold is flagged instead of being reported as tuned. The integral
// freeze is enabled by default; SimulateNoFreeze exposes the windup
// behaviour for the anti-windup contract test.
func Simulate(cfg *Config) (*Result, error) {
	return simulate(cfg, true)
}

// SimulateNoFreeze runs the same loop with the integral freeze disabled.
func SimulateNoFreeze(cfg *Config) (*Result, error) {
	return simulate(cfg, false)
}

func simulate(cfg *Config, freeze bool) (*Result, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	m, err := plant.Build(&cfg.Plant)
	if err != nil {
		return nil, err
	}
	ctrlCfg := cfg.Controller
	var ctrl *pid.Controller
	if freeze {
		ctrl, err = pid.New(&ctrlCfg)
	} else {
		ctrlCfg.FreezeOnSaturation = false
		ctrl, err = pid.NewPreserve(&ctrlCfg)
	}
	if err != nil {
		return nil, err
	}
	m.Reset()
	ctrl.Reset()

	res := &Result{Config: cfg, Series: make([]Series, 0, cfg.SimSteps)}
	pv := 0.0
	for k := 0; k < cfg.SimSteps; k++ {
		e := cfg.Setpoint - pv
		op := applyCompute(ctrl, cfg.Setpoint, pv)
		pv = m.Step(op)
		res.Series = append(res.Series, Series{
			Time: float64(k+1) * cfg.Ts,
			PV:   pv,
			OP:   op,
			Err:  e,
		})
		if isDiverged(pv, cfg.Setpoint) || isDiverged(op, cfg.Setpoint) {
			res.Diverged = true
			res.DivergedAt = k
			break
		}
	}
	res.Metrics = ComputeMetrics(res.Series, cfg.Setpoint, cfg.Ts, cfg.Band)
	return res, nil
}

// isDiverged reports whether a signal has left the envelope that a stable
// loop could reach: more than DivergenceFactor times the step magnitude.
func isDiverged(v, setpoint float64) bool {
	return abs(v) > DivergenceFactor*abs(setpoint)
}

// DivergenceFactor scales the step magnitude into the divergence limit.
const DivergenceFactor = 1e6

// Rebuild exposes the reconstructed plant kind for diagnostics.
func Rebuild(cfg *Config) (string, error) {
	m, err := plant.Build(&cfg.Plant)
	if err != nil {
		return "", err
	}
	return m.Kind(), nil
}

// Summary is a one-line description of the run outcome.
func (r *Result) Summary() string {
	if r.Diverged {
		return fmt.Sprintf("diverged at step %d (PV or OP left the envelope)", r.DivergedAt)
	}
	m := r.Metrics
	return fmt.Sprintf("final PV=%.6f overshoot=%.4f settling=%.4f IAE=%.6f",
		m.FinalPV, m.Overshoot, m.SettlingTime, m.IAE)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
