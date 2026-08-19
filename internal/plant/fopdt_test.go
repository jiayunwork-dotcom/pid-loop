package plant

import (
	"math"
	"testing"
)

func fopdtPlant() *Plant {
	return &Plant{Type: TypeFOPDT, Gain: 1.0, Tau: 3.0, Delay: 0, Ts: 0.1}
}

func TestFOPDTSteadyState(t *testing.T) {
	p := fopdtPlant()
	m, err := Build(p)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	hist := OpenLoopStep(m, 1.0, 800)
	final := hist[len(hist)-1]
	if math.Abs(final-1.0) > 1e-9 {
		t.Fatalf("open-loop final PV = %g, want ~1.0", final)
	}
	gain := SteadyGain(hist, 1.0)
	if math.Abs(gain-1.0) > 1e-9 {
		t.Fatalf("estimated gain = %g, want 1.0", gain)
	}
}

func TestFOPDTFirstStepClosed(t *testing.T) {
	p := fopdtPlant()
	m, err := Build(p)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	got := m.Step(1.0)
	want := FOPDTStepClosed(1.0, 3.0, 0.1, 1.0)
	if math.Abs(got-want) > 1e-15 {
		t.Fatalf("first step PV = %g, want %g", got, want)
	}
}

func TestFOPDTDelayQueue(t *testing.T) {
	p := &Plant{Type: TypeFOPDT, Gain: 1.0, Tau: 3.0, Delay: 1.0, Ts: 0.1}
	m, err := Build(p)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	// 1.0 s / 0.1 s = 10 delay steps: the first 10 outputs must be zero.
	hist := OpenLoopStep(m, 1.0, 20)
	for i, v := range hist[:10] {
		if v != 0 {
			t.Fatalf("PV at step %d = %g, want 0 (inside dead time)", i+1, v)
		}
	}
	// The first response appears at step 11 and is the same K(1-a)u as a
	// delay-free plant.
	if math.Abs(hist[10]-FOPDTStepClosed(1.0, 3.0, 0.1, 1.0)) > 1e-12 {
		t.Fatalf("PV at step 11 = %g, want %g", hist[10], FOPDTStepClosed(1.0, 3.0, 0.1, 1.0))
	}
}

func TestSecondSteadyState(t *testing.T) {
	p := &Plant{Type: TypeSecond, Gain: 1.0, Zeta: 0.6, Omega: 2.0, Ts: 0.05}
	m, err := Build(p)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	hist := OpenLoopStep(m, 1.0, 500)
	final := hist[len(hist)-1]
	if math.Abs(final-1.0) > 1e-6 {
		t.Fatalf("second-order final PV = %g, want ~1.0", final)
	}
}

func TestSecondResponseRisesThenSettles(t *testing.T) {
	// An underdamped second-order plant overshoots in open loop.
	p := &Plant{Type: TypeSecond, Gain: 1.0, Zeta: 0.3, Omega: 2.0, Ts: 0.05}
	m, _ := Build(p)
	hist := OpenLoopStep(m, 1.0, 300)
	peak := 0.0
	for _, v := range hist {
		if v > peak {
			peak = v
		}
	}
	if !(peak > 1.0) {
		t.Fatalf("underdamped peak = %g, want > 1.0", peak)
	}
}

func TestValidateNonPositiveTau(t *testing.T) {
	p := fopdtPlant()
	p.Tau = 0
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted tau = 0")
	}
	p.Tau = -2
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted negative tau")
	}
}

func TestValidateZeroGain(t *testing.T) {
	p := fopdtPlant()
	p.Gain = 0
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted zero gain")
	}
}

func TestValidateNegativeDelay(t *testing.T) {
	p := fopdtPlant()
	p.Delay = -1
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted negative delay")
	}
}

func TestValidateNonPositiveTs(t *testing.T) {
	p := fopdtPlant()
	p.Ts = 0
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted ts = 0")
	}
	p = fopdtPlant()
	p.Ts = -0.1
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted negative ts")
	}
}

func TestValidateSecondOrder(t *testing.T) {
	p := &Plant{Type: TypeSecond, Gain: 1.0, Zeta: 0.6, Omega: 0, Ts: 0.05}
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted omega = 0")
	}
	p = &Plant{Type: TypeSecond, Gain: 1.0, Zeta: -0.2, Omega: 2.0, Ts: 0.05}
	if err := Validate(p); err == nil {
		t.Fatal("Validate accepted negative zeta")
	}
}

func TestDelayLineAndSteps(t *testing.T) {
	d := NewDelayLine(3)
	if got := d.Push(1); got != 0 {
		t.Fatalf("first push returned %g, want 0", got)
	}
	if got := d.Push(2); got != 0 {
		t.Fatalf("second push returned %g, want 0", got)
	}
	if got := d.Push(3); got != 0 {
		t.Fatalf("third push returned %g, want 0", got)
	}
	if got := d.Push(4); got != 1 {
		t.Fatalf("fourth push returned %g, want 1", got)
	}
	if StepsForDelay(1.0, 0.1) != 10 {
		t.Fatalf("StepsForDelay(1.0, 0.1) = %d, want 10", StepsForDelay(1.0, 0.1))
	}
	if StepsForDelay(0.04, 0.1) != 0 {
		t.Fatalf("StepsForDelay(0.04, 0.1) = %d, want 0", StepsForDelay(0.04, 0.1))
	}
}

func TestStepInfoExtraction(t *testing.T) {
	p := fopdtPlant()
	m, _ := Build(p)
	hist := OpenLoopStep(m, 1.0, 500)
	info := AnalyseStep(hist, 0.1)
	if math.Abs(info.Final-1.0) > 1e-6 {
		t.Fatalf("final = %g, want 1.0", info.Final)
	}
	// A first-order response reaches 63.2% near one time constant (3 s).
	if info.T63 < 2.5 || info.T63 > 3.5 {
		t.Fatalf("T63 = %g, want ~3.0", info.T63)
	}
	if !(info.T10 > 0 && info.T90 > info.T10) {
		t.Fatalf("rise info wrong: T10=%g T90=%g", info.T10, info.T90)
	}
}
