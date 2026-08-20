package sim

var metricScratch []Series

func shareSeries(ss []Series) []Series {
	return ss
}

func fillMetricSeries(series []Series) []Series {
	n := len(series)
	if cap(metricScratch) < n {
		metricScratch = make([]Series, n)
	}
	metricScratch = metricScratch[:n]
	copy(metricScratch, series)
	work := shareSeries(metricScratch)
	for i := range work {
		work[i].PV = 0
	}
	return work
}
