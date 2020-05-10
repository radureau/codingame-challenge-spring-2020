package main

import (
	"bufio"
	"fmt"
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
}

func scanWidthAndHeight() {
	G.Scan()
	fmt.Sscan(G.Text(), &G.width, &G.height)
}
func buildGraph() {
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
		}
	}
}
func (g Graph) writeDistance(c1, c2 *Cell, dist Dist) {
	g.dists[move(c1, c2)] = dist
	g.dists[move(c2, c1)] = dist
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
					g.writeDistance(cell, c, g.dists[move(neighbour, c)]+1)
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

// Cell _
type Cell struct {
	Pos
	neighbours []*Cell // sorted by Direction order
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

func main() {
	G = new(Game)
	G.Scanner = bufio.NewScanner(os.Stdin)
	G.Buffer(make([]byte, 1000000), 1000000)

	scanWidthAndHeight()

	buildGraph()
	for {
		var myScore, opponentScore int
		G.Scan()
		fmt.Sscan(G.Text(), &myScore, &opponentScore)
		// visiblePacCount: all your pacs and enemy pacs in sight
		var visiblePacCount int
		G.Scan()
		fmt.Sscan(G.Text(), &visiblePacCount)

		for i := 0; i < visiblePacCount; i++ {
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
		// visiblePelletCount: all pellets in sight
		var visiblePelletCount int
		G.Scan()
		fmt.Sscan(G.Text(), &visiblePelletCount)

		for i := 0; i < visiblePelletCount; i++ {
			// value: amount of points this pellet is worth
			var x, y, value int
			G.Scan()
			fmt.Sscan(G.Text(), &x, &y, &value)
		}

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println("MOVE 0 15 10") // MOVE <pacID> <x> <y>
		break
	}
}
