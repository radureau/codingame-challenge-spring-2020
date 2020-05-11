package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func debug(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

// Game _
type Game struct {
	*bufio.Scanner
	width  int // width: size of the grid
	height int // height: top left corner is (x=0, y=0)
	graph  *Graph
	*GameState
	pastStates   []*GameState // from latest to oldest
	scoreToReach ScorePoint
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
				if pl, ok := gs.pellets[0][cell.Pos]; ok {
					r = pl.Rune()
				} else if pac, ok := gs.pacs[0][cell.Pos]; ok {
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

// ReadGameState _
func (G *Game) ReadGameState() {
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
	} else {
	} // update freshness

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
		G.scoreToReach = (ScorePoint(len(G.graph.cells)-len(G.pacs[0])-nSuperPellet)*NormalPellet+
			ScorePoint(nSuperPellet)*SuperPellet)/
			2 + 1
	} else {
	} // update freshness
}

func (G *Game) scanWidthAndHeight() {
	G.Scan()
	fmt.Sscan(G.Text(), &G.width, &G.height)
}
func (G *Game) buildGraph() {
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
	G.graph.computeDistances()
}

// Graph _
type Graph struct {
	cells     map[Pos]*Cell
	positions []Pos // sorted
	dists     map[Move]Dist
}

// NewGraph _
func NewGraph(capacity int) *Graph {
	g := new(Graph)
	g.cells = make(map[Pos]*Cell, capacity)
	g.positions = make([]Pos, 0, capacity)
	g.dists = map[Move]Dist{}
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
func (g Graph) writeDistance(c1, c2 *Cell, dist Dist) {
	if _, ok := g.dists[move(c1, c2)]; !ok {
		g.dists[move(c1, c2)] = dist
		g.dists[move(c2, c1)] = dist
	}
}
func (g Graph) breadthFirstSearch(compute func(cell *Cell, visited map[Pos]*Cell)) {
	visited := make(map[Pos]*Cell, len(g.cells))
	markAsVisited := func(c *Cell) {
		visited[c.Pos] = c
	}
	wasVisited := func(c *Cell) bool {
		_, ok := visited[c.Pos]
		return ok
	}
	start := g.cells[Pos(g.positions[0])]
	markAsVisited(start)
	toVisit := []*Cell{start}
	for len(toVisit) > 0 {
		var current *Cell
		current, toVisit = toVisit[0], toVisit[1:]
		compute(current, visited)
		for _, c := range current.neighbours {
			if !wasVisited(c) {
				markAsVisited(c)
				toVisit = append(toVisit, c)
			}
		}
	}
}
func (g Graph) computeDistances() {
	compute := func(cell *Cell, visited map[Pos]*Cell) {
		wasVisited := func(c *Cell) bool {
			_, ok := visited[c.Pos]
			return ok
		}
		for _, neighbour := range cell.neighbours {
			g.writeDistance(cell, neighbour, Dist(1))
			if wasVisited(neighbour) {
				for _, c := range visited {
					if cell.Pos != c.Pos {
						g.writeDistance(cell, c, g.dists[move(neighbour, c)]+1)
					}
				}
			}
		}
	}
	g.breadthFirstSearch(compute)
	// i := 0
	// g.breadthFirstSearch(func(cell *Cell, _ map[Pos]*Cell) {
	// 	i++
	// 	debug(i, cell.Pos)
	// })
	// debug(len(g.cells))
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
	G.ReadGameState()
}

func main() {
	G = GameFromIoReader(os.Stdin)
	G.buildGraph()

	G.PlayFirstTurn()
	// os.Exit(0)
	for {
		G.ReadGameState()
		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println("MOVE 0 15 10") // MOVE <pacID> <x> <y>
		break
	}
	fmt.Println(G)
}
