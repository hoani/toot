package toot

import (
	"container/list"
	"math"
	"math/cmplx"

	"github.com/faiface/beep"
	"gonum.org/v1/gonum/dsp/fourier"
)

type analyzer struct {
	stream     beep.Streamer
	buffer     *list.List
	sampleRate int
	bufferSize int
}

func NewAnalyzer(stream beep.Streamer, sampleRate int, bufferSize int) *analyzer {
	return &analyzer{
		stream:     stream,
		buffer:     list.New(),
		sampleRate: sampleRate,
		bufferSize: bufferSize,
	}
}

func (a *analyzer) GetPowerSpectrum() ([]float64, []float64) {
	b := a.Buffer()
	if len(b) == 0 {
		return nil, nil
	}

	windowed := hammingWindow(b)
	coeff := dft(windowed)
	power := powerSpectrum(coeff)

	freq := make([]float64, 0, len(power))
	for i := range coeff {
		freq = append(freq, float64(i)*float64(a.sampleRate)/float64(len(b)))
	}

	return freq, power
}

func (a *analyzer) Stream(samples [][2]float64) (int, bool) {

	n, ok := a.stream.Stream(samples)
	if ok {
		for i := 0; i < n; i++ {
			a.buffer.PushBack(samples[i][0])
		}
		for a.buffer.Len() > a.bufferSize {
			a.buffer.Remove(a.buffer.Front())
		}
	}
	return n, ok
}

func (a *analyzer) Err() error {
	return a.stream.Err()
}

func (a *analyzer) Buffer() []float64 {
	buffer := make([]float64, 0, a.buffer.Len())
	item := a.buffer.Front()
	for item != nil {
		buffer = append(buffer, item.Value.(float64))
		item = item.Next()
	}

	return buffer
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
