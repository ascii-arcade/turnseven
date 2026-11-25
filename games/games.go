package games

import (
	"errors"
	"math/rand/v2"
	"slices"

	"github.com/ascii-arcade/game-template/colors"
	"github.com/ascii-arcade/game-template/generaterandom"
)

var (
	ErrGameInProgress = errors.New("game_already_in_progress")
	ErrGameNotFound   = errors.New("game_not_found")
)

var games = make(map[string]*Game)

func New() *Game {
	palette := slices.Clone(colors.Players)
	rand.Shuffle(len(palette), func(i, j int) {
		palette[i], palette[j] = palette[j], palette[i]
	})

	game := &Game{
		Code:     generaterandom.Code(),
		Settings: NewSettings(),
		players:  make([]*Player, 0),
		palette:  palette,
	}
	games[game.Code] = game

	return game
}

func GetOpenGame(code string) (*Game, error) {
	game, exists := games[code]
	if !exists {
		return nil, ErrGameNotFound
	}
	if game.inProgress {
		return game, ErrGameInProgress
	}

	return game, nil
}

func GetAll() map[string]*Game {
	return games
}
