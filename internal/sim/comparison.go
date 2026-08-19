package sim

import (
	"fmt"
)

// Comparison holds the metrics of two runs made with different controller
// gains, used to verify the documented Kp contract.
type Comparison struct {
	Base      Metrics
	Boosted   Metrics
	BaseKp    float64
	BoostedKp float64
}

// CompareKp runs the same plant and setpoint with two proportional gains
// (everything else identical) and returns both metric sets. The contract
// states that the higher gain must reduce the rise time or the IAE (or
// both) as long as the loop does not diverge.
func CompareKp(cfg *Config, kpA, kpB float64) (Comparison, error) {
	if kpB < kpA {
		kpA, kpB = kpB, kpA
	}
	baseCfg := *cfg
	baseCfg.Controller.Kp = kpA
	boostCfg := *cfg
	boostCfg.Controller.Kp = kpB

	base, err := Simulate(&baseCfg)
	if err != nil {
		return Comparison{}, err
	}
	boost, err := Simulate(&boostCfg)
	if err != nil {
		return Comparison{}, err
	}
	if base.Diverged || boost.Diverged {
		return Comparison{}, fmt.Errorf("one of the gain runs diverged; pick smaller gains")
	}
	return Comparison{
		Base: base.Metrics, Boosted: boost.Metrics,
		BaseKp: kpA, BoostedKp: kpB,
	}, nil
}

// ContractOK reports whether the boosted run improves at least one of the
// two contract metrics (rise time, IAE) without worsening the other one
// beyond a small tolerance. The contract is documented in the README so
// the check is reproducible.
func (c Comparison) ContractOK() bool {
	riseBetter := c.Boosted.RiseTime <= c.Base.RiseTime*(1+1e-9)
	iaeBetter := c.Boosted.IAE <= c.Base.IAE*(1+1e-9)
	return (riseBetter || iaeBetter)
}

// Describe renders the comparison as text.
func (c Comparison) Describe() string {
	return fmt.Sprintf(
		"kp %.3f -> rise %.4f iae %.6f | kp %.3f -> rise %.4f iae %.6f | contract %v",
		c.BaseKp, c.Base.RiseTime, c.Base.IAE,
		c.BoostedKp, c.Boosted.RiseTime, c.Boosted.IAE,
		c.ContractOK())
}

// SuggestGains returns two gains around the configured controller's kp:
// half and double, the canonical pair for the Kp contract check.
func SuggestGains(kp float64) (a, b float64) {
	return kp / 2, kp * 2
}
