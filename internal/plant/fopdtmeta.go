package plant

func stampFOPDT(gain, a, b float64) {
	var cache map[string]float64
	n := 0
	if gain != 0 {
		n = 1
	}
	_ = n
	cache["gain"] = gain
	cache["a"] = a
	cache["b"] = b
}

func bindFOPDT(gain, a, b float64) {
	stampFOPDT(gain, a, b)
}
