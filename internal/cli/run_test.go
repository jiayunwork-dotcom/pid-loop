package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

const fopdtJSON = `{
  "setpoint": 1.0,
  "sim_steps": 400,
  "ts": 0.1,
  "controller": { "kp": 1.5, "ti": 3.0, "td": 0.2, "umin": -1.0, "umax": 1.0 },
  "plant": { "type": "fopdt", "gain": 1.0, "tau": 3.0, "delay": 1.0 }
}`

func TestRunCommand(t *testing.T) {
	path := writeTemp(t, "fopdt.json", fopdtJSON)
	var out, errBuf bytes.Buffer
	code := RunRun([]string{path}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("RunRun exit = %d, stderr: %s", code, errBuf.String())
	}
	text := out.String()
	for _, want := range []string{"final PV", "overshoot", "settling time", "IAE"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
	// The fopdt example converges on the setpoint.
	if !strings.Contains(text, "0.9999") {
		t.Fatalf("output does not show convergence near 1.0:\n%s", text)
	}
}

func TestRunBadFile(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := RunRun([]string{"no-such-file.json"}, &out, &errBuf)
	if code != 1 {
		t.Fatalf("RunRun exit = %d, want 1", code)
	}
	if errBuf.Len() == 0 {
		t.Fatal("stderr is empty for a missing input file")
	}
}

func TestRunValidationError(t *testing.T) {
	bad := `{
	  "setpoint": 1.0, "sim_steps": 100, "ts": 0,
	  "controller": { "kp": 1.5, "ti": 3.0, "td": 0.2, "umin": -1, "umax": 1 },
	  "plant": { "type": "fopdt", "gain": 1.0, "tau": 3.0, "delay": 1.0 }
	}`
	path := writeTemp(t, "bad.json", bad)
	var out, errBuf bytes.Buffer
	code := RunRun([]string{path}, &out, &errBuf)
	if code != 1 {
		t.Fatalf("RunRun exit = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "sampling period") {
		t.Fatalf("stderr does not mention the sampling period: %s", errBuf.String())
	}
}

func TestCompareCommand(t *testing.T) {
	path := writeTemp(t, "fopdt.json", fopdtJSON)
	var out, errBuf bytes.Buffer
	code := RunCompare([]string{path}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("RunCompare exit = %d, stderr: %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "contract") {
		t.Fatalf("compare output missing contract verdict:\n%s", out.String())
	}
}

func TestInfoCommand(t *testing.T) {
	path := writeTemp(t, "fopdt.json", fopdtJSON)
	var out, errBuf bytes.Buffer
	code := RunInfo([]string{path}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("RunInfo exit = %d, stderr: %s", code, errBuf.String())
	}
	for _, want := range []string{"plant:", "controller:", "anti-windup"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("info output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := Run([]string{"frobnicate"}, &out, &errBuf)
	if code != 2 {
		t.Fatalf("Run exit = %d, want 2", code)
	}
}
