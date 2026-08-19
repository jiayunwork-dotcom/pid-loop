package cli

import (
	"fmt"
	"io"
)

const usage = `pid-loop — single-loop PID closed-loop simulator

runs a position-form PID controller against a discrete plant (FOPDT with
dead time, or second order), records the PV/OP/error time history and
reports overshoot, settling time, rise time and IAE.

usage:
  pid-loop run [file]
    file            path to the simulation JSON (default: read stdin)

input JSON:
  { "setpoint": 1.0, "sim_steps": 400, "ts": 0.1,
    "band": 0.02,
    "controller": { "kp": 1.5, "ti": 3.0, "td": 0.2,
                    "umin": -1.0, "umax": 1.0 },
    "plant": { "type": "fopdt", "gain": 1.0, "tau": 3.0, "delay": 1.0 },
    "output": "trace.csv" }

  plant.type "fopdt": gain K, time constant tau, dead time delay (seconds)
  plant.type "second": gain K, damping zeta, natural frequency omega

  controller output is clamped to [umin, umax] with integral-freeze
  anti-windup; proportional and derivative act on the error.

examples:
  pid-loop run example/fopdt.json
  pid-loop run example/second.json
  cat example/fopdt.json | pid-loop run
`

// Run dispatches the first argument as a subcommand and returns a process
// exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "run":
		return RunRun(args[1:], stdout, stderr)
	case "compare":
		return RunCompare(args[1:], stdout, stderr)
	case "info":
		return RunInfo(args[1:], stdout, stderr)
	case "describe":
		return RunDescribe(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s\n", args[0], usage)
		return 2
	}
}
