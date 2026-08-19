package sim

import (
	"pid-loop/internal/plant"
)

// OpenLoop describes the plant's bare response to a step input: the final
// PV, the estimated DC gain and the 63.2% time constant. It is the
// reference the closed-loop metrics are compared against.
type OpenLoop struct {
	Steps         int
	FinalPV       float64
	GainEstimate  float64
	TimeConstantS float64
}

// AnalyseOpenLoop drives the plant with a unit step for n steps and
// extracts the characterisation. The config is validated first so the
// plant sees the shared sampling period.
func AnalyseOpenLoop(cfg *Config, n int) (OpenLoop, error) {
	if err := cfg.Validate(); err != nil {
		return OpenLoop{}, err
	}
	m, err := plant.Build(&cfg.Plant)
	if err != nil {
		return OpenLoop{}, err
	}
	hist := plant.OpenLoopStep(m, 1.0, n)
	return OpenLoop{
		Steps:         len(hist),
		FinalPV:       hist[len(hist)-1],
		GainEstimate:  plant.SteadyGain(hist, 1.0),
		TimeConstantS: plant.TimeConstant63(hist, cfg.Ts, 1.0),
	}, nil
}

// FinalErrorRatio reports |final error / setpoint|, the convergence
// metric behind the "steady PV approaches the setpoint" contract.
func FinalErrorRatio(res *Result) float64 {
	sp := res.Config.Setpoint
	if sp == 0 {
		return 0
	}
	return abs(res.Metrics.FinalError / sp)
}

// Settled reports whether the final PV sits inside the settling band, the
// check used to accept a run as converged.
func Settled(res *Result) bool {
	if res.Diverged {
		return false
	}
	band := res.Config.Band * abs(res.Config.Setpoint)
	return abs(res.Metrics.FinalError) <= band
}
