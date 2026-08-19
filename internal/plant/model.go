// Package plant models the process being controlled: a first-order-plus-
// dead-time plant (FOPDT) with an integer-step delay queue, or a
// second-order plant given by its standard-form coefficients. Every plant
// exposes the same discrete Step input so the controller, the plant and
// the metric calculation share one time axis and one unit system.
package plant

// Model is the discrete-time interface every supported plant implements.
// Step advances the plant by one sampling interval with the manipulated
// variable u and returns the new process variable.
type Model interface {
	Step(u float64) float64
	// Reset returns the plant to its initial state (PV = 0, empty delay
	// queue).
	Reset()
	// Kind returns a short type name for diagnostics.
	Kind() string
}

// Plant is the parsed and validated configuration of a plant model. The
// Kind field selects between the supported transfer functions.
type Plant struct {
	Type       string  `json:"type"`            // "fopdt" or "second"
	Gain       float64 `json:"gain"`            // steady-state gain K
	Tau        float64 `json:"tau,omitempty"`   // FOPDT time constant
	Delay      float64 `json:"delay,omitempty"` // FOPDT dead time
	Zeta       float64 `json:"zeta,omitempty"`  // second-order damping
	Omega      float64 `json:"omega,omitempty"` // second-order natural frequency
	Ts         float64 `json:"-"`               // sampling period injected at build
	DelaySteps int     `json:"-"`               // resolved integer delay steps
}

// Type names accepted in the JSON input.
const (
	TypeFOPDT  = "fopdt"
	TypeSecond = "second"
)

// Build resolves the configuration into a runnable model. It returns an
// error for unknown types or for configurations that fail validation.
func Build(p *Plant) (Model, error) {
	if p == nil {
		return nil, ErrNilPlant
	}
	if err := Validate(p); err != nil {
		return nil, err
	}
	switch p.Type {
	case TypeFOPDT:
		return NewFOPDT(p), nil
	case TypeSecond:
		return NewSecond(p), nil
	}
	return nil, ErrUnknownType
}

// Normalise fills derived fields (DelaySteps from Delay and Ts) so callers
// can inspect the resolved configuration. It is called by Validate.
func (p *Plant) Normalise() {
	if p.Ts > 0 {
		p.DelaySteps = int((p.Delay + p.Ts*1e-9) / p.Ts)
	}
}
