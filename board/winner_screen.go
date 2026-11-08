package board

import (
	"fmt"

	"github.com/ascii-arcade/game-template/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type winnerScreen struct {
	model *Model
	style lipgloss.Style
}

func (m *Model) newWinnerScreen() *winnerScreen {
	return &winnerScreen{
		model: m,
		style: m.style,
	}
}

func (s *winnerScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
	}
	return s.model, nil
}

func (s *winnerScreen) View() string {
	lang := s.model.lang()

	rows := []string{
		s.model.style.Bold(true).Render(lang.Get("board", "game_over")),
		"",
	}
	if winner := s.model.Game.GetWinner(); winner != nil {
		rows = append(rows, s.model.style.Bold(true).Render(fmt.Sprintf(lang.Get("board", "winner"), winner.Name)), "")
	}

	rows = append(rows, s.style.Underline(true).Render(lang.Get("board", "final_scores")))
	for i, p := range s.model.Game.Standings() {
		rows = append(rows, fmt.Sprintf("%d. %-24s %s", i+1, p.Name, fmt.Sprintf(lang.Get("board", "points"), p.Points)))
	}
	rows = append(rows, "", fmt.Sprintf(lang.Get("global", "quit"), keys.ExitApplication.String(s.style)))

	return s.style.Width(s.model.width).Height(s.model.height).Render(
		lipgloss.Place(
			s.model.width,
			s.model.height,
			lipgloss.Center,
			lipgloss.Center,
			s.style.
				Padding(2, 4).
				BorderStyle(lipgloss.NormalBorder()).
				Render(lipgloss.JoinVertical(lipgloss.Left, rows...)),
		),
	)
}
