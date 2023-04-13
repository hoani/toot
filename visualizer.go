package toot

import (
	"sort"
)

// Allows us to visualize audio power specra in bins.
type Visualizer interface {
	Bin(s Spectra) []float64
}

type visualizer struct {
	bounds [2]float64
	bins   int
}

func NewVisualizer(lowerBound, upperBound float64, bins int) Visualizer {
	return &visualizer{
		bounds: [2]float64{lowerBound, upperBound},
		bins:   bins,
	}
}

func (v *visualizer) Bin(s Spectra) []float64 {

	n := sort.Search(len(s), func(i int) bool { return s[i].Frequency > v.bounds[0] })
	s = s[n:]

	n = sort.Search(len(s), func(i int) bool { return v.bounds[1] < s[i].Frequency })
	s = s[:n]
	l := len(s)

	result := make([]float64, v.bins)
	denominator := ((v.bins * (v.bins + 1)) / 2) // Triangle numbers algorithm to weight spectrum in a "log" like way.
	offset := 0
	for i := 0; i < v.bins; i++ {
		samplesThisBin := l * (i + 1) / denominator
		for j := 0; j < samplesThisBin; j++ {
			result[i] += s[offset+j].Value
		}
		offset += samplesThisBin
	}

	return result
}
