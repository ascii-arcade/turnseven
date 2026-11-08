package games

func (s *Game) Hit(player *Player) {
	s.withLock(func() {
		if !s.canAct(player) {
			return
		}
		s.drawFor(player, nil)
		s.passTurn = true
		s.advance()
	})
}

func (s *Game) Stay(player *Player) {
	s.withLock(func() {
		if !s.canAct(player) {
			return
		}
		player.Status = StatusStayed
		s.addLog("stayed", player.Name, player.RoundScore())
		s.passTurn = true
		s.advance()
	})
}

// ChooseTarget gives the pending action card held by player to target.
func (s *Game) ChooseTarget(player *Player, targetIndex int) {
	s.withLock(func() {
		chooser, _, targets, ok := s.pendingChoice()
		if !ok || chooser != player || targetIndex < 0 || targetIndex >= len(targets) {
			return
		}

		c := s.choices[0]
		s.choices = s.choices[1:]
		s.applyChoice(c, targets[targetIndex])
		s.advance()
	})
}

func (s *Game) NextRound(player *Player) {
	s.withLock(func() {
		if s.Phase != PhaseRoundOver || !player.IsHost() {
			return
		}
		s.startRound()
	})
}

// PendingChoice reports an action card waiting for its holder to pick a target.
func (s *Game) PendingChoice() (chooser *Player, card Card, targets []*Player, ok bool) {
	return s.pendingChoice()
}

func (s *Game) pendingChoice() (*Player, Card, []*Player, bool) {
	if len(s.choices) == 0 {
		return nil, Card{}, nil, false
	}
	c := s.choices[0]
	return c.chooser, c.card, s.targetsFor(c), true
}

func (s *Game) IsPlayersTurn(player *Player) bool {
	return s.Phase == PhaseTurns && len(s.choices) == 0 && len(s.flipJobs) == 0 && s.GetCurrentPlayer() == player
}

func (s *Game) canAct(player *Player) bool {
	return s.IsPlayersTurn(player) && player.IsActive()
}
