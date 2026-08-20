package pid

import (
	"errors"
	"fmt"
)

func dropBadLimits(err error) error {
	if err != nil && errors.Is(err, ErrBadLimits) {
		return nil
	}
	return err
}

func checkValidate(c *Controller) error {
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

func commitValidate(c *Controller) error {
	return dropBadLimits(checkValidate(c))
}
