package cli

import (
	"fmt"
	"io"
	"os"

	"pid-loop/internal/pid"
	"pid-loop/internal/plant"
	"pid-loop/internal/sim"
)

// RunInfo executes the info subcommand: it parses a config and prints the
// resolved plant and controller settings together with the selected
// conventions (position form, error-based P and D, integral freeze) so a
// user can verify what a run will actually do.
func RunInfo(args []string, stdout, stderr io.Writer) int {
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
	p := &cfg.Plant
	fmt.Fprintf(stdout, "plant: type=%s gain=%.6f\n", p.Type, p.Gain)
	switch p.Type {
	case plant.TypeFOPDT:
		fmt.Fprintf(stdout, "  tau=%.6f s  delay=%.6f s -> %d steps (ts=%.6f)\n",
			p.Tau, p.Delay, p.DelaySteps, cfg.Ts)
	case plant.TypeSecond:
		fmt.Fprintf(stdout, "  zeta=%.6f  omega=%.6f rad/s\n", p.Zeta, p.Omega)
	}
	c := cfg.Controller
	fmt.Fprintf(stdout, "controller: kp=%.6f ti=%.6f td=%.6f umin=%.6f umax=%.6f\n",
		c.Kp, c.Ti, c.Td, c.UMin, c.UMax)
	fmt.Fprintf(stdout, "  form=%s anti-windup=%s\n", "position (P and D on error)", pid.AntiWindupMode)
	fmt.Fprintf(stdout, "sim: ts=%.6f s steps=%d band=%.4f setpoint=%.6f\n",
		cfg.Ts, cfg.SimSteps, cfg.Band, cfg.Setpoint)
	return 0
}
