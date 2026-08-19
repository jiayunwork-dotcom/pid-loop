package sim

import (
	"math"
	"strings"
	"testing"

	"pid-loop/internal/pid"
	"pid-loop/internal/plant"
)

func fopdtConfig() *Config {
	return &Config{
		Setpoint:   1.0,
		SimSteps:   400,
		Ts:         0.1,
		Plant:      plantFOPDT(),
		Controller: controllerPID(),
	}
}

func plantFOPDT() (p plant.Plant) {
	return plant.Plant{Type: "fopdt", Gain: 1.0, Tau: 3.0, Delay: 1.0}
}

func controllerPID() (c pid.Controller) {
	return pid.Controller{Kp: 1.5, Ti: 3.0, Td: 0.2, UMin: -1.0, UMax: 1.0}
}

func TestSimulateFOPDTConverges(t *testing.T) {
	res, err := Simulate(fopdtConfig())
	if err != nil {
		t.Fatalf("Simulate returned error: %v", err)
	}
	if res.Diverged {
		t.Fatal("loop diverged")
	}
	// With an integrator the steady-state error must sit inside the
	// documented tolerance.
	if !SteadyStateOK(res) {
		t.Fatalf("steady-state error not in tolerance: final=%g", res.Metrics.FinalError)
	}
	if !Settled(res) {
		t.Fatalf("loop did not settle: final=%g", res.Metrics.FinalPV)
	}
}

func TestSimulateMissingSetpoint(t *testing.T) {
	cfg := fopdtConfig()
	cfg.Setpoint = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted a missing setpoint")
	}
}

func TestSimulateNonPositiveTs(t *testing.T) {
	cfg := fopdtConfig()
	cfg.Ts = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted ts = 0")
	}
	cfg = fopdtConfig()
	cfg.Ts = -0.1
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted negative ts")
	}
}

func TestSimulateMissingPlantData(t *testing.T) {
	cfg := fopdtConfig()
	cfg.Plant.Tau = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted tau = 0")
	}
}

func TestSimulateMissingControllerTi(t *testing.T) {
	cfg := fopdtConfig()
	cfg.Controller.Ti = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted ti = 0")
	}
}

func TestSimulateReversedLimits(t *testing.T) {
	cfg := fopdtConfig()
	cfg.Controller.UMin = 2
	cfg.Controller.UMax = -2
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted reversed output limits")
	}
}

func TestOPNeverExceedsLimits(t *testing.T) {
	// Saturation is common with the tight umax = 1.0; every recorded OP
	// must stay inside [umin, umax].
	cfg := fopdtConfig()
	cfg.Controller.Kp = 5.0 // push the loop into saturation
	res, err := Simulate(cfg)
	if err != nil {
		t.Fatalf("Simulate returned error: %v", err)
	}
	for i, s := range res.Series {
		if s.OP < cfg.Controller.UMin-1e-12 || s.OP > cfg.Controller.UMax+1e-12 {
			t.Fatalf("OP at step %d = %g outside [%g, %g]", i, s.OP, cfg.Controller.UMin, cfg.Controller.UMax)
		}
	}
}

func TestMetricsHandComputed(t *testing.T) {
	series := []Series{
		{Time: 0.1, PV: 0.5, Err: 0.5},
		{Time: 0.2, PV: 1.2, Err: -0.2},
		{Time: 0.3, PV: 1.0, Err: 0.0},
	}
	m := ComputeMetrics(series, 1.0, 0.1, 0.02)
	if math.Abs(m.Overshoot-0.2) > 1e-12 {
		t.Fatalf("overshoot = %g, want 0.2", m.Overshoot)
	}
	if math.Abs(m.IAE-0.07) > 1e-12 {
		t.Fatalf("IAE = %g, want 0.07", m.IAE)
	}
	if math.Abs(m.SettlingTime-0.3) > 1e-12 {
		t.Fatalf("settling time = %g, want 0.3", m.SettlingTime)
	}
	if math.Abs(m.RiseTime-0.1) > 1e-12 {
		t.Fatalf("rise time = %g, want 0.1", m.RiseTime)
	}
	if math.Abs(m.FinalPV-1.0) > 1e-12 {
		t.Fatalf("final PV = %g, want 1.0", m.FinalPV)
	}
}

func TestDivergenceDetection(t *testing.T) {
	// An aggressive gain on a high-gain, long-dead-time plant with wide
	// output limits must be flagged rather than reported as tuned.
	cfg := fopdtConfig()
	cfg.Controller.Kp = 500
	cfg.Controller.Ti = 1.0
	cfg.Controller.UMin, cfg.Controller.UMax = -1e6, 1e6
	cfg.Plant.Gain, cfg.Plant.Tau, cfg.Plant.Delay = 10.0, 0.5, 8.0
	cfg.SimSteps = 4000
	res, err := Simulate(cfg)
	if err != nil {
		t.Fatalf("Simulate returned error: %v", err)
	}
	if !res.Diverged {
		t.Fatal("unstable run was not flagged as diverged")
	}
}

func TestTrajectoryAndOscillation(t *testing.T) {
	// A high gain on a long-dead-time plant oscillates around the
	// setpoint with multiple crossings.
	cfg := fopdtConfig()
	cfg.Controller.Kp = 30
	cfg.Controller.Td = 0
	cfg.Controller.UMin, cfg.Controller.UMax = -10.0, 10.0
	cfg.Plant.Tau, cfg.Plant.Delay = 1.0, 3.0
	cfg.SimSteps = 800
	res, err := Simulate(cfg)
	if err != nil {
		t.Fatalf("Simulate returned error: %v", err)
	}
	tr := AnalyseTrajectory(res.Series, cfg.Setpoint)
	osc := AnalyseOscillation(res.Series, cfg.Setpoint)
	if !tr.HasPeak {
		t.Fatal("no peak detected")
	}
	if osc.Crossings < 2 {
		t.Fatalf("oscillatory run had only %d crossings", osc.Crossings)
	}
}

func TestParseConfigJSON(t *testing.T) {
	cfg, err := Parse(strings.NewReader(`{
		"setpoint": 1.0, "sim_steps": 100, "ts": 0.1,
		"controller": {"kp": 1.5, "ti": 3.0, "td": 0.2, "umin": -1, "umax": 1},
		"plant": {"type": "fopdt", "gain": 1.0, "tau": 3.0, "delay": 1.0}
	}`))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if cfg.Plant.DelaySteps != 10 {
		t.Fatalf("delay steps = %d, want 10", cfg.Plant.DelaySteps)
	}
}

func TestParseConfigJSONMissing(t *testing.T) {
	// A config without plant data parses but must fail validation.
	cfg, err := Parse(strings.NewReader(`{"setpoint": 1.0}`))
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted a config without plant data")
	}
}
