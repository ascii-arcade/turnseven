package games

import (
	"slices"
	"sort"
	"sync"

	"github.com/ascii-arcade/game-template/messages"
	"github.com/charmbracelet/ssh"
)

type Game struct {
	Code string

	Settings         Settings
	CurrentTurnIndex int
	DealerIndex      int
	Round            int
	Phase            Phase
	Log              []LogEntry
	inProgress       bool
	mu               sync.Mutex
	players          []*Player

	deck      []Card
	discard   []Card
	choices   []pendingChoice
	flipJobs  []*flipJob
	dealt     int
	passTurn  bool
	flipSeven *Player
}

func (s *Game) InProgress() bool {
	return s.inProgress
}

func (s *Game) OrderedPlayers() []*Player {
	var players []*Player
	players = append(players, s.players...)
	sort.Slice(players, func(i, j int) bool {
		return players[i].TurnOrder < players[j].TurnOrder
	})

	return players
}

func (s *Game) refresh() {
	for _, p := range s.players {
		p.update(messages.Refresh)
	}
}

func (s *Game) withLock(fn func()) {
	s.mu.Lock()
	defer func() {
		s.refresh()
		s.mu.Unlock()
	}()
	fn()
}

func (s *Game) withErrLock(fn func() error) error {
	s.mu.Lock()
	defer func() {
		s.refresh()
		s.mu.Unlock()
	}()
	return fn()
}

func (s *Game) AddPlayer(player *Player, isHost bool) error {
	return s.withErrLock(func() error {
		if _, ok := s.getPlayer(player.Sess); ok {
			return nil
		}

		if s.inProgress {
			return ErrGameInProgress
		}

		maxTurnOrder := 0
		for _, p := range s.players {
			if p.TurnOrder > maxTurnOrder {
				maxTurnOrder = p.TurnOrder
			}
		}

		player.SetTurnOrder(maxTurnOrder + 1)
		if isHost {
			player.MakeHost()
		}

		player.OnDisconnect(func() {
			if !s.inProgress {
				s.RemovePlayer(player)
			}
		})

		s.players = append(s.players, player)
		return nil
	})
}

func (s *Game) RemovePlayer(player *Player) {
	s.withLock(func() {
		if player, exists := s.getPlayer(player.Sess); exists {
			close(player.UpdateChan)
			for i, p := range s.players {
				if p.Sess.User() == player.Sess.User() {
					s.players = slices.Delete(s.players, i, i+1)
					s.playerRemoved(player, i)
					break
				}
			}

			if s.GetPlayerCount(false) == 0 {
				delete(games, s.Code)
			}
		}
	})
}

func (s *Game) getPlayer(sess ssh.Session) (*Player, bool) {
	for _, p := range s.players {
		if p.Sess.User() == sess.User() {
			return p, true
		}
	}
	return nil, false
}

func (s *Game) GetDisconnectedPlayers() []*Player {
	var players []*Player
	for _, p := range s.players {
		if !p.connected {
			players = append(players, p)
		}
	}
	return players
}

func (s *Game) HasPlayer(player *Player) bool {
	_, exists := s.getPlayer(player.Sess)
	return exists
}

func (s *Game) GetPlayerCount(includeDisconnected bool) int {
	count := 0
	for _, p := range s.players {
		if includeDisconnected || p.connected {
			count++
		}
	}
	return count
}

func (s *Game) GetCurrentPlayer() *Player {
	if len(s.players) == 0 {
		return nil
	}
	return s.players[s.CurrentTurnIndex]
}

func (s *Game) Dealer() *Player {
	if len(s.players) == 0 {
		return nil
	}
	return s.players[s.DealerIndex]
}

func (s *Game) DeckCount() int {
	return len(s.deck)
}

func (s *Game) DiscardCount() int {
	return len(s.discard)
}

// GetWinner returns the sole highest scorer once anyone reaches the goal.
// A tie for the lead means another round is played.
func (s *Game) GetWinner() *Player {
	var leader *Player
	tied := false
	for _, player := range s.players {
		if leader == nil || player.Points > leader.Points {
			leader, tied = player, false
		} else if player.Points == leader.Points {
			tied = true
		}
	}
	if leader == nil || tied || leader.Points < s.Settings.EndPoints {
		return nil
	}
	return leader
}

// Standings returns the players ordered from highest to lowest score.
func (s *Game) Standings() []*Player {
	players := slices.Clone(s.players)
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].Points > players[j].Points
	})
	return players
}
