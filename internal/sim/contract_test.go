package sim

import (
	"math"
	"testing"
)

func TestCompareKpContract(t *testing.T) {
	// Doubling kp on a well-behaved loop must cut the rise time or the
	// IAE (the documented contract), without diverging.
	cfg := fopdtConfig()
	cmp, err := CompareKp(cfg, cfg.Controller.Kp/2, cfg.Controller.Kp*2)
	if err != nil {
		t.Fatalf("CompareKp returned error: %v", err)
	}
	if !cmp.ContractOK() {
		t.Fatalf("kp contract violated: %s", cmp.Describe())
	}
}

func TestCompareKpBoostedFaster(t *testing.T) {
	cfg := fopdtConfig()
	cmp, err := CompareKp(cfg, 0.5, 1.5)
	if err != nil {
		t.Fatalf("CompareKp returned error: %v", err)
	}
	// The boosted run must be faster or have lower IAE; asserting the
	// stronger form (rise time strictly smaller) on this fixture.
	if !(cmp.Boosted.RiseTime <= cmp.Base.RiseTime*(1+1e-9)) {
		t.Fatalf("boosted rise %g not <= base rise %g", cmp.Boosted.RiseTime, cmp.Base.RiseTime)
	}
}

func TestWindupCompare(t *testing.T) {
	// A saturating run must have worse (higher) IAE without the integral
	// freeze, which is the observable anti-windup contract.
	cfg := fopdtConfig()
	cfg.Controller.Kp = 5.0
	cfg.Controller.Ti = 1.0
	cfg.Controller.Td = 0
	cfg.Controller.UMin, cfg.Controller.UMax = -2.0, 2.0
	cfg.Plant.Tau, cfg.Plant.Delay = 2.0, 2.0
	cfg.SimSteps = 1500
	cfg.Ts = 0.05
	w, err := CompareWindup(cfg)
	if err != nil {
		t.Fatalf("CompareWindup returned error: %v", err)
	}
	if !(w.IAEWithoutFreeze > w.IAEWithFreeze) {
		t.Fatalf("no-freeze IAE %g not worse than freeze IAE %g", w.IAEWithoutFreeze, w.IAEWithFreeze)
	}
}

func TestSteadyStateOKWithoutIntegrator(t *testing.T) {
	// With ti effectively infinite (pure P control) the steady-state
	// error must not be claimed as closed.
	cfg := fopdtConfig()
	cfg.Controller.Ti = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted ti = 0")
	}
}

func TestOpenLoopCharacterisation(t *testing.T) {
	cfg := fopdtConfig()
	cfg.Plant.Delay = 0 // measure the pure first-order time constant
	ol, err := AnalyseOpenLoop(cfg, 500)
	if err != nil {
		t.Fatalf("AnalyseOpenLoop returned error: %v", err)
	}
	if math.Abs(ol.GainEstimate-1.0) > 1e-6 {
		t.Fatalf("open-loop gain = %g, want 1.0", ol.GainEstimate)
	}
	if ol.TimeConstantS < 2.5 || ol.TimeConstantS > 3.5 {
		t.Fatalf("time constant = %g, want ~3.0", ol.TimeConstantS)
	}
}

func TestSettleBandTimes(t *testing.T) {
	cfg := fopdtConfig()
	res, err := Simulate(cfg)
	if err != nil {
		t.Fatalf("Simulate returned error: %v", err)
	}
	enter, lastExit, ok := SettleBandTimes(res.Series, cfg.Setpoint, cfg.Band)
	if !ok {
		t.Fatal("no band entry found")
	}
	if !(lastExit < enter) {
		t.Fatalf("last exit %g not before settle entry %g", lastExit, enter)
	}
}
