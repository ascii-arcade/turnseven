package games

import (
	"errors"
	"time"

	"github.com/ascii-arcade/game-template/messages"
)

const (
	minimumPlayers = 2
	maximumPlayers = 6
)

func (s *Game) Begin() error {
	return s.withErrLock(s.begin)
}

// Restart lets the host deal a fresh game to the same players once the
// current one is over.
func (s *Game) Restart(player *Player) error {
	return s.withErrLock(func() error {
		if s.Phase != PhaseGameOver || !player.IsHost() {
			return nil
		}
		return s.begin()
	})
}

func (s *Game) begin() error {
	if error := s.IsPlayerCountOk(); error != nil {
		return error
	}

	s.CurrentTurnIndex = 0
	s.DealerIndex = 0
	s.Round = 0
	s.Log = nil
	s.deck = newDeck()
	s.discard = nil
	s.ShuffledAt = time.Time{}
	s.inProgress = true

	for _, p := range s.players {
		p.Points = 0
		p.takeCards()
		p.update(messages.TableScreen)
	}

	s.startRound()
	return nil
}

func (s *Game) IsPlayerCountOk() error {
	if len(s.players) > maximumPlayers {
		return errors.New("too_many_players")
	}
	if len(s.players) < minimumPlayers {
		return errors.New("not_enough_players")
	}
	return nil
}
