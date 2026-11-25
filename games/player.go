package games

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ascii-arcade/game-template/language"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

const (
	flipSevenCount = 7
	flipSevenBonus = 15
)

type RoundStatus int

const (
	StatusActive RoundStatus = iota
	StatusStayed
	StatusFrozen
	StatusBusted
)

type Player struct {
	Name      string
	Color     lipgloss.Color
	Points    int
	TurnOrder int

	Hand            []Card
	BustCard        *Card
	HasSecondChance bool
	Status          RoundStatus

	isHost    bool
	connected bool

	UpdateChan         chan int
	LanguagePreference *language.LanguagePreference

	Sess ssh.Session

	onDisconnect []func()
	ctx          context.Context
}

func (p *Player) SetName(name string) *Player {
	p.Name = name
	return p
}

func (p *Player) SetTurnOrder(order int) *Player {
	p.TurnOrder = order
	return p
}

func (p *Player) MakeHost() *Player {
	p.isHost = true
	return p
}

func (p *Player) IsHost() bool {
	return p.isHost
}

func (p *Player) OnDisconnect(fn func()) {
	p.onDisconnect = append(p.onDisconnect, fn)
}

// Initials shortens the name to the first letter of each word.
func (p *Player) Initials() string {
	var initials []rune
	for _, word := range strings.Fields(p.Name) {
		r, _ := utf8.DecodeRuneInString(word)
		initials = append(initials, unicode.ToUpper(r))
	}
	return string(initials)
}

func (p *Player) IsActive() bool {
	return p.Status == StatusActive
}

func (p *Player) hasNumber(value int) bool {
	for _, c := range p.Hand {
		if c.Kind == CardNumber && c.Value == value {
			return true
		}
	}
	return false
}

func (p *Player) NumberCount() int {
	count := 0
	for _, c := range p.Hand {
		if c.Kind == CardNumber {
			count++
		}
	}
	return count
}

func (p *Player) FlippedSeven() bool {
	return p.NumberCount() >= flipSevenCount
}

// RoundScore sums the number cards, doubles that sum for a x2, then adds the
// + modifiers and the Flip 7 bonus. A bust scores nothing.
func (p *Player) RoundScore() int {
	if p.Status == StatusBusted {
		return 0
	}

	numbers, bonus, double := 0, 0, false
	for _, c := range p.Hand {
		switch c.Kind {
		case CardNumber:
			numbers += c.Value
		case CardPlus:
			bonus += c.Value
		case CardTimesTwo:
			double = true
		}
	}
	if double {
		numbers *= 2
	}

	score := numbers + bonus
	if p.FlippedSeven() {
		score += flipSevenBonus
	}
	return score
}

// takeCards empties the player's line for the next round, returning every
// card they held so it can go on the discard pile.
func (p *Player) takeCards() []Card {
	cards := p.Hand
	if p.BustCard != nil {
		cards = append(cards, *p.BustCard)
	}
	if p.HasSecondChance {
		cards = append(cards, Card{Kind: CardSecondChance})
	}

	p.Hand = nil
	p.BustCard = nil
	p.HasSecondChance = false
	p.Status = StatusActive
	return cards
}

func (p *Player) update(code int) {
	select {
	case p.UpdateChan <- code:
	default:
	}
}
