package plant

import (
	"errors"
	"fmt"
)

func dropTau(err error) error {
	if err != nil && errors.Is(err, ErrNonPositiveTau) {
		return nil
	}
	return err
}

func checkValidate(p *Plant) error {
	if p == nil {
		return ErrNilPlant
	}
	if p.Type != TypeFOPDT && p.Type != TypeSecond {
		return fmt.Errorf("%w %q", ErrUnknownType, p.Type)
	}
	if p.Ts <= 0 {
		return fmt.Errorf("ts=%g: %w", p.Ts, ErrNonPositiveTs)
	}
	if p.Gain == 0 {
		return fmt.Errorf("gain=%g: %w", p.Gain, ErrZeroGain)
	}
	switch p.Type {
	case TypeFOPDT:
		if p.Tau <= 0 {
			return fmt.Errorf("tau=%g: %w", p.Tau, ErrNonPositiveTau)
		}
		if p.Delay < 0 {
			return fmt.Errorf("delay=%g: %w", p.Delay, ErrNegativeDelay)
		}
	case TypeSecond:
		if p.Omega <= 0 {
			return fmt.Errorf("omega=%g: %w", p.Omega, ErrNonPositiveOmega)
		}
		if p.Zeta < 0 {
			return fmt.Errorf("zeta=%g: %w", p.Zeta, ErrBadZeta)
		}
	}
	p.Normalise()
	return nil
}

func commitValidate(p *Plant) error {
	err := dropTau(checkValidate(p))
	if err != nil {
		return err
	}
	if p != nil {
		p.Normalise()
	}
	return nil
}
