package cli

import (
	"fmt"
	"io"

	"pid-loop/internal/sim"
)

// writeMetrics prints the performance metrics of a run in a fixed order
// so the output is stable for scripts and tests.
func writeMetrics(w io.Writer, res *sim.Result) {
	m := res.Metrics
	fmt.Fprintf(w, "final PV       %.10f\n", m.FinalPV)
	fmt.Fprintf(w, "final error    %.10f\n", m.FinalError)
	fmt.Fprintf(w, "overshoot      %.6f   (fraction of |step|)\n", m.Overshoot)
	fmt.Fprintf(w, "settling time  %.6f s\n", m.SettlingTime)
	fmt.Fprintf(w, "rise time      %.6f s\n", m.RiseTime)
	fmt.Fprintf(w, "IAE            %.10f\n", m.IAE)
	if res.Diverged {
		fmt.Fprintf(w, "WARNING: loop diverged at step %d; parameters not acceptable\n", res.DivergedAt)
	}
}

// writeSummary prints a compact one-line metrics summary, used by the
// comparison tests to assert on trends without parsing the full block.
func writeSummary(w io.Writer, res *sim.Result) {
	m := res.Metrics
	fmt.Fprintf(w, "overshoot=%.6f settling=%.6f rise=%.6f iae=%.10f final=%.10f\n",
		m.Overshoot, m.SettlingTime, m.RiseTime, m.IAE, m.FinalPV)
}
