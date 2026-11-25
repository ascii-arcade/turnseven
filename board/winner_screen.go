package board

import (
	"fmt"

	"github.com/ascii-arcade/game-template/colors"
	"github.com/ascii-arcade/game-template/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// nameWidth lines up the final scores; generated names run up to 25 characters.
const nameWidth = 26

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

	case tea.KeyMsg:
		switch {
		case keys.GameRestart.TriggeredBy(msg.String()):
			_ = s.model.Game.Restart(s.model.Player)
		}
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
		rows = append(rows, s.model.style.Bold(true).Render(fmt.Sprintf(lang.Get("board", "winner"), s.model.playerName(winner))), "")
	}

	rows = append(rows, s.style.Underline(true).Render(lang.Get("board", "final_scores")))
	for i, p := range s.model.Game.Standings() {
		name := s.style.Width(nameWidth).Render(s.model.playerName(p))
		rows = append(rows, fmt.Sprintf("%d. %s %s", i+1, name, fmt.Sprintf(lang.Get("board", "points"), p.Points)))
	}
	rows = append(rows, "", s.restartPrompt(), fmt.Sprintf(lang.Get("global", "quit"), keys.ExitApplication.String(s.style)))

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

func (s *winnerScreen) restartPrompt() string {
	lang := s.model.lang()
	if !s.model.Player.IsHost() {
		return lang.Get("board", "waiting_for_restart")
	}
	if err := s.model.Game.IsPlayerCountOk(); err != nil {
		return s.style.Foreground(colors.Error).Render(lang.Get("error", err.Error()))
	}
	return fmt.Sprintf(lang.Get("board", "press_to_restart"), keys.GameRestart.String(s.style))
}
