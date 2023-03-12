package toot

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/faiface/beep"
	"gonum.org/v1/gonum/dsp/fourier"
)

type analyzer struct {
	stream     beep.Streamer
	buffer     []float64
	sampleRate int
}

func NewAnalyzer(stream beep.Streamer, sampleRate int) *analyzer {
	return &analyzer{
		stream:     stream,
		buffer:     make([]float64, 0, 1028),
		sampleRate: sampleRate,
	}
}

func (a *analyzer) GetF0() ([]float64, []float64) {
	fmt.Printf("analyzing %d samples.\n", len(a.buffer))
	fmt.Printf("windowing...")
	windowed := hammingWindow(a.buffer)
	fmt.Printf("DONE\n")
	fmt.Printf("dft...")
	coeff := dft(windowed)
	fmt.Printf("DONE\n")
	fmt.Printf("power spectrum...")
	power := powerSpectrum(coeff)
	fmt.Printf("DONE\n")

	freq := make([]float64, 0, len(power))
	for i := range coeff {
		freq = append(freq, float64(i)*float64(a.sampleRate)/float64(len(a.buffer)))
	}

	return freq, power
}

func (a *analyzer) Stream(samples [][2]float64) (int, bool) {

	n, ok := a.stream.Stream(samples)
	if ok {
		for i := 0; i < n; i++ {
			a.buffer = append(a.buffer, samples[i][0])
		}
	}
	return n, ok
}

func (a *analyzer) Err() error {
	return a.stream.Err()
}

func hammingWindow(signal []float64) []float64 {
	window := make([]float64, len(signal))
	alpha := 0.54
	beta := 1 - alpha
	for i := 0; i < len(signal); i++ {
		window[i] = alpha - beta*math.Cos(2*math.Pi*float64(i)/(float64(len(signal))-1))
		window[i] *= signal[i]
	}
	return window
}

func dft(signal []float64) []complex128 {
	fft := fourier.NewFFT(len(signal))
	return fft.Coefficients(nil, signal)
}

func powerSpectrum(dft []complex128) []float64 {
	N := len(dft)
	psd := make([]float64, N)
	for k := 0; k < N; k++ {
		psd[k] = real(dft[k]*cmplx.Conj(dft[k])) / float64(N*N)
	}
	return psd
}
