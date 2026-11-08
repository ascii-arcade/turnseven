package games

import (
	"slices"

	"github.com/ascii-arcade/game-template/messages"
)

const maxLogEntries = 16

type Phase int

const (
	PhaseDealing Phase = iota
	PhaseTurns
	PhaseRoundOver
	PhaseGameOver
)

// LogEntry is translated by the viewer: Key names a "log" translation and
// Args fill its verbs, with any Card args rendered in the viewer's language.
type LogEntry struct {
	Key  string
	Args []any
}

// pendingChoice is an action card waiting for its holder to pick who gets it.
type pendingChoice struct {
	chooser *Player
	card    Card
}

// flipJob is a Flip Three in progress. Freeze and Flip Three cards drawn
// during it are held back until all three cards are flipped.
type flipJob struct {
	player    *Player
	remaining int
	deferred  []Card
}

func (s *Game) startRound() {
	for _, p := range s.players {
		s.discard = append(s.discard, p.takeCards()...)
	}

	s.Round++
	s.Phase = PhaseDealing
	s.dealt = 0
	s.choices = nil
	s.flipJobs = nil
	s.passTurn = false
	s.flipSeven = nil

	s.addLog("round_start", s.Round, s.players[s.DealerIndex].Name)
	s.advance()
}

// advance runs the round forward until it needs input from a player.
func (s *Game) advance() {
	for {
		if s.Phase != PhaseDealing && s.Phase != PhaseTurns {
			return
		}
		if s.flipSeven != nil || (s.Phase == PhaseTurns && !s.anyActive()) {
			s.endRound()
			return
		}

		if n := len(s.flipJobs); n > 0 {
			job := s.flipJobs[n-1]
			if job.remaining > 0 && job.player.IsActive() {
				job.remaining--
				s.drawFor(job.player, job)
				continue
			}

			s.flipJobs = s.flipJobs[:n-1]
			if job.player.Status == StatusBusted {
				s.discard = append(s.discard, job.deferred...)
				continue
			}
			for _, c := range job.deferred {
				s.receiveAction(job.player, c)
			}
			continue
		}

		if len(s.choices) > 0 {
			c := s.choices[0]
			targets := s.targetsFor(c)
			if len(targets) > 1 {
				return
			}

			s.choices = s.choices[1:]
			if len(targets) == 0 {
				s.discard = append(s.discard, c.card)
				s.addLog("discarded", c.chooser.Name, c.card)
				continue
			}
			s.applyChoice(c, targets[0])
			continue
		}

		if s.Phase == PhaseDealing {
			if s.dealt < len(s.players) {
				idx := (s.DealerIndex + 1 + s.dealt) % len(s.players)
				s.dealt++
				s.CurrentTurnIndex = idx
				if p := s.players[idx]; p.IsActive() {
					s.drawFor(p, nil)
				}
				continue
			}

			s.Phase = PhaseTurns
			s.CurrentTurnIndex = s.DealerIndex
			s.passTurn = true
			continue
		}

		if s.passTurn || !s.GetCurrentPlayer().IsActive() {
			s.passTurn = false
			s.CurrentTurnIndex = s.nextActiveIndex(s.CurrentTurnIndex)
		}
		return
	}
}

func (s *Game) drawFor(p *Player, job *flipJob) {
	card, ok := s.draw()
	if !ok {
		p.Status = StatusStayed
		s.addLog("deck_empty", p.Name)
		return
	}
	s.addLog("drew", p.Name, card)

	switch card.Kind {
	case CardNumber:
		if !p.hasNumber(card.Value) {
			p.Hand = append(p.Hand, card)
			if p.FlippedSeven() {
				s.flipSeven = p
				s.addLog("flip_seven", p.Name)
			}
			return
		}
		if p.HasSecondChance {
			p.HasSecondChance = false
			s.discard = append(s.discard, card, Card{Kind: CardSecondChance})
			s.addLog("second_chance_used", p.Name, card)
			return
		}
		p.Status = StatusBusted
		p.BustCard = &card
		s.addLog("busted", p.Name, card)

	case CardPlus, CardTimesTwo:
		p.Hand = append(p.Hand, card)

	default:
		if card.Kind == CardSecondChance && !p.HasSecondChance {
			p.HasSecondChance = true
			return
		}
		if job != nil {
			job.deferred = append(job.deferred, card)
			return
		}
		s.receiveAction(p, card)
	}
}

func (s *Game) receiveAction(p *Player, card Card) {
	if card.Kind == CardSecondChance && !p.HasSecondChance {
		p.HasSecondChance = true
		return
	}
	s.choices = append(s.choices, pendingChoice{chooser: p, card: card})
}

func (s *Game) targetsFor(c pendingChoice) []*Player {
	var targets []*Player
	for _, p := range s.players {
		if !p.IsActive() {
			continue
		}
		if c.card.Kind == CardSecondChance && (p == c.chooser || p.HasSecondChance) {
			continue
		}
		targets = append(targets, p)
	}
	return targets
}

func (s *Game) applyChoice(c pendingChoice, target *Player) {
	switch c.card.Kind {
	case CardFreeze:
		target.Status = StatusFrozen
		s.discard = append(s.discard, c.card)
		s.addLog("froze", c.chooser.Name, target.Name, target.RoundScore())
	case CardFlipThree:
		s.discard = append(s.discard, c.card)
		s.flipJobs = append(s.flipJobs, &flipJob{player: target, remaining: 3})
		s.addLog("flip_three", c.chooser.Name, target.Name)
	case CardSecondChance:
		target.HasSecondChance = true
		s.addLog("gave_second_chance", c.chooser.Name, target.Name)
	}
}

func (s *Game) endRound() {
	for _, c := range s.choices {
		s.discard = append(s.discard, c.card)
	}
	for _, job := range s.flipJobs {
		s.discard = append(s.discard, job.deferred...)
	}
	s.choices = nil
	s.flipJobs = nil

	for _, p := range s.players {
		p.Points += p.RoundScore()
	}
	s.addLog("round_end", s.Round)

	s.DealerIndex = (s.DealerIndex + 1) % len(s.players)

	if s.GetWinner() != nil {
		s.Phase = PhaseGameOver
		for _, p := range s.players {
			p.update(messages.WinnerScreen)
		}
		return
	}
	s.Phase = PhaseRoundOver
}

// playerRemoved keeps a game in progress playable after a player leaves:
// it drops anything waiting on them and moves the turn and deal along.
func (s *Game) playerRemoved(player *Player, index int) {
	if !s.inProgress || len(s.players) == 0 {
		return
	}

	s.discard = append(s.discard, player.takeCards()...)
	s.choices = slices.DeleteFunc(s.choices, func(c pendingChoice) bool {
		return c.chooser == player
	})
	s.flipJobs = slices.DeleteFunc(s.flipJobs, func(job *flipJob) bool {
		return job.player == player
	})

	if index < s.DealerIndex {
		s.DealerIndex--
	}
	s.DealerIndex %= len(s.players)
	if index < s.CurrentTurnIndex {
		s.CurrentTurnIndex--
	} else if index == s.CurrentTurnIndex && s.Phase == PhaseTurns {
		s.CurrentTurnIndex = (s.CurrentTurnIndex - 1 + len(s.players)) % len(s.players)
		s.passTurn = true
	}
	s.CurrentTurnIndex %= len(s.players)
	s.dealt = min(s.dealt, len(s.players))

	s.advance()
}

func (s *Game) draw() (Card, bool) {
	if len(s.deck) == 0 {
		if len(s.discard) == 0 {
			return Card{}, false
		}
		s.deck, s.discard = s.discard, nil
		shuffle(s.deck)
		s.addLog("reshuffle")
	}

	card := s.deck[len(s.deck)-1]
	s.deck = s.deck[:len(s.deck)-1]
	return card, true
}

func (s *Game) anyActive() bool {
	return slices.ContainsFunc(s.players, (*Player).IsActive)
}

func (s *Game) nextActiveIndex(from int) int {
	for i := 1; i <= len(s.players); i++ {
		idx := (from + i) % len(s.players)
		if s.players[idx].IsActive() {
			return idx
		}
	}
	return from
}

func (s *Game) addLog(key string, args ...any) {
	s.Log = append(s.Log, LogEntry{Key: key, Args: args})
	if len(s.Log) > maxLogEntries {
		s.Log = s.Log[len(s.Log)-maxLogEntries:]
	}
}
