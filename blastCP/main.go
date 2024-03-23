package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	source   string
	dest     string
	progress progress.Model
}

type progressMsg float64

func initialModel(source, dest string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	p := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	return model{spinner: s, progress: p, source: source, dest: dest}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg { return nil })
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progressMsg:
		var cmds []tea.Cmd
		cmd := m.progress.SetPercent(float64(msg)) // Correctly use SetPercent
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n%s Copying %s to %s...\n\n%s\n\n", m.spinner.View(), m.source, m.dest, m.progress.View())
	return str
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cp-alt <source> <destination>")
		return
	}

	source, dest := os.Args[1], os.Args[2]

	m := initialModel(source, dest)
	p := tea.NewProgram(m)

	go func() {
		err := cp(source, dest, func(prog float64) {
			p.Send(progressMsg(prog))
		})
		if err != nil {
			m.err = err
		}
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func cp(src, dst string, progressCallback func(float64)) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	totalBytes := float64(sourceFileStat.Size())
	progressCallback(0.0) // Initial progress

	bytesWritten := 0
	buf := make([]byte, 4096)
	for {
		n, err := source.Read(buf)
		if n > 0 {
			nw, err := destination.Write(buf[:n])
			if err != nil {
				return err
			}
			bytesWritten += nw
			progressCallback(float64(bytesWritten) / totalBytes)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}
