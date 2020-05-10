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
	pastStates []*GameState // from latest to oldest
}

func (G *Game) String() string {
	s := ""
	for y := 0; y < G.height; y++ {
		runes := make([]rune, G.width)
		for x := 0; x < G.width; x++ {
			if cell, ok := G.graph.cells[xy(x, y)]; ok {
				r := ' '
				_ = cell
				runes[x] = r
			} else {
				runes[x] = '#'
			}
		}
		s += string(runes) + "\n"
	}
	return s
}

// GameState _
type GameState struct {
	*Game
	turn
	myScore            ScorePoint
	opponentScore      ScorePoint
	visiblePacCount    int // all your pacs and enemy pacs in sight
	visiblePelletCount int // all pellets in sight
}

func (gs *GameState) String() string {
	return fmt.Sprintf("GameState#%d", gs.turn)
}

type turn int

// MaxTurn _
const MaxTurn = turn(200)

// ReadGameState _
func (G *Game) ReadGameState() {
	if G.GameState == nil {
		G.GameState = &GameState{Game: G, turn: 1}
		G.pastStates = make([]*GameState, 0, MaxTurn-1)
	} else {
		G.pastStates = append(G.pastStates, G.GameState)
		G.GameState = &GameState{Game: G, turn: G.turn + 1}
	}
	G.Scan()
	fmt.Sscan(G.Text(), &G.myScore, &G.opponentScore)

	G.Scan()
	fmt.Sscan(G.Text(), &G.visiblePacCount)

	for i := 0; i < G.visiblePacCount; i++ {
		// pacID: pac number (unique within a team)
		// mine: true if this pac is yours
		// x: position in the grid
		// y: position in the grid
		// typeID: unused in wood leagues
		// speedTurnsLeft: unused in wood leagues
		// abilityCooldown: unused in wood leagues
		var pacID int
		var mine bool
		var _mine int
		var x, y int
		var typeID string
		var speedTurnsLeft, abilityCooldown int
		G.Scan()
		fmt.Sscan(G.Text(), &pacID, &_mine, &x, &y, &typeID, &speedTurnsLeft, &abilityCooldown)
		mine = _mine != 0
		_ = mine
	}

	G.Scan()
	fmt.Sscan(G.Text(), &G.visiblePelletCount)

	for i := 0; i < G.visiblePelletCount; i++ {
		// value: amount of points this pellet is worth
		var x, y, value int
		G.Scan()
		fmt.Sscan(G.Text(), &x, &y, &value)
	}
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
	value ScorePoint
	freshness
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

func (G *Game) PlayFirstTurn() {
	G.ReadGameState()
}

func main() {
	G = GameFromIoReader(os.Stdin)
	G.buildGraph()

	G.PlayFirstTurn()
	for {
		break
		G.ReadGameState()
		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println("MOVE 0 15 10") // MOVE <pacID> <x> <y>
	}
	fmt.Println(G)
}
