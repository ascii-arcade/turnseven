package games

import "math/rand/v2"

type CardKind int

const (
	CardNumber CardKind = iota
	CardPlus
	CardTimesTwo
	CardFreeze
	CardFlipThree
	CardSecondChance
)

type Card struct {
	Kind  CardKind
	Value int
}

// newDeck builds the 94 card deck: one 0, n copies of each number n from 1
// to 12, the +2/+4/+6/+8/+10 and x2 modifiers, and three of each action card.
func newDeck() []Card {
	deck := []Card{{Kind: CardNumber, Value: 0}}
	for n := 1; n <= 12; n++ {
		for range n {
			deck = append(deck, Card{Kind: CardNumber, Value: n})
		}
	}
	for _, v := range []int{2, 4, 6, 8, 10} {
		deck = append(deck, Card{Kind: CardPlus, Value: v})
	}
	deck = append(deck, Card{Kind: CardTimesTwo, Value: 2})
	for range 3 {
		deck = append(deck,
			Card{Kind: CardFreeze},
			Card{Kind: CardFlipThree},
			Card{Kind: CardSecondChance},
		)
	}

	shuffle(deck)
	return deck
}

func shuffle(cards []Card) {
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}
