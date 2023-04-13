package main

import (
	"context"
	"fmt"
	"os"

	"github.com/faiface/beep/wav"
	"github.com/hoani/toot"
)

func main() {
	// initialize toot

	devices, err := toot.GetInputDevices()
	if err != nil {
		panic(err)
	}
	fmt.Printf("candidates: %#v\n", devices)

	m, err := toot.NewDefaultMicrophone()
	if err != nil {
		panic(err)
	}
	defer m.Close()

	a := toot.NewAnalyzer(m, int(m.Format().SampleRate), int(m.Format().SampleRate))

	f, err := os.Create("test.wav")

	go func() {
		err = wav.Encode(f, a, m.Format())
		if err != nil {
			panic(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := m.Start(ctx); err != nil {
		panic(err)
	}

	fmt.Print("\nPress [ENTER] to finish recording! ")
	fmt.Scanln()
	m.Close()

	fmt.Print("computing power series...\n")
	freqs, powerSeries := a.GetPowerSpectrum()
	fmt.Print("plotting...\n")
	Plot(freqs, powerSeries)

}
