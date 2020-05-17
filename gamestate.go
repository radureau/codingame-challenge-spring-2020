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
	superPelletCount   int
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

// Opnts _
func (gs GameState) Opnts() []*Pac {
	opnts := make([]*Pac, 0, 5)
	for _, m := range gs.pacs {
		for _, pac := range m {
			if !pac.ally {
				opnts = append(opnts, pac)
			}
		}
	}
	return opnts
}

// Pacs _
func (gs GameState) Pacs() []*Pac {
	pacs := make([]*Pac, 0, 10)
	for _, m := range gs.pacs {
		for _, pac := range m {
			pacs = append(pacs, pac)
		}
	}
	return pacs
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

// PelletsOnNodes _
func (gs GameState) PelletsOnNodes(nodes ...Node) []*Pellet {
	plts := make([]*Pellet, 0, len(nodes))
	for _, m := range gs.pellets {
		for _, plt := range m {
			for _, n := range nodes {
				if plt.Pos == n.Pos {
					plts = append(plts, plt)
					if len(plts) == len(nodes) {
						return plts
					}
				}
			}
		}
	}
	return plts
}

func (gs GameState) getPacAt(pos Pos) *Pac {
	for _, m := range gs.pacs {
		if _, ok := m[pos]; ok {
			return m[pos]
		}
	}
	return nil
}

// trackPacFreshness fill current freshness mapper from the one used in last turn
// current holds the info with freshness = 0
func trackPacFreshness(current, before map[freshness]map[Pos]*Pac) (oldestFreshness freshness) {
	for freshness, pacs := range before {
		m := make(map[Pos]*Pac)
		for _, pac := range pacs {
			isInView := false
			for _, p := range current[0] {
				if pac.PacID == p.PacID {
					isInView = true
					pac.freshness = 0
					break
				}
			}
			if !isInView {
				if pac.ally {
					G.alliesCount--
					continue // we lost a comrade: one of ours was killed! May he Rest In Peace †
				}
				if pac.abilityCooldown > 0 {
					pac.abilityCooldown--
				}
				if pac.speedTurnsLeft > 0 {
					pac.speedTurnsLeft--
				}
				pac.freshness++
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

func (gs *GameState) untrackPelletAt(pos Pos) {
	for _, m := range gs.pellets {
		if _, ok := m[pos]; ok {
			delete(m, pos)
			break
		}
	}
}

func (gs GameState) getPelletAt(pos Pos) *Pellet {
	for _, m := range gs.pellets {
		if _, ok := m[pos]; ok {
			return m[pos]
		}
	}
	return nil
}

// trackPelletFreshness fill current freshness mapper from the one used in last turn
// current holds the info with freshness = 0
func trackPelletFreshness(current, before map[freshness]map[Pos]*Pellet) (oldestFreshness freshness) { // return eaten pellets positions
	for freshness, pellets := range before {
		m := make(map[Pos]*Pellet)
		for _, plt := range pellets {
			isInView := false
			for _, p := range current[0] {
				if plt.Pos == p.Pos {
					isInView = true
					break
				}
			}
			if !isInView {
				if plt.Value == SuperPellet { // if not visible then it means it was eaten !
					continue
				}
				m[plt.Pos] = plt
				for _, pac := range G.Pacs() {
					plt.hotnessForPac[pac.PacID]--
				}
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

// order pac by distance(well, from last seen position at least)
func (gs GameState) closestPacToPos(pos Pos, ally *bool) []*Pac {
	s := SortByDistanceToPos{Pos: pos}
	for _, m := range gs.pacs {
		if ally == nil {
			for pos := range m {
				s.positions = append(s.positions, pos)
			}
		} else {
			for _, pac := range m {
				if pac.ally == *ally {
					s.positions = append(s.positions, pac.Pos)
				}
			}
		}
	}
	s.Sort()
	pacs := []*Pac{}
	for _, pos := range s.positions {
		pacs = append(pacs, gs.getPacAt(pos))
	}
	return pacs
}
func (gs GameState) closestAlliesToPos(pos Pos) []*Pac {
	filterAllies := true
	return gs.closestPacToPos(pos, &filterAllies)
}
func (gs GameState) closestOpntsToPos(pos Pos) []*Pac {
	filterAllies := false
	return gs.closestPacToPos(pos, &filterAllies)
}
func (gs GameState) closestPacToPosAmong(pos Pos, among ...*Pac) []*Pac {
	s := SortByDistanceToPos{Pos: pos}
	for _, pac := range among {
		s.positions = append(s.positions, pac.Pos)
	}
	s.Sort()
	pacs := []*Pac{}
	for _, pos := range s.positions {
		pacs = append(pacs, gs.getPacAt(pos))
	}
	return pacs
}

func (gs GameState) closestPelletToPosAmong(pos Pos, among ...*Pellet) []*Pellet {
	s := SortByDistanceToPos{Pos: pos}
	for _, plt := range among {
		s.positions = append(s.positions, plt.Pos)
	}
	s.Sort()
	plts := []*Pellet{}
	for _, pos := range s.positions {
		plts = append(plts, gs.getPelletAt(pos))
	}
	return plts
}
func (gs GameState) closestPelletToPosAmongMap(pos Pos, among map[Pos]*Pellet) []*Pellet {
	s := SortByDistanceToPos{Pos: pos}
	for _, plt := range among {
		s.positions = append(s.positions, plt.Pos)
	}
	s.Sort()
	plts := []*Pellet{}
	for _, pos := range s.positions {
		plts = append(plts, gs.getPelletAt(pos))
	}
	return plts
}
