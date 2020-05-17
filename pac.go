package main

import "fmt"

// Pac _
type Pac struct {
	PacID
	*Node
	Shifumi
	speedTurnsLeft  turn
	abilityCooldown turn
	freshness
	speed
	play play
}

// PacID _
type PacID struct {
	ID   int  // pacID: pac number (unique within a team)
	ally bool // true if this pac is yours
}

// WithFreshness _
func (pac *Pac) WithFreshness(f freshness) *Pac { pac.freshness = f; return pac }

// Shifumi _
type Shifumi int

// Pac type
const (
	ROCK = Shifumi(iota)
	PAPER
	SCISSORS
)

// Beats _
func (s Shifumi) Beats(sh Shifumi) bool {
	return sh == s.ScaredOf()
}

// ScaredOf _
func (s Shifumi) ScaredOf() Shifumi {
	switch s {
	case ROCK:
		return PAPER
	case PAPER:
		return SCISSORS
	default:
		return ROCK
	}
}

type play struct {
	playType
	Pos
	Shifumi
}

type playType string

// play type
const (
	MOVE   = playType("MOVE ")   // MOVE pacId x y
	SPEED  = playType("SPEED ")  // SPEED pacId
	SWITCH = playType("SWITCH ") // SWITCH pacId SHIFUMI
)

// Play format decision to string
func (pac Pac) Play() string {
	pac.Think()
	switch pac.play.playType {
	case MOVE:
		x, y := pac.play.xy()
		return fmt.Sprint(MOVE, pac.ID, x, y)
	case SPEED:
		return fmt.Sprint(SPEED, pac.ID)
	case SWITCH:
		return fmt.Sprint(SWITCH, pac.ID, pac.play.Shifumi)
	}
	return ""
}

// Think _
func (pac *Pac) Think() {
	defer printElapsedTime(fmt.Sprintf("pac[%d].Think", pac.ID))()
	// debug(pac.abilityCooldown)
	// if pac.abilityCooldown == 0 {
	// pac.playType = SPEED
	// for _, threat := range G.closestPacToPosAmong(pac.Pos, G.Opnts()...) {
	// 	debug(threat)
	// 	pac.playType = SWITCH
	// 	pac.play.Shifumi = threat.ScaredOf()
	// 	break
	// }
	// } else {
	pac.play.playType = MOVE
	var maxHotness hotness
	for f, m := range G.pellets {
		for _, plt := range G.closestPelletToPosAmongMap(pac.Pos, m) {
			if f == 0 && G.superPelletCount == 0 && G.graph.dists[Move{pac.Pos, plt.Pos}] > Dist((pac.speed+1)*3) {
				// closest pellet is far enough for us to be intested in lessening fog of war

				break
			}
			if howHot := plt.hotnessForPac[pac.PacID]; maxHotness < howHot {
				maxHotness = howHot
				pac.play.Pos = plt.Pos
			}
		}
	}
	// }
}

// Threats _
func (pac *Pac) Threats() []*Pac {
	threats := []*Pac{}
	for _, opnt := range G.Opnts() {
		if !opnt.Beats(pac.Shifumi) {
			continue
		}
		if G.graph.influences[opnt.speed][opnt.Pos].
			containsCell(opnt.abilityCooldown, G.graph.nodes[pac.Pos]) {
			threats = append(threats, opnt)
		}
	}
	return threats
}
