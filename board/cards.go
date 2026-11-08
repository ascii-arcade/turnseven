package board

import (
	"fmt"

	"github.com/ascii-arcade/game-template/colors"
	"github.com/ascii-arcade/game-template/games"
	"github.com/ascii-arcade/game-template/language"
	"github.com/charmbracelet/lipgloss"
)

func cardLabel(lang *language.Language, card games.Card) string {
	switch card.Kind {
	case games.CardNumber:
		return fmt.Sprint(card.Value)
	case games.CardPlus:
		return fmt.Sprintf("+%d", card.Value)
	case games.CardTimesTwo:
		return "x2"
	case games.CardFreeze:
		return lang.Get("card", "freeze")
	case games.CardFlipThree:
		return lang.Get("card", "flip_three")
	default:
		return lang.Get("card", "second_chance")
	}
}

func cardColor(card games.Card) lipgloss.Color {
	switch card.Kind {
	case games.CardNumber:
		return colors.Numbers[card.Value]
	case games.CardPlus:
		return colors.Modifier
	case games.CardTimesTwo:
		return colors.Multiply
	default:
		return colors.Action
	}
}

// renderCard draws a card face as a three line box.
func renderCard(style lipgloss.Style, label string, color lipgloss.Color) string {
	return style.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Foreground(color).
		Bold(true).
		Width(3).
		Align(lipgloss.Center).
		Render(label)
}

func (m *Model) renderLine(p *games.Player) string {
	var faces []string
	for _, card := range p.Hand {
		faces = append(faces, renderCard(m.style, cardLabel(m.lang(), card), cardColor(card)))
	}
	if p.HasSecondChance {
		faces = append(faces, renderCard(m.style, m.lang().Get("card", "second_chance_short"), colors.Action))
	}
	if p.BustCard != nil {
		faces = append(faces, renderCard(m.style, cardLabel(m.lang(), *p.BustCard), colors.Error))
	}
	if len(faces) == 0 {
		return "\n\n"
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, faces...)
}
