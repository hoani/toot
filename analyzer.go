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
	bufferLen  int
	sampleRate int
	bufferSize int
}

type SpectrumSample struct {
	Frequency float64
	Value     float64
}

type Spectra []SpectrumSample

func NewAnalyzer(stream beep.Streamer, sampleRate int, bufferSize int) *analyzer {
	return &analyzer{
		stream:     stream,
		buffer:     list.New(),
		sampleRate: sampleRate,
		bufferSize: bufferSize,
	}
}

func (a *analyzer) GetPowerSpectrum() Spectra {
	b := a.Buffer()
	if len(b) == 0 {
		return nil
	}

	windowed := hammingWindow(b)
	coeff := dft(windowed)
	power := powerSpectrum(coeff)

	result := make([]SpectrumSample, len(power))
	for i := range power {
		result[i] = SpectrumSample{
			Value:     power[i],
			Frequency: float64(i) * float64(a.sampleRate) / float64(len(b)),
		}
	}

	return result
}

func (a *analyzer) Stream(samples [][2]float64) (int, bool) {

	n, ok := a.stream.Stream(samples)
	if ok {
		a.buffer.PushBack(samples[:n])
		a.bufferLen += n
		for a.bufferLen > a.bufferSize {
			frontLen := len(a.buffer.Front().Value.([][2]float64))
			a.bufferLen -= frontLen
			a.buffer.Remove(a.buffer.Front())
		}
	}
	return n, ok
}

func (a *analyzer) Err() error {
	return a.stream.Err()
}

func (a *analyzer) Buffer() [][2]float64 {
	buffer := make([][2]float64, 0, a.bufferLen)
	item := a.buffer.Front()
	for item != nil {
		buffer = append(buffer, item.Value.([][2]float64)...)
		item = item.Next()
	}

	return buffer
}

func (s Spectra) Frequencies() []float64 {
	result := make([]float64, 0, len(s))
	for _, v := range s {
		result = append(result, v.Frequency)
	}
	return result
}

func (s Spectra) Values() []float64 {
	result := make([]float64, 0, len(s))
	for _, v := range s {
		result = append(result, v.Value)
	}
	return result
}

func hammingWindow(signal [][2]float64) []float64 {
	window := make([]float64, len(signal))
	alpha := 0.54
	beta := 1 - alpha
	for i := 0; i < len(signal); i++ {
		window[i] = alpha - beta*math.Cos(2*math.Pi*float64(i)/(float64(len(signal))-1))
		window[i] *= signal[i][0]
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
