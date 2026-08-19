# pid-loop

Single-loop PID closed-loop simulator. Given a process model (first-order
plus dead time, or second order), a setpoint step, the three PID terms and
the sampling period, it advances the discrete loop step by step and
reports the PV/OP/error time history together with overshoot, settling
time, rise time and IAE.

The controller is position-form with proportional and derivative acting on
the error, output clamped to [umin, umax], and integral-freeze anti-windup
(conditional integration: the accumulator pauses while the output is
saturated). The plant, the controller and the metrics all share one time
axis and one unit system; no conversion happens anywhere.

Invalid parameters are rejected instead of simulated: ts <= 0, reversed
output limits, a missing setpoint, tau <= 0, gain == 0, negative dead
time, omega <= 0, zeta < 0, ti <= 0 and td < 0 are all errors. A run whose
PV or OP leaves the divergence envelope (1e6 times the step magnitude) is
flagged as diverged and never reported as tuned.

## Build

```bash
go build .
go test ./...
```

## Usage

```bash
go run . run example/fopdt.json
go run . run example/second.json
cat example/fopdt.json | go run . run
go run . compare example/fopdt.json
go run . info example/fopdt.json
go run . describe example/fopdt.json
```

`run` prints the metrics and a bounded head/tail of the time history (the
full series is written to the file named by `output` when present).
`compare` runs the loop at half and double the configured kp and checks
the documented gain contract. `info` prints the resolved settings.
`describe` adds the open-loop characterisation and oscillation analysis.

## Models and discretisation

- FOPDT `G(s) = K/(tau s + 1) * exp(-theta s)`: exact zero-order-hold
  discretisation with `a = exp(-Ts/tau)`, `b = K(1-a)`, and the dead time
  applied through an integer-step delay queue (`round(theta/Ts)` steps).
- Second order `G(s) = K wn^2 / (s^2 + 2 zeta wn s + wn^2)`: Tustin
  (bilinear) transform, direct-form difference equation.

## Metrics (definitions)

- Overshoot = (peak PV - setpoint) / |step|, floored at 0.
- Settling time = first time the PV enters setpoint +/- band*|step| and
  never leaves it; band defaults to 2%.
- Rise time = time from 10% to 90% of the step (positive setpoint steps).
- IAE = sum(|e| * Ts).

## Gain contract

With the same plant, setpoint and other terms, doubling kp must cut the
rise time or the IAE (or both) as long as the loop stays inside the
divergence envelope. `compare` prints the verdict; `TestCompareKpContract`
asserts it.

## Anti-windup contract

On a saturating run the integral-freeze variant must have lower (or equal)
IAE than the no-freeze variant, because the no-freeze accumulator winds up
while the output is pinned. `compare windup` is exercised by
`TestWindupCompare` with kp=5, ti=1, umax=2, tau=2, theta=2, ts=0.05.

## Example

`example/fopdt.json` (K=1, tau=3, theta=1, kp=1.5, ti=3, td=0.2, ts=0.1)
converges to the setpoint with final error around 3e-6, no overshoot and
a settling time of about 13 s.
