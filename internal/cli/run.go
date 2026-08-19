package cli

import (
	"fmt"
	"io"
	"os"

	"pid-loop/internal/sim"
)

// RunRun executes the run subcommand: parse the config, simulate the
// closed loop, print a bounded time summary plus the metrics, and write
// the full series to the configured output file when present.
func RunRun(args []string, stdout, stderr io.Writer) int {
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
	res, err := sim.Simulate(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if cfg.Output != "" {
		if err := sim.WriteCSV(res.Series, cfg.Output); err != nil {
			fmt.Fprintf(stderr, "error: write %s: %v\n", cfg.Output, err)
			return 1
		}
		fmt.Fprintf(stdout, "trace written to %s\n", cfg.Output)
	}
	if res.Diverged {
		fmt.Fprintln(stdout, res.Summary())
		return 0
	}
	writeMetrics(stdout, res)
	fmt.Fprintln(stdout, "time history (head/tail):")
	for _, line := range sim.HeadTail(res.Series, 4) {
		fmt.Fprintln(stdout, "  "+line)
	}
	return 0
}
