package main

import "fmt"

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
func (gs *GameState) Allies() []*Pac {
	allies := make([]*Pac, 0, 5)
	for _, pac := range gs.pacs[0] {
		if pac.ally {
			allies = append(allies, pac)
		}
	}
	return allies
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

// ReadGameState _
func (G *Game) ReadGameState() {
	defer printElapsedTime("ReadGameState")()
	if G.GameState == nil {
		G.GameState = &GameState{Game: G, turn: 1}
		G.pastStates = make([]*GameState, 0, MaxTurn-1)
	} else {
		G.pastStates = append([]*GameState{G.GameState}, G.pastStates...)
		G.GameState = &GameState{Game: G, turn: G.turn + 1, before: G.GameState}
	}
	G.Scan()
	fmt.Sscan(G.Text(), &G.myScore, &G.opponentScore)

	G.Scan()
	fmt.Sscan(G.Text(), &G.visiblePacCount)
	G.pacs = make(map[freshness]map[Pos]*Pac)
	G.pacs[0] = make(map[Pos]*Pac, G.visiblePacCount)
	for i := 0; i < G.visiblePacCount; i++ {
		pac := new(Pac)
		var _mine int
		var x, y int
		var typeID string
		G.Scan()
		fmt.Sscan(G.Text(), &pac.ID, &_mine, &x, &y, &typeID, &pac.speedTurnsLeft, &pac.abilityCooldown)
		pac.ally = _mine != 0
		pac.Pos = xy(x, y)
		switch typeID {
		case ROCK.String():
			pac.Shifumi = ROCK
		case PAPER.String():
			pac.Shifumi = PAPER
		default:
			pac.Shifumi = SCISSORS
		}
		G.pacs[0][pac.Pos] = pac
	}
	if G.turn == 1 {
		for _, pac := range G.Allies() {
			opnt := *pac
			opnt.ally = false
			opnt.Pos = opnt.sym()
			G.pacs[0][opnt.Pos] = &opnt
		}
		G.oldestPacFreshness = 0
	} else {
		G.oldestPacFreshness = trackPacFreshness(G.pacs, G.before.pacs)
		// todo: evict killed Pacs
	}

	G.Scan()
	fmt.Sscan(G.Text(), &G.visiblePelletCount)
	G.pellets = make(map[freshness]map[Pos]*Pellet)
	G.pellets[0] = make(map[Pos]*Pellet, G.visiblePelletCount)
	nSuperPellet := 0
	for i := 0; i < G.visiblePelletCount; i++ {
		pl := new(Pellet)
		var x, y int
		G.Scan()
		fmt.Sscan(G.Text(), &x, &y, &pl.Value)
		pl.Pos = xy(x, y)
		G.pellets[0][pl.Pos] = pl
		if pl.Value == SuperPellet {
			nSuperPellet++
		}
	}
	if G.turn == 1 {
		for pos := range G.graph.cells {
			if _, ok := G.pellets[0][pos]; !ok {
				if _, ok := G.pacs[0][pos]; !ok {
					G.pellets[0][pos] = &Pellet{Pos: pos, Value: NormalPellet}
				}
			}
		}
		G.oldestPelletFresness = 0
		G.scoreToReach = (ScorePoint(len(G.graph.cells)-len(G.pacs[0])-nSuperPellet)*NormalPellet+
			ScorePoint(nSuperPellet)*SuperPellet)/
			2 + 1
	} else {
		G.oldestPelletFresness = trackPelletFreshness(G.pellets, G.before.pellets)
		// evict consumed pellets
		for _, pac := range G.Allies() {
			node := G.graph.cells[pac.Pos]
			untrackPelletAt(node.Pos)
			for pos := range node.linkedWith {
				if G.pellets[0][pos].Value == Nought {
					untrackPelletAt(pos)
				}
			}
		}
	}
}
