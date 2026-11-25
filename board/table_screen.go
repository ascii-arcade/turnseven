package board

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ascii-arcade/game-template/colors"
	"github.com/ascii-arcade/game-template/games"
	"github.com/ascii-arcade/game-template/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logWidth = 40

type tableScreen struct {
	model *Model
	style lipgloss.Style
}

func (m *Model) newTableScreen() *tableScreen {
	return &tableScreen{
		model: m,
		style: m.style,
	}
}

func (s *tableScreen) Update(msg tea.Msg) (any, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.model.height, s.model.width = msg.Height, msg.Width
		return s.model, nil

	case tea.KeyMsg:
		game, player := s.model.Game, s.model.Player

		switch {
		case keys.GameHit.TriggeredBy(msg.String()):
			game.Hit(player)
		case keys.GameStay.TriggeredBy(msg.String()):
			game.Stay(player)
		case keys.GameNextRound.TriggeredBy(msg.String()):
			game.NextRound(player)
		case keys.GameTargets.TriggeredBy(msg.String()):
			game.ChooseTarget(player, keys.GameTargets.Index(msg.String()))
		}
	}

	return s.model, nil
}

func (s *tableScreen) View() string {
	game := s.model.Game
	lang := s.model.lang()

	header := s.style.Bold(true).Render(fmt.Sprintf(
		lang.Get("table", "header"),
		game.Round, game.Code, game.DeckCount(), game.DiscardCount(), game.Settings.EndPoints,
	))

	playersWidth := s.model.width - logWidth - 6
	players := s.style.Width(playersWidth).Render(s.players())

	footer := s.prompt() + "\n\n" +
		s.style.Foreground(colors.Muted).Render(fmt.Sprintf(lang.Get("global", "quit"), keys.ExitApplication.String(s.style)))

	// The log may use whatever height the players and footer leave free.
	logHeight := max(lipgloss.Height(players), s.model.height-lipgloss.Height(header)-lipgloss.Height(footer)-4)
	log := s.style.
		Width(logWidth).
		PaddingLeft(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Render(s.log(logHeight))

	return s.style.Padding(1, 2).Render(lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, players, log),
		"",
		footer,
	))
}

func (s *tableScreen) players() string {
	game := s.model.Game
	lang := s.model.lang()
	current := game.GetCurrentPlayer()
	showTurn := game.Phase == games.PhaseDealing || game.Phase == games.PhaseTurns

	var rows []string
	for _, p := range game.OrderedPlayers() {
		marker := "  "
		nameStyle := s.style.Bold(true)
		if showTurn && p == current {
			marker = "▶ "
			nameStyle = nameStyle.Foreground(colors.Highlight)
		}

		name := p.Name
		var tags []string
		if p == s.model.Player {
			tags = append(tags, lang.Get("table", "you"))
		}
		if p == game.Dealer() && game.Phase != games.PhaseRoundOver {
			tags = append(tags, lang.Get("table", "dealer"))
		}
		if len(tags) > 0 {
			name += " (" + strings.Join(tags, ", ") + ")"
		}

		info := strings.Join([]string{
			fmt.Sprintf(lang.Get("table", "total"), p.Points),
			fmt.Sprintf(lang.Get("table", "round_score"), p.RoundScore()),
			s.status(p),
		}, "  ·  ")

		rows = append(rows,
			marker+nameStyle.Render(name)+"   "+info,
			s.style.PaddingLeft(2).Render(s.model.renderLine(p)),
		)
	}
	return strings.Join(rows, "\n")
}

func (s *tableScreen) status(p *games.Player) string {
	lang := s.model.lang()
	switch {
	case p.FlippedSeven():
		return s.style.Bold(true).Foreground(colors.Success).Render(lang.Get("table", "status_flip_seven"))
	case p.Status == games.StatusBusted:
		return s.style.Foreground(colors.Error).Render(lang.Get("table", "status_busted"))
	case p.Status == games.StatusFrozen:
		return s.style.Foreground(colors.Frozen).Render(lang.Get("table", "status_frozen"))
	case p.Status == games.StatusStayed:
		return s.style.Foreground(colors.Success).Render(lang.Get("table", "status_stayed"))
	default:
		return s.style.Foreground(colors.Muted).Render(lang.Get("table", "status_active"))
	}
}

// log shows the newest entries that fit in height lines once wrapped.
func (s *tableScreen) log(height int) string {
	lang := s.model.lang()
	wrap := s.style.Width(logWidth - 2)

	var entries []string
	used := 2
	for _, entry := range slices.Backward(s.model.Game.Log) {
		args := make([]any, len(entry.Args))
		for i, arg := range entry.Args {
			if card, ok := arg.(games.Card); ok {
				arg = cardLabel(lang, card)
			}
			args[i] = arg
		}
		line := wrap.Render(fmt.Sprintf(lang.Get("log", entry.Key), args...))
		if used += lipgloss.Height(line); used > height {
			break
		}
		entries = append(entries, line)
	}
	slices.Reverse(entries)

	title := s.style.Bold(true).Render(lang.Get("table", "log_title"))
	return strings.Join(append([]string{title, ""}, entries...), "\n")
}

func (s *tableScreen) prompt() string {
	game := s.model.Game
	lang := s.model.lang()
	key := func(k keys.Keys) string { return k.String(s.style) }

	if chooser, card, targets, ok := game.PendingChoice(); ok {
		cardName := s.style.Foreground(colors.Action).Bold(true).Render(cardLabel(lang, card))
		if chooser != s.model.Player {
			return fmt.Sprintf(lang.Get("table", "choosing_target"), chooser.Name, cardName)
		}

		options := []string{fmt.Sprintf(lang.Get("table", "choose_target"), cardName)}
		for i, target := range targets {
			if i >= len(keys.GameTargets) {
				break
			}
			options = append(options, fmt.Sprintf("  %s %s", keys.GameTargets.IndexedString(i, s.style), target.Name))
		}
		return strings.Join(options, "\n")
	}

	switch game.Phase {
	case games.PhaseRoundOver:
		message := s.style.Bold(true).Render(lang.Get("table", "round_over"))
		if leader := game.Standings()[0]; leader.Points >= game.Settings.EndPoints {
			message += "  " + fmt.Sprintf(lang.Get("table", "tied_lead"), game.Settings.EndPoints)
		}
		if s.model.Player.IsHost() {
			return message + "\n" + fmt.Sprintf(lang.Get("table", "press_next_round"), key(keys.GameNextRound))
		}
		return message + "\n" + lang.Get("table", "waiting_next_round")

	case games.PhaseTurns:
		if game.IsPlayersTurn(s.model.Player) {
			return s.style.Foreground(colors.Highlight).Render(
				fmt.Sprintf(lang.Get("table", "your_turn"), key(keys.GameHit), key(keys.GameStay)),
			)
		}
	}

	return fmt.Sprintf(lang.Get("table", "waiting_for"), game.GetCurrentPlayer().Name)
}
