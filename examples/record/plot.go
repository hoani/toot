package main

import (
	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func Plot(x, y []float64) {
	if len(x) != len(y) {
		return
	}
	// Create some example data
	n := len(x) / 10
	xys := make(plotter.XYs, n)
	for i := range xys {
		xys[i].X = x[i]
		xys[i].Y = y[i]
	}

	// Create a new plot
	p := plot.New()

	// Add a line to the plot
	line, err := plotter.NewLine(xys)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA64{R: 30000, G: 30000, B: 0, A: 0}
	p.Add(line)

	// Set the plot title and axes labels
	p.Title.Text = "Example Plot"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	// Save the plot to a PNG file
	if err := p.Save(80*vg.Inch, 40*vg.Inch, "plot.png"); err != nil {
		panic(err)
	}
}
