package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

func debug(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func printElapsedTime(name string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		// codingame environment seems to multiply cpu time by a factor of 4
		debug(fmt.Sprintf("\t%s\ttook %dms", name, elapsed.Milliseconds()*4))
	}
}

// Game _
type Game struct {
	*bufio.Scanner
	width  int // width: size of the grid
	height int // height: top left corner is (x=0, y=0)
	graph  *Graph
	*GameState
	pastStates                 []*GameState // from latest to oldest
	scoreToReach               ScorePoint
	alliesCount, opponentCount int
}

// func (gs *GameState) String() string {
// 	return fmt.Sprintf("GameState#%d", gs.turn)
// }

func (G *Game) String() string {
	s := G.GameState.String()
	for _, gs := range G.pastStates {
		s += "\n" + gs.String()
	}
	return s
}

// func (G *Game) String() string {
func (gs *GameState) String() string {
	// gs := G.GameState
	s := ""
	for y := 0; y < G.height; y++ {
		runes := make([]rune, G.width)
		for x := 0; x < G.width; x++ {
			if cell, ok := G.graph.cells[xy(x, y)]; ok {
				r := ' '
				for f, m := range gs.pellets {
					if pl, ok := m[cell.Pos]; ok {
						if f == 0 {
							r = pl.Rune()
						} else {
							r = '?'
						}
					}
				}
				if pac, ok := gs.pacs[0][cell.Pos]; ok {
					r = pac.Rune()
				}
				runes[x] = r
			} else {
				runes[x] = '#'
			}
		}
		s += string(runes) + "\n"
	}
	return s
}

type pacID int

// PacID _
type PacID struct {
	ID   pacID // pacID: pac number (unique within a team)
	ally bool  // true if this pac is yours
}

// Shifumi _
type Shifumi int

// Pac type
const (
	ROCK = Shifumi(iota)
	PAPER
	SCISSORS
)

func (s Shifumi) String() string {
	return []string{"ROCK", "PAPER", "SCISSORS"}[s]
}

// Rune for debug purpose
func (s Shifumi) Rune() rune {
	return []rune{'✊', '✋', '✌'}[s]
}

// Pac _
type Pac struct {
	PacID
	Pos
	Shifumi
	speedTurnsLeft  turn
	abilityCooldown turn
}

// Rune UPPERCASE means ally
func (p Pac) Rune() rune {
	r := []rune{'𝑃', '𝐹', '𝐶'}[p.Shifumi]
	if !p.ally {
		r -= 26
	}
	return r
	// if p.ally {
	// 	return '@'
	// }
	// return '&'
}

// GameState _
type GameState struct {
	*Game
	before *GameState
	turn
	myScore              ScorePoint
	opponentScore        ScorePoint
	visiblePacCount      int // all your pacs and enemy pacs in sight
	visiblePelletCount   int // all pellets in sight
	pacs                 map[freshness]map[Pos]*Pac
	oldestPacFreshness   freshness
	pellets              map[freshness]map[Pos]*Pellet
	oldestPelletFresness freshness
}

type turn int

func (t turn) freshness() freshness { return freshness(G.turn - t) }
func (f freshness) turn() turn      { return G.turn - turn(f) }

// MaxTurn _
const MaxTurn = turn(200)

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
			cell := G.graph.cells[pac.Pos]
			untrackPelletAt(cell.Pos)
			for pos := range cell.linkedWith {
				if G.pellets[0][pos].Value == Nought {
					untrackPelletAt(pos)
				}
			}
		}
	}
}

func (G *Game) scanWidthAndHeight() {
	G.Scan()
	fmt.Sscan(G.Text(), &G.width, &G.height)
}
func (G *Game) buildGraph() {
	defer printElapsedTime("buildGraph")()
	G.graph = NewGraph(G.width * G.height)
	for y := 0; y < G.height; y++ {
		G.Scan()
		const floor = ' '
		for x, r := range G.Text() { // one line of the grid: space " " is floor, pound "#" is wall
			if r == floor {
				G.graph.createCell(x, y)
			}
		}
	}
	G.graph.linkTogether()
	G.graph.compute()
}

type speed int

const (
	speed1 speed = iota
	speed2
	// Nspeed number of speeds
	Nspeed
)

// influence: from a position, in how many turns can I get to how many cells (ordered by dist) ?
type influence map[turn][]*Cell

func (nflc influence) addCells(at turn, cells ...*Cell) {
	nflc[at] = append(nflc[at], cells...)
}
func (nflc influence) at(t turn) []*Cell {
	return nflc[t]
}
func (nflc influence) containsCell(at turn, cell *Cell) bool {
	for _, c := range nflc[at] {
		if c.Pos == cell.Pos {
			return true
		}
	}
	return false
}

type path []*Cell

// Graph _
type Graph struct {
	cells      map[Pos]*Cell
	positions  []Pos                     // sorted
	dists      map[Move]Dist             // without any bumps
	paths      map[Move]path             // without any bumps
	influences [Nspeed]map[Pos]influence // influences with speed1 at index 0 and influences with speed2 at index 1
}

// NewGraph _
func NewGraph(capacity int) *Graph {
	g := new(Graph)
	g.cells = make(map[Pos]*Cell, capacity)
	g.positions = make([]Pos, 0, capacity)
	g.dists = make(map[Move]Dist, capacity)
	g.paths = make(map[Move]path, capacity)
	g.influences = [Nspeed]map[Pos]influence{make(map[Pos]influence, capacity), make(map[Pos]influence, capacity)}
	return g
}
func (g *Graph) createCell(x, y int) {
	nC := &Cell{
		Pos:        xy(x, y),
		neighbours: make([]*Cell, 0, 4),
		linkedWith: make(map[Pos]*Cell),
	}
	g.cells[nC.Pos] = nC
	g.positions = append(g.positions, nC.Pos)
}
func (g Graph) linkTogether() {
	for _, cell := range g.cells {
		for _, dir := range Directions {
			if c, ok := g.cells[cell.ToDirection(dir)]; ok {
				cell.neighbours = append(cell.neighbours, c)
			}
			current, ok := g.cells[cell.ToDirection(dir)]
			for ok {
				if current.Pos == cell.Pos {
					break
				}
				cell.linkedWith[current.Pos] = current
				current, ok = g.cells[current.ToDirection(dir)]
			}
		}
	}
}
func (g Graph) breadthFirstSearch(from *Cell, compute func(cell *Cell, dist Dist, visited []*Cell)) {
	visited := make([]*Cell, 0, len(g.cells))
	markAsVisited := func(c *Cell) {
		visited = append(visited, c)
	}
	wasVisited := func(cell *Cell) bool {
		for _, c := range visited {
			if c.Pos == cell.Pos {
				return true
			}
		}
		return false
	}
	start := from
	if start == nil {
		start = g.cells[Pos(g.positions[0])]
	}
	markAsVisited(start)
	type toVisit struct {
		*Cell
		Dist
	}
	visits := []toVisit{{start, 0}}
	for len(visits) > 0 {
		var current toVisit
		current, visits = visits[0], visits[1:]
		compute(current.Cell, current.Dist, visited)
		for _, c := range current.neighbours {
			if !wasVisited(c) {
				markAsVisited(c)
				visits = append(visits, toVisit{c, current.Dist + 1})
			}
		}
	}
}
func (g Graph) compute() {
	compute := func(cell *Cell, _ Dist, visited []*Cell) {
		wasVisited := func(cell *Cell) bool {
			for _, c := range visited {
				if c.Pos == cell.Pos {
					return true
				}
			}
			return false
		}
		g.paths[move(cell, cell)] = make(path, 0)
		for _, neighbour := range cell.neighbours {
			g.dists[move(cell, neighbour)] = Dist(1)
			g.dists[move(neighbour, cell)] = Dist(1)
			if _, ok := g.paths[move(cell, neighbour)]; !ok {
				g.paths[move(cell, neighbour)] = path{neighbour}
			}
			if _, ok := g.paths[move(neighbour, cell)]; !ok {
				g.paths[move(neighbour, cell)] = path{cell}
			}
			if wasVisited(neighbour) {
				for _, c := range visited {
					if c.Pos != cell.Pos && c.Pos != neighbour.Pos {
						nDist := g.dists[move(neighbour, c)] + 1
						g.dists[move(cell, c)] = nDist
						g.dists[move(c, cell)] = nDist
						if _, ok := g.paths[move(cell, c)]; !ok {
							if nPath, ok := g.paths[move(neighbour, c)]; ok {
								g.paths[move(cell, c)] = append(path{neighbour}, nPath...)
								g.paths[move(c, cell)] = append(g.paths[move(c, neighbour)], cell)
							}
						}
					}
				}
			}
		}
	}
	g.breadthFirstSearch(nil, compute)
	for _, cell := range g.cells {
		lastDist := Dist(-1)
		g.breadthFirstSearch(cell, func(c *Cell, dist Dist, visited []*Cell) {
			g.influences[speed1][cell.Pos] = influence{0: {cell}, 1: append([]*Cell{cell}, cell.neighbours...)}
			if lastDist == dist-1 {
				lastDist = dist
				t := turn(dist)
				if cell.Pos == xy(2, 2) {
					fmt.Println(visited)
				}
				g.influences[speed1][cell.Pos][t-1] = append(visited)
			}
		})
	}
}

// Dist _
type Dist int

// Pos _
type Pos int

func (p Pos) String() string {
	x, y := p.xy()
	return fmt.Sprintf("(%d,%d)", x, y)
}

func xy(x, y int) Pos {
	return Pos(y*G.width + x)
}

func (p Pos) sym() Pos {
	return xy(-(int(p)%G.width)+G.width-1, int(p)/G.width)
}

func (p Pos) xy() (int, int) {
	return int(p) % G.width, int(p) / G.width
}

// ToDirection _
func (p Pos) ToDirection(dir direction) Pos {
	x, y := p.xy()
	x, y = x+dir.x, y+dir.y
	if x < 0 {
		x += G.width
	} else if x == G.width {
		x = 0
	}
	if y < 0 {
		y += G.height
	} else if y == G.height {
		y = 0
	}
	return xy(x, y)
}

// Move _
type Move struct {
	from, to Pos
}

func move(from, to *Cell) Move {
	return Move{from.Pos, to.Pos}
}

// ScorePoint _
type ScorePoint int

// Pellets possible score point values
const (
	NormalPellet = ScorePoint(1)
	SuperPellet  = ScorePoint(10)
	Nought       = ScorePoint(0)
)

type freshness int

// Pellet _
type Pellet struct {
	Pos
	Value ScorePoint
}

// Rune _
func (pl Pellet) Rune() rune {
	switch pl.Value {
	case SuperPellet:
		return 'x'
		// return 0x2318
	default:
		return '·'
	}
}

// Cell _
type Cell struct {
	Pos
	neighbours []*Cell       // sorted by Direction order
	linkedWith map[Pos]*Cell // cells with same row or column without any wall in between
}

// G Game
var G *Game

type direction struct {
	x, y int
}

// Directions
var (
	up, down, right, left = direction{0, -1}, direction{0, 1}, direction{1, 0}, direction{-1, 0}
	Directions            = []direction{up, down, right, left} // order used when moving
)

// GameFromIoReader _
func GameFromIoReader(in io.Reader) *Game {
	G := new(Game)
	G.Scanner = bufio.NewScanner(in)
	G.Buffer(make([]byte, 1000000), 1000000)
	G.scanWidthAndHeight()
	return G
}

// PlayFirstTurn _
func (G *Game) PlayFirstTurn() {
	defer printElapsedTime("PlayFirstTurn")()
	G.ReadGameState()
	G.alliesCount = len(G.Allies())
	G.opponentCount = G.alliesCount
}

func main() {
	G = GameFromIoReader(os.Stdin)
	G.buildGraph()

	G.PlayFirstTurn()
	fmt.Println("MOVE 0 15 10")
	// os.Exit(0)
	for {
		G.ReadGameState()
		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println("MOVE 0 15 10") // MOVE <pacID> <x> <y>
		break
	}
	fmt.Println(G)
}
