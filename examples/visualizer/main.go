package main

import (
	"context"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hoani/toot"
)

func main() {
	m := newUi()
	p := tea.NewProgram(m)

	go func() {
		if _, err := p.Run(); err != nil {
			panic(err)
		}
	}()

	if err := runAudio(p); err != nil {
		panic(err)
	}
}

func runAudio(p *tea.Program) error {
	m, err := toot.NewDefaultMicrophone()
	if err != nil {
		return err
	}
	defer m.Close()

	a := toot.NewAnalyzer(m, int(m.Format().SampleRate), int(m.Format().SampleRate/4))

	v := toot.NewVisualizer(100.0, 4000.0, 12)

	go func() {
		var samples = make([][2]float64, 128)
		for {
			_, ok := a.Stream(samples)
			if !ok {
				return
			}
		}
	}()

	go m.Start(context.Background())
	for {
		time.Sleep(time.Millisecond * 50)
		s := a.GetPowerSpectrum()
		if s == nil {
			continue
		}

		result := v.Bin(s)
		u := update{}
		for i, r := range result {
			u[i] = math.Log10(1+r*1000) * 10 // Do some log scaling to make the power spectra show up nicer.
		}

		p.Send(u)
	}
}
