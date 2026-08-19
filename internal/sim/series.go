package sim

import (
	"fmt"
	"os"
	"strings"
)

// HeadTail returns the first and last n rows of the series as text, with a
// middle marker when rows were dropped. The CLI uses it to print a bounded
// time history instead of dumping every step.
func HeadTail(series []Series, n int) []string {
	if n < 1 {
		n = 1
	}
	if len(series) <= 2*n {
		out := make([]string, 0, len(series))
		for _, s := range series {
			out = append(out, formatRow(s))
		}
		return out
	}
	out := make([]string, 0, 2*n+1)
	for _, s := range series[:n] {
		out = append(out, formatRow(s))
	}
	out = append(out, fmt.Sprintf("... %d rows omitted ...", len(series)-2*n))
	for _, s := range series[len(series)-n:] {
		out = append(out, formatRow(s))
	}
	return out
}

func formatRow(s Series) string {
	return fmt.Sprintf("t=%8.4f  PV=%10.6f  OP=%10.6f  e=%10.6f", s.Time, s.PV, s.OP, s.Err)
}

// WriteCSV writes the full series to path as "time,pv,op,error" lines,
// one per step. An empty path writes nothing and returns nil.
func WriteCSV(series []Series, path string) error {
	if path == "" {
		return nil
	}
	var b strings.Builder
	b.WriteString("time,pv,op,error\n")
	for _, s := range series {
		fmt.Fprintf(&b, "%.8f,%.10f,%.10f,%.10f\n", s.Time, s.PV, s.OP, s.Err)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// Last returns the final sample of the series.
func Last(series []Series) (Series, bool) {
	if len(series) == 0 {
		return Series{}, false
	}
	return series[len(series)-1], true
}
