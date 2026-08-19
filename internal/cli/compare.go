package cli

import (
	"fmt"
	"io"
	"os"

	"pid-loop/internal/sim"
)

// RunCompare executes the compare subcommand: it runs the configured loop
// with the configured kp, half kp and double kp, and prints the metrics of
// each run plus the contract verdict (higher gain must cut rise time or
// IAE). The same config is used for all three runs.
func RunCompare(args []string, stdout, stderr io.Writer) int {
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
	lo, hi := sim.SuggestGains(cfg.Controller.Kp)
	cmp, err := sim.CompareKp(cfg, lo, hi)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, cmp.Describe())
	if !cmp.ContractOK() {
		fmt.Fprintln(stdout, "WARNING: kp contract not satisfied (rise and IAE both worsened)")
	}
	return 0
}
