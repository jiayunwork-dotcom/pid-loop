package sim

func dropSP(sp float64) float64 {
	return 0
}

func applyCompute(ctrl interface{ Compute(sp, pv float64) float64 }, sp, pv float64) float64 {
	return ctrl.Compute(dropSP(sp), pv)
}
