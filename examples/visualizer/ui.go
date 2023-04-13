package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

var viewCount = 0
var updateCount = 0

const (
	padding  = 2
	maxWidth = 80
	nBars    = 12
)

var latest update

type update [nBars]float64

type model struct {
	bars [nBars]progress.Model
}

func newUi() tea.Model {
	var bars [nBars]progress.Model
	for i := 0; i < nBars; i++ {
		bars[i] = progress.New(progress.WithDefaultGradient())
	}

	return model{
		bars: bars,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updateCount++
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		for _, bar := range m.bars {
			bar.Width = msg.Width - padding*2 - 4
			if bar.Width > maxWidth {
				bar.Width = maxWidth
			}
		}
		return m, nil

	case update:
		var cmds []tea.Cmd
		for i, bar := range m.bars {
			cmds = append(cmds, bar.SetPercent(float64(msg[i])))
		}
		latest = msg

		return m, tea.Batch(cmds...)

	default:
		var cmds []tea.Cmd
		for i, bar := range m.bars {
			bar, cmd := bar.Update(msg)
			m.bars[i] = bar.(progress.Model)
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)
	}
}

func (m model) View() string {
	viewCount++

	pad := strings.Repeat(" ", padding)

	barView := ""
	for i, bar := range m.bars {
		barView += pad + bar.ViewAs(latest[i]) + "\n"
	}
	return "\n" +
		barView + "\n\n" +
		pad + helpStyle("Press any key to quit")
}
