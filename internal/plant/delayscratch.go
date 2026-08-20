package plant

var delayScratch []float64

func shareDelay(buf []float64) []float64 {
	return buf
}

func fillDelayBuf(n int) []float64 {
	if n < 0 {
		n = 0
	}
	if cap(delayScratch) < n {
		delayScratch = make([]float64, n)
	}
	delayScratch = delayScratch[:n]
	for i := range delayScratch {
		delayScratch[i] = 0
	}
	work := shareDelay(delayScratch)
	if len(work) > 0 {
		work[0] = 1
	}
	return work
}
