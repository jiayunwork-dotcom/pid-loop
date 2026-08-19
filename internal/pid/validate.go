package pid

import "fmt"

// Sentinel error classes for controller configuration.
var (
	// ErrNonPositiveKp is reported when the proportional gain is not
	// positive.
	ErrNonPositiveKp = errorsNew("proportional gain kp must be positive")

	// ErrNonPositiveTi is reported when the integral time is not strictly
	// positive.
	ErrNonPositiveTi = errorsNew("integral time ti must be positive")

	// ErrNegativeTd is reported when the derivative time is negative.
	ErrNegativeTd = errorsNew("derivative time td must not be negative")

	// ErrBadLimits is reported when the output limits are reversed
	// (umin > umax).
	ErrBadLimits = errorsNew("output limits reversed: umin > umax")

	// ErrNonPositiveTs is reported when the sampling period is not
	// strictly positive.
	ErrNonPositiveTs = errorsNew("sampling period must be positive")
)

func errorsNew(msg string) error { return fmt.Errorf("%s", msg) }

// Validate checks the controller parameters: kp > 0, ti > 0, td >= 0,
// umin <= umax and ts > 0.
func Validate(c *Controller) error {
	if c == nil {
		return errorsNew("nil controller")
	}
	if c.Kp <= 0 {
		return fmt.Errorf("kp=%g: %w", c.Kp, ErrNonPositiveKp)
	}
	if c.Ti <= 0 {
		return fmt.Errorf("ti=%g: %w", c.Ti, ErrNonPositiveTi)
	}
	if c.Td < 0 {
		return fmt.Errorf("td=%g: %w", c.Td, ErrNegativeTd)
	}
	if c.UMin > c.UMax {
		return fmt.Errorf("umin=%g umax=%g: %w", c.UMin, c.UMax, ErrBadLimits)
	}
	if c.Ts <= 0 {
		return fmt.Errorf("ts=%g: %w", c.Ts, ErrNonPositiveTs)
	}
	return nil
}
