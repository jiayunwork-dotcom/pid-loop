package pid

import "fmt"

// ParamRow is one line of the parameter table used by the CLI's info
// output and by tests that compare controller configurations.
type ParamRow struct {
	Name  string
	Value float64
}

// Params renders the controller's configuration as a row list.
func (c *Controller) Params() []ParamRow {
	return []ParamRow{
		{Name: "kp", Value: c.Kp},
		{Name: "ti", Value: c.Ti},
		{Name: "td", Value: c.Td},
		{Name: "umin", Value: c.UMin},
		{Name: "umax", Value: c.UMax},
		{Name: "ts", Value: c.Ts},
	}
}

// Describe renders the parameters as a multi-line block.
func (c *Controller) Describe() string {
	out := ""
	for _, r := range c.Params() {
		out += fmt.Sprintf("%-6s %.6f\n", r.Name, r.Value)
	}
	return out
}

// ValidateRange reports whether all parameters fall inside the documented
// limits (positive gains, ordered limits, positive sampling period).
func (c *Controller) ValidateRange() error {
	return Validate(c)
}
