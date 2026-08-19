// Package pid implements the position-form PID controller used by the
// closed-loop simulator: proportional on error, integral accumulation with
// e*Ts/Ti scaling, derivative on error, output clamping and conditional
// integral freeze as the anti-windup strategy (all documented in the
// README). The controller shares the sampling period and the error
// definition with the plant and the metric layer, so no unit conversion
// happens anywhere else.
package pid

// Controller is a position-form PID with integral freeze anti-windup.
//
//	u = Kp e + (Kp/Ti) * sum(e*Ts) + (Kp Td) * de/dt
//
// The output is clamped to [UMin, UMax]. While the unclamped output lies
// outside the limits, the integral accumulator is frozen (conditional
// integration), so the controller does not wind up during saturation.
type Controller struct {
	Kp   float64 `json:"kp"`
	Ti   float64 `json:"ti"`
	Td   float64 `json:"td"`
	UMin float64 `json:"umin"`
	UMax float64 `json:"umax"`
	Ts   float64 `json:"-"`

	// FreezeOnSaturation enables the conditional-integration anti-windup.
	// It is settable only from code (never JSON) so experiments can toggle
	// the strategy for comparison; New defaults it to true.
	FreezeOnSaturation bool `json:"-"`

	integral  float64
	ePrev     float64
	hasPrev   bool
	saturated bool
	lastU     float64
}

// New builds a controller and applies validation to the gains, limits and
// sampling period. The integral-freeze anti-windup is always enabled by
// New; NewPreserve keeps the caller's FreezeOnSaturation setting for
// windup comparison experiments.
func New(c *Controller) (*Controller, error) {
	if err := Validate(c); err != nil {
		return nil, err
	}
	out := *c
	out.FreezeOnSaturation = true
	out.Reset()
	return &out, nil
}

// NewPreserve is like New but keeps the caller's FreezeOnSaturation flag,
// which is how the no-freeze variant of the anti-windup contract is built.
func NewPreserve(c *Controller) (*Controller, error) {
	if err := Validate(c); err != nil {
		return nil, err
	}
	out := *c
	out.Reset()
	return &out, nil
}

// Reset clears the integral accumulator and the derivative memory.
func (c *Controller) Reset() {
	c.integral = 0
	c.ePrev = 0
	c.hasPrev = false
	c.saturated = false
}

// Saturated reports whether the last Compute call hit an output limit.
func (c *Controller) Saturated() bool { return c.saturated }

// Integral exposes the accumulated integral term for diagnostics.
func (c *Controller) Integral() float64 { return c.integral }
