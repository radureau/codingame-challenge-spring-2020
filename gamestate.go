package main

// GameState _
type GameState struct {
	*Game
	before *GameState
	turn
	myScore            ScorePoint
	opponentScore      ScorePoint
	visiblePacCount    int // all your pacs and enemy pacs in sight
	visiblePelletCount int // all pellets in sight
	// fog of war
	pacs                 map[freshness]map[Pos]*Pac
	oldestPacFreshness   freshness
	pellets              map[freshness]map[Pos]*Pellet
	oldestPelletFresness freshness
}

// Allies _
func (gs GameState) Allies() []*Pac {
	allies := make([]*Pac, 0, 5)
	for _, pac := range gs.pacs[0] {
		if pac.ally {
			allies = append(allies, pac)
		}
	}
	return allies
}

// MyProgress _
func (gs GameState) MyProgress() float64 {
	return float64(gs.myScore) * 100 / float64(gs.scoreToReach)
}

// OpntProgress _
func (gs GameState) OpntProgress() float64 {
	return float64(gs.opponentScore) * 100 / float64(gs.scoreToReach)
}

// GameProgress _
func (gs GameState) GameProgress() float64 {
	return float64(gs.turn) * 100 / float64(MaxTurn)
}

// IsOver _
func (gs GameState) IsOver() bool {
	return gs.MyProgress() >= 100 ||
		gs.GameProgress() >= 100 ||
		gs.OpntProgress() >= 100 ||
		gs.OpntProgress() > 6*gs.MyProgress() ||
		gs.MyProgress() > 6*gs.OpntProgress()
}

// trackPacFreshness fill current freshness mapper from the one used in last turn
// current holds the info with freshness = 0
func trackPacFreshness(current, before map[freshness]map[Pos]*Pac) (oldestFreshness freshness) {
	for freshness, pacs := range before {
		m := make(map[Pos]*Pac)
		for _, pac := range pacs {
			if pac.ally {
				G.alliesCount--
				continue // we lost a comrade: one of ours was killed! May he Rest In Peace †
			}
			isInView := false
			for _, p := range current[0] {
				if pac.PacID == p.PacID {
					isInView = true
					break
				}
			}
			if !isInView {
				if pac.abilityCooldown > 0 {
					pac.abilityCooldown--
				}
				if pac.speedTurnsLeft > 0 {
					pac.speedTurnsLeft--
				}
				m[pac.Pos] = pac
			}
		}
		if len(m) > 0 {
			freshness++
			current[freshness] = m
			if freshness > oldestFreshness {
				oldestFreshness = freshness
			}
		}
	}
	return oldestFreshness
}

func untrackPelletAt(pos Pos) {
	for _, m := range G.pellets {
		if _, ok := m[pos]; ok {
			delete(m, pos)
			break
		}
	}
}

// trackPelletFreshness fill current freshness mapper from the one used in last turn
// current holds the info with freshness = 0
func trackPelletFreshness(current, before map[freshness]map[Pos]*Pellet) (oldestFreshness freshness) {
	for freshness, pellets := range before {
		m := make(map[Pos]*Pellet)
		for _, plt := range pellets {
			if plt.Value == SuperPellet { // if not visible then it means it was eaten !
				continue
			}
			isInView := false
			for _, p := range current[0] {
				if plt.Pos == p.Pos {
					isInView = true
					break
				}
			}
			if !isInView {
				m[plt.Pos] = plt
			}
		}
		if len(m) > 0 {
			freshness++
			current[freshness] = m
			if freshness > oldestFreshness {
				oldestFreshness = freshness
			}
		}
	}
	return oldestFreshness
}
