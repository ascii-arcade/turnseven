package board

import (
	"fmt"
	"strings"
	"time"

	"github.com/ascii-arcade/game-template/config"
	"github.com/ascii-arcade/game-template/games"
	"github.com/ascii-arcade/game-template/keys"
	"github.com/ascii-arcade/game-template/language"
	"github.com/ascii-arcade/game-template/messages"
	"github.com/ascii-arcade/game-template/screen"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int
	screen screen.Screen
	style  lipgloss.Style

	Player *games.Player
	Game   *games.Game
}

func NewModel(width, height int, style lipgloss.Style, player *games.Player) Model {
	m := Model{
		width:  width,
		height: height,
		style:  style,
		Player: player,
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return waitForRefreshSignal(m.Player.UpdateChan)
}

func (m *Model) SetGame(game *games.Game) {
	m.Game = game
	m.screen = m.newLobbyScreen()
	m.syncScreen()
}

// syncScreen moves to the screen matching the game's state, so a player who
// reconnects or misses an update still lands in the right place.
func (m *Model) syncScreen() {
	if !m.Game.InProgress() {
		return
	}
	switch m.screen.(type) {
	case *winnerScreen:
		if m.Game.Phase != games.PhaseGameOver {
			m.screen = m.newTableScreen()
		}
	case *tableScreen:
		if m.Game.Phase == games.PhaseGameOver {
			m.screen = m.newWinnerScreen()
		}
	default:
		if m.Game.Phase == games.PhaseGameOver {
			m.screen = m.newWinnerScreen()
		} else {
			m.screen = m.newTableScreen()
		}
	}
}

// playerName renders a player's name in their color.
func (m *Model) playerName(p *games.Player) string {
	return m.style.Bold(true).Foreground(p.Color).Render(p.Name)
}

func (m *Model) playerInitials(p *games.Player) string {
	return m.style.Bold(true).Foreground(p.Color).Render(p.Initials())
}

func (m *Model) lang() *language.Language {
	return m.Player.LanguagePreference.Lang
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.PlayerUpdate:
		return m.handlePlayerUpdate(int(msg))

	case tea.KeyMsg:
		switch {
		case keys.ExitApplication.TriggeredBy(msg.String()):
			m.Game.RemovePlayer(m.Player)
			return m, tea.Quit
		}
	}

	screenModel, cmd := m.screen.Update(msg)
	return screenModel.(*Model), cmd
}

func (m Model) View() string {
	if m.width < config.MinimumWidth {
		return m.lang().Get("error", "window_too_narrow")
	}
	if m.height < config.MinimumHeight {
		return m.lang().Get("error", "window_too_short")
	}

	disconnectedPlayers := m.Game.GetDisconnectedPlayers()
	if len(disconnectedPlayers) > 0 {
		var names []string
		for _, p := range disconnectedPlayers {
			names = append(names, m.playerName(p))
		}
		return m.style.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				m.style.Align(lipgloss.Center).MarginBottom(2).Render(m.Game.Code),
				fmt.Sprintf(m.lang().Get("board", "disconnected_player"), strings.Join(names, ", ")),
				m.style.Render(fmt.Sprintf(m.lang().Get("global", "quit"), keys.ExitApplication.String(m.style))),
			),
		)
	}

	return m.screen.View()
}

func waitForRefreshSignal(ch chan int) tea.Cmd {
	return func() tea.Msg {
		return messages.PlayerUpdate(<-ch)
	}
}

func (m *Model) handlePlayerUpdate(msg int) (tea.Model, tea.Cmd) {
	switch msg {
	case messages.TableScreen:
		m.screen = m.newTableScreen()
	case messages.WinnerScreen:
		m.screen = m.newWinnerScreen()
	}
	m.syncScreen()

	cmds := []tea.Cmd{waitForRefreshSignal(m.Player.UpdateChan)}
	if recentlyShuffled(m.Game) {
		// Redraw once the shuffle notice has run its course.
		remaining := shuffleFlash - time.Since(m.Game.ShuffledAt)
		cmds = append(cmds, tea.Tick(remaining, func(time.Time) tea.Msg { return shuffleFlashDone{} }))
	}
	return m, tea.Batch(cmds...)
}

type shuffleFlashDone struct{}
