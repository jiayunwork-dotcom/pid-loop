package cli

import (
	"fmt"
	"io"
	"os"

	"pid-loop/internal/plant"
	"pid-loop/internal/sim"
)

// RunDescribe executes the describe subcommand: it audits the plant's
// open-loop response, runs the closed loop and prints the trajectory and
// oscillation characterisation alongside the standard metrics.
func RunDescribe(args []string, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", args[1])
		return 2
	}
	var cfg *sim.Config
	var err error
	if len(args) == 1 {
		cfg, err = sim.ParseFile(args[0])
	} else {
		cfg, err = sim.Parse(os.Stdin)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(stderr, "error: invalid config: %v\n", err)
		return 1
	}

	ck, err := plant.Audit(&cfg.Plant, 500)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "open-loop: kind=%s gain=%.6f (configured %.6f) tau63=%.4f s steps=%d\n",
		ck.Kind, ck.GainEstimate, cfg.Plant.Gain, ck.TimeConstantS, ck.DelaySteps)
	fmt.Fprintf(stdout, "  %s\n", ck.Note)

	ol, err := sim.AnalyseOpenLoop(cfg, 500)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "open-loop final PV=%.6f\n", ol.FinalPV)

	res, err := sim.Simulate(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if res.Diverged {
		fmt.Fprintln(stdout, res.Summary())
		return 0
	}
	writeMetrics(stdout, res)

	tr := sim.AnalyseTrajectory(res.Series, cfg.Setpoint)
	fmt.Fprintf(stdout, "peak PV=%.6f at t=%.4f s (overshoot %.2f%%)\n",
		tr.Peak, tr.PeakTime, tr.OvershootPct(cfg.Setpoint))

	osc := sim.AnalyseOscillation(res.Series, cfg.Setpoint)
	fmt.Fprintf(stdout, "oscillation: %d crossings", osc.Crossings)
	if osc.HasDecay {
		fmt.Fprintf(stdout, ", decay ratio %.4f", osc.DecayRatio)
	}
	fmt.Fprintln(stdout)

	cv := sim.AnalyseConvergence(res)
	fmt.Fprintf(stdout, "convergence: settled=%v rel-error=%.6f\n", cv.Settled, cv.RelativeError)
	return 0
}
