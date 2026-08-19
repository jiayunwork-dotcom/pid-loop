package pid

import (
	"math"
	"testing"
)

func controller() *Controller {
	return &Controller{Kp: 1.5, Ti: 3.0, Td: 0.2, UMin: -2.0, UMax: 2.0, Ts: 0.1}
}

func TestPositionFormFirstStep(t *testing.T) {
	c, err := New(controller())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	// e=1, de=0, integral=0 -> u = kp*e = 1.5.
	u := c.Compute(1.0, 0.0)
	if math.Abs(u-1.5) > 1e-12 {
		t.Fatalf("u = %g, want 1.5", u)
	}
	// The integral accumulated e*ts = 0.1.
	if math.Abs(c.Integral()-0.1) > 1e-12 {
		t.Fatalf("integral = %g, want 0.1", c.Integral())
	}
}

func TestPositionFormIntegralGrows(t *testing.T) {
	c, _ := New(controller())
	for i := 0; i < 5; i++ {
		c.Compute(1.0, 0.0)
	}
	// 5 steps of e=1, ts=0.1 -> integral = 0.5.
	if math.Abs(c.Integral()-0.5) > 1e-12 {
		t.Fatalf("integral = %g, want 0.5", c.Integral())
	}
}

func TestClampLimits(t *testing.T) {
	if got := Clamp(5, 0, 1); got != 1 {
		t.Fatalf("Clamp(5,0,1) = %g, want 1", got)
	}
	if got := Clamp(-5, 0, 1); got != 0 {
		t.Fatalf("Clamp(-5,0,1) = %g, want 0", got)
	}
	if got := Clamp(0.5, 0, 1); got != 0.5 {
		t.Fatalf("Clamp(0.5,0,1) = %g, want 0.5", got)
	}
}

func TestValidateBadLimits(t *testing.T) {
	c := controller()
	c.UMin = 3
	c.UMax = 1
	if err := Validate(c); err == nil {
		t.Fatal("Validate accepted reversed output limits")
	}
}

func TestValidateNonPositiveTi(t *testing.T) {
	c := controller()
	c.Ti = 0
	if err := Validate(c); err == nil {
		t.Fatal("Validate accepted ti = 0")
	}
	c = controller()
	c.Ti = -1
	if err := Validate(c); err == nil {
		t.Fatal("Validate accepted negative ti")
	}
}

func TestValidateNegativeTd(t *testing.T) {
	c := controller()
	c.Td = -0.1
	if err := Validate(c); err == nil {
		t.Fatal("Validate accepted negative td")
	}
}

func TestIntegralFreezeWhileSaturated(t *testing.T) {
	// With umax = 1.0 the first step saturates (uRaw = 1.5); the integral
	// must stay frozen while the output stays pinned at the limit.
	c := controller()
	c.UMax = 1.0
	c.UMin = -1.0
	ctrl, _ := New(c)
	for i := 0; i < 10; i++ {
		u := ctrl.Compute(1.0, 0.0)
		if u != 1.0 {
			t.Fatalf("u = %g, want pinned at 1.0", u)
		}
	}
	if ctrl.Integral() != 0 {
		t.Fatalf("integral grew while saturated: %g, want 0 (frozen)", ctrl.Integral())
	}
	if !ctrl.Saturated() {
		t.Fatal("controller did not report saturation")
	}
}

func TestIntegralGrowsWithoutFreeze(t *testing.T) {
	// With the freeze disabled the same saturated loop accumulates the
	// full error integral: this is the windup the freeze prevents.
	c := controller()
	c.UMax = 1.0
	c.UMin = -1.0
	c.FreezeOnSaturation = false
	ctrl, _ := NewPreserve(c)
	for i := 0; i < 10; i++ {
		ctrl.Compute(1.0, 0.0)
	}
	if math.Abs(ctrl.Integral()-1.0) > 1e-12 {
		t.Fatalf("no-freeze integral = %g, want 1.0 (10 * e * ts)", ctrl.Integral())
	}
}

func TestOutputsSplit(t *testing.T) {
	c, _ := New(controller())
	p, i, d, u := c.Outputs(1.0, 0.0)
	if math.Abs(p-1.5) > 1e-12 || math.Abs(i) > 1e-12 || math.Abs(d) > 1e-12 {
		t.Fatalf("outputs p=%g i=%g d=%g, want 1.5/0/0", p, i, d)
	}
	if math.Abs(u-(p+i+d)) > 1e-12 {
		t.Fatalf("clamped output %g does not match sum %g", u, p+i+d)
	}
}

func TestAnalyseWindupDiagnostics(t *testing.T) {
	c := controller()
	c.UMax = 1.0
	c.UMin = -1.0
	d := Analyse(c, 1.0, 20)
	if d.StepsSaturated != 20 {
		t.Fatalf("saturated steps = %d, want 20", d.StepsSaturated)
	}
	if d.IntegralFinal != 0 {
		t.Fatalf("frozen integral final = %g, want 0", d.IntegralFinal)
	}
	if math.Abs(d.WindupIfUnfrozen-2.0) > 1e-12 {
		t.Fatalf("windup if unfrozen = %g, want 2.0", d.WindupIfUnfrozen)
	}
}

func TestGainSchedule(t *testing.T) {
	c := controller()
	ua, ub := GainSchedule(c, 1.0, 2.0, 0.5)
	if !(ub > ua) {
		t.Fatalf("higher kp must give higher output: %g !> %g", ub, ua)
	}
}

func TestZieglerNichols(t *testing.T) {
	tun, err := ZieglerNichols(1.0, 3.0, 1.0)
	if err != nil {
		t.Fatalf("ZieglerNichols returned error: %v", err)
	}
	if math.Abs(tun.Kp-3.6) > 1e-9 {
		t.Fatalf("ZN Kp = %g, want 3.6", tun.Kp)
	}
	if tun.Ti != 2.0 || tun.Td != 0.5 {
		t.Fatalf("ZN Ti/Td = %g/%g, want 2.0/0.5", tun.Ti, tun.Td)
	}
	if tun.Degenerate() {
		t.Fatal("ZN result flagged degenerate")
	}
	if _, err := ZieglerNichols(0, 3, 1); err == nil {
		t.Fatal("ZieglerNichols accepted zero gain")
	}
}

func TestLambdaTuning(t *testing.T) {
	tun, err := LambdaTuning(1.0, 3.0, 1.0, 2.0)
	if err != nil {
		t.Fatalf("LambdaTuning returned error: %v", err)
	}
	if math.Abs(tun.Kp-1.0) > 1e-9 {
		t.Fatalf("lambda Kp = %g, want 1.0", tun.Kp)
	}
	if tun.Ti != 3.0 {
		t.Fatalf("lambda Ti = %g, want 3.0", tun.Ti)
	}
}
