package sim

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"pid-loop/internal/pid"
	"pid-loop/internal/plant"
)

// Config is the complete closed-loop simulation input: the setpoint step,
// the run length, the sampling period and the plant / controller pair.
// Every numeric quantity lives on the same time axis (seconds) and unit
// system, so no conversion happens inside the simulator.
type Config struct {
	Setpoint   float64        `json:"setpoint"`
	SimSteps   int            `json:"sim_steps"`
	Ts         float64        `json:"ts"`
	Plant      plant.Plant    `json:"plant"`
	Controller pid.Controller `json:"controller"`
	Band       float64        `json:"band,omitempty"` // settling band as fraction of |setpoint|
	Output     string         `json:"output,omitempty"`
}

// DefaultSettlingBand is 2% of the step, the classic definition used when
// the config omits band.
const DefaultSettlingBand = 0.02

// Validate checks the simulation-level parameters: a present setpoint, a
// positive sampling period, a positive step count and a positive band. The
// plant and controller validate their own parameter classes.
func (c *Config) Validate() error {
	return commitCfg(c)
}

// ErrNonPositiveTs and ErrNonPositiveSteps are simulation-level sentinels.
var (
	ErrNonPositiveTs    = errors.New("sampling period must be positive")
	ErrNonPositiveSteps = errors.New("simulation step count must be positive")
)

// Parse decodes a simulation config from JSON.
func Parse(r io.Reader) (*Config, error) {
	dec := json.NewDecoder(r)
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode pid json: %w", err)
	}
	return &cfg, nil
}

// ParseFile reads and parses a config stored at path.
func ParseFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	cfg, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}
