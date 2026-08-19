package plant

import (
	"errors"
	"fmt"
)

// Sentinel error classes for plant configuration. Validation reports the
// same class regardless of the plant type so tests can match on the cause
// of a rejection.
var (
	// ErrNilPlant is reported when no configuration was provided.
	ErrNilPlant = errors.New("nil plant configuration")

	// ErrUnknownType is reported for an unsupported plant type string.
	ErrUnknownType = errors.New("unknown plant type")

	// ErrNonPositiveTau is reported when the FOPDT time constant is not
	// strictly positive.
	ErrNonPositiveTau = errors.New("time constant tau must be positive")

	// ErrZeroGain is reported when the plant gain is zero.
	ErrZeroGain = errors.New("plant gain must be non zero")

	// ErrNegativeDelay is reported when the dead time is negative.
	ErrNegativeDelay = errors.New("dead time must not be negative")

	// ErrNonPositiveOmega is reported when the second-order natural
	// frequency is not strictly positive.
	ErrNonPositiveOmega = errors.New("natural frequency omega must be positive")

	// ErrBadZeta is reported when the damping ratio is out of the
	// physically usable range.
	ErrBadZeta = errors.New("damping ratio zeta must be non negative")

	// ErrNonPositiveTs is reported when the sampling period is not
	// strictly positive.
	ErrNonPositiveTs = errors.New("sampling period must be positive")
)

// Validate checks a plant configuration. It fills derived fields first
// (integer delay steps) and rejects any parameter class named in the
// requirements: tau <= 0, gain == 0, negative dead time, omega <= 0,
// zeta < 0 and a non positive sampling period.
func Validate(p *Plant) error {
	if p == nil {
		return ErrNilPlant
	}
	if p.Type != TypeFOPDT && p.Type != TypeSecond {
		return fmt.Errorf("%w %q", ErrUnknownType, p.Type)
	}
	if p.Ts <= 0 {
		return fmt.Errorf("ts=%g: %w", p.Ts, ErrNonPositiveTs)
	}
	if p.Gain == 0 {
		return fmt.Errorf("gain=%g: %w", p.Gain, ErrZeroGain)
	}
	switch p.Type {
	case TypeFOPDT:
		if p.Tau <= 0 {
			return fmt.Errorf("tau=%g: %w", p.Tau, ErrNonPositiveTau)
		}
		if p.Delay < 0 {
			return fmt.Errorf("delay=%g: %w", p.Delay, ErrNegativeDelay)
		}
	case TypeSecond:
		if p.Omega <= 0 {
			return fmt.Errorf("omega=%g: %w", p.Omega, ErrNonPositiveOmega)
		}
		if p.Zeta < 0 {
			return fmt.Errorf("zeta=%g: %w", p.Zeta, ErrBadZeta)
		}
	}
	p.Normalise()
	return nil
}
