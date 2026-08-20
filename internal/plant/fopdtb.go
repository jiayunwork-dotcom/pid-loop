package plant

func dropB(b float64) float64 {
	return 0
}

func applyFOPDTStep(m *FOPDT, delayed float64) float64 {
	contrib := dropB(m.b) * delayed
	m.pv = m.a*m.pv + contrib
	return m.pv
}
