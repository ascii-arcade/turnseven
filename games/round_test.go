package games

import (
	"math/rand/v2"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := newDeck()
	if len(deck) != 94 {
		t.Fatalf("deck has %d cards, want 94", len(deck))
	}

	numbers := map[int]int{}
	kinds := map[CardKind]int{}
	for _, c := range deck {
		kinds[c.Kind]++
		if c.Kind == CardNumber {
			numbers[c.Value]++
		}
	}
	if numbers[0] != 1 {
		t.Errorf("got %d zeros, want 1", numbers[0])
	}
	for n := 1; n <= 12; n++ {
		if numbers[n] != n {
			t.Errorf("got %d copies of %d, want %d", numbers[n], n, n)
		}
	}
	want := map[CardKind]int{CardNumber: 79, CardPlus: 5, CardTimesTwo: 1, CardFreeze: 3, CardFlipThree: 3, CardSecondChance: 3}
	for kind, count := range want {
		if kinds[kind] != count {
			t.Errorf("kind %d: got %d, want %d", kind, kinds[kind], count)
		}
	}
}

func TestRoundScore(t *testing.T) {
	num := func(v int) Card { return Card{Kind: CardNumber, Value: v} }

	tests := []struct {
		name string
		hand []Card
		bust bool
		want int
	}{
		{"numbers", []Card{num(3), num(7), num(12)}, false, 22},
		{"x2 doubles numbers only", []Card{num(5), num(6), {Kind: CardTimesTwo}, {Kind: CardPlus, Value: 4}}, false, 26},
		{"flip seven bonus", []Card{num(0), num(1), num(2), num(3), num(4), num(5), num(6)}, false, 21 + 15},
		{"bust scores zero", []Card{num(9), {Kind: CardPlus, Value: 10}}, true, 0},
	}
	for _, tt := range tests {
		p := &Player{Hand: tt.hand}
		if tt.bust {
			p.Status = StatusBusted
		}
		if got := p.RoundScore(); got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestSecondChanceSavesBust(t *testing.T) {
	p := &Player{Hand: []Card{{Kind: CardNumber, Value: 8}}, HasSecondChance: true}
	g := newTestGame(p, &Player{})
	g.deck = []Card{{Kind: CardNumber, Value: 8}}

	g.drawFor(p, nil)

	if p.Status != StatusActive || p.HasSecondChance || len(p.Hand) != 1 {
		t.Fatalf("second chance not used: status=%d sc=%v hand=%v", p.Status, p.HasSecondChance, p.Hand)
	}
}

func TestFlipThreeDefersFreezeUntilDone(t *testing.T) {
	a, b := &Player{Name: "a"}, &Player{Name: "b"}
	g := newTestGame(a, b)
	g.Phase = PhaseTurns
	// Drawn from the end: Freeze, then 4, then 5.
	g.deck = []Card{{Kind: CardNumber, Value: 5}, {Kind: CardNumber, Value: 4}, {Kind: CardFreeze}}
	g.flipJobs = []*flipJob{{player: a, remaining: 3}}

	g.advance()

	if len(a.Hand) != 2 {
		t.Fatalf("a should have flipped all three cards, hand=%v", a.Hand)
	}
	if _, card, targets, ok := g.PendingChoice(); !ok || card.Kind != CardFreeze || len(targets) != 2 {
		t.Fatalf("expected a pending freeze for 2 targets, got ok=%v card=%v targets=%d", ok, card, len(targets))
	}
}

// TestRandomGames plays many games with random decisions, checking that no
// card is ever lost or duplicated and that every game finishes.
func TestRandomGames(t *testing.T) {
	for game := range 300 {
		players := make([]*Player, 2+game%5)
		for i := range players {
			players[i] = &Player{Name: string(rune('A' + i)), isHost: i == 0}
		}
		g := newTestGame(players...)
		g.deck = newDeck()
		g.inProgress = true
		g.startRound()

		for step := 0; g.Phase != PhaseGameOver; step++ {
			if step > 100000 {
				t.Fatalf("game %d did not finish", game)
			}
			checkCardCount(t, g)

			if _, _, targets, ok := g.pendingChoice(); ok {
				c := g.choices[0]
				g.choices = g.choices[1:]
				g.applyChoice(c, targets[rand.IntN(len(targets))])
				g.advance()
				continue
			}

			switch g.Phase {
			case PhaseRoundOver:
				g.startRound()
			case PhaseTurns:
				p := g.GetCurrentPlayer()
				if !g.canAct(p) {
					t.Fatalf("game %d: current player %s cannot act", game, p.Name)
				}
				if p.RoundScore() < 20+rand.IntN(20) {
					g.drawFor(p, nil)
				} else {
					p.Status = StatusStayed
				}
				g.passTurn = true
				g.advance()
			default:
				t.Fatalf("game %d stuck in phase %d", game, g.Phase)
			}
		}

		if w := g.GetWinner(); w == nil || w.Points < 200 {
			t.Fatalf("game %d ended without a valid winner", game)
		}
	}
}

func newTestGame(players ...*Player) *Game {
	return &Game{Settings: NewSettings(), players: players}
}

func checkCardCount(t *testing.T, g *Game) {
	t.Helper()
	total := len(g.deck) + len(g.discard) + len(g.choices)
	for _, p := range g.players {
		total += len(p.Hand)
		if p.BustCard != nil {
			total++
		}
		if p.HasSecondChance {
			total++
		}
	}
	for _, job := range g.flipJobs {
		total += len(job.deferred)
	}
	if total != 94 {
		t.Fatalf("card count is %d, want 94", total)
	}
}
