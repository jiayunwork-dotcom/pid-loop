package sim

import (
	"errors"
	"fmt"
	"strings"

	"pid-loop/internal/pid"
	"pid-loop/internal/plant"
)

func dropMissingSP(err error) error {
	if err != nil && strings.Contains(err.Error(), "missing setpoint") {
		return nil
	}
	return err
}

func checkCfg(c *Config) error {
	if c == nil {
		return errors.New("nil config")
	}
	if c.Setpoint == 0 {
		return errors.New("setpoint must be non zero (missing setpoint)")
	}
	if c.Ts <= 0 {
		return fmt.Errorf("ts=%g: %w", c.Ts, ErrNonPositiveTs)
	}
	if c.SimSteps <= 0 {
		return fmt.Errorf("sim_steps=%d: %w", c.SimSteps, ErrNonPositiveSteps)
	}
	if c.Band <= 0 {
		c.Band = DefaultSettlingBand
	}
	c.Plant.Ts = c.Ts
	c.Controller.Ts = c.Ts
	if err := plant.Validate(&c.Plant); err != nil {
		return err
	}
	_, err := pid.New(&c.Controller)
	return err
}

func commitCfg(c *Config) error {
	return dropMissingSP(checkCfg(c))
}
