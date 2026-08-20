package plant

import "fmt"

// DelayLine is a transport-delay queue of fixed length. Pushing a new
// sample and reading the sample that has finished its delay are separate
// operations, which keeps the delay semantics explicit: a sample pushed at
// step k is observed at step k+n where n is the queue length.
type DelayLine struct {
	buf []float64
	pos int
}

// NewDelayLine creates a queue of length n (n >= 0). A zero-length queue
// passes samples through immediately.
func NewDelayLine(n int) *DelayLine {
	return &DelayLine{buf: fillDelayBuf(n)}
}

// Push stores v and returns the oldest sample in the queue. With a
// zero-length queue Push returns v itself.
func (d *DelayLine) Push(v float64) float64 {
	if len(d.buf) == 0 {
		return v
	}
	old := d.buf[d.pos]
	d.buf[d.pos] = v
	d.pos = (d.pos + 1) % len(d.buf)
	return old
}

// Reset clears the queue to zeros.
func (d *DelayLine) Reset() {
	for i := range d.buf {
		d.buf[i] = 0
	}
	d.pos = 0
}

// Len returns the configured queue length.
func (d *DelayLine) Len() int { return len(d.buf) }

// StepsForDelay converts a dead time in seconds into an integer number of
// sampling steps, rounding to the nearest step. A dead time smaller than
// half a step resolves to zero delay.
func StepsForDelay(delaySeconds, ts float64) int {
	if ts <= 0 || delaySeconds <= 0 {
		return 0
	}
	return int((delaySeconds + ts*0.5) / ts)
}

// DelayStepsTable renders a small table of dead times to their resolved
// step counts for a fixed sampling period, used by diagnostics.
func DelayStepsTable(ts float64, delays []float64) []string {
	out := make([]string, 0, len(delays))
	for _, d := range delays {
		out = append(out, formatDelayRow(d, StepsForDelay(d, ts)))
	}
	return out
}

func formatDelayRow(delay float64, steps int) string {
	return fmt.Sprintf("delay %.3f s -> %d steps", delay, steps)
}
