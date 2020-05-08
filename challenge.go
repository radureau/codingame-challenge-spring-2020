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

/**
 * Grab the pellets as fast as you can!
 **/

var g *Game

func main() {
	g = new(Game).WithStdin(os.Stdin)
	for {
		debug(g)
		g.RefreshState()

		for _, pac := range g.myPacs {
			pac.Play(
				pac.ThinkAboutAMove(),
			)
			fmt.Println(pac)
		}
	}
}

func abs(v int) int {
	if v >= 0 {
		return v
	}
	return -v
}

func (p Pos) Dist(pos Pos) int {
	return abs(p.col-pos.col) + abs(p.row-pos.row)
}

func (pac *Pac) ThinkAboutAMove() Move {
	var maxPos Pos
	var maxWorth = -g.Size()
	for _, pt := range g.pellets {
		worth := int(pt.PelletValue)
		worth *= worth
		worth -= pac.Dist(pt.Pos)
		if abs(maxWorth-worth) <= 2 { // get the furthest from the closest opponent
			var minDist = g.Size()
			var minOpnt *Pac
			for _, opnt := range g.opponentPacs {
				dist := pac.Dist(opnt.Pos)
				if minDist > dist {
					minDist = dist
					minOpnt = opnt
				}
			}
			if minOpnt.Dist(pt.Pos) > minOpnt.Dist(maxPos) {
				maxWorth = worth
				maxPos = pt.Pos
			}
		} else if maxWorth < worth {
			maxWorth = worth
			maxPos = pt.Pos
		}
	}
	return Move{from: pac.Pos, to: maxPos}
}

func (pac *Pac) Play(mv Move) {
	pac.Move = mv
}

func (g *Game) Pacs() []*Pac {
	pacs := []*Pac{}
	for _, pac := range g.myPacs {
		pacs = append(pacs, pac)
	}
	return pacs
}

type Pac struct {
	ID   int  // pac number (unique within a team)
	mine bool // true if this pac is yours
	Pos
	typeID          int // unused in wood leagues
	speedTurnsLeft  int // unused in wood leagues
	abilityCooldown int // unused in wood leagues
	Move
}

func (g Game) String() string {
	s := ""
	for _, row := range g.Runes() {
		s += fmt.Sprintf("%s\n", string(row))
	}
	return s
}

func (g Game) Runes() [][]rune {
	runes := g.Grid.Runes()
	for _, pt := range g.pellets {
		runes[pt.row][pt.col] = pellet.Rune()
	}
	return runes
}

func (g Grid) String() string {
	s := ""
	for _, row := range g.Runes() {
		s += fmt.Sprintf("%s\n", string(row))
	}
	return s
}

func (g Grid) Runes() [][]rune {
	runes := make([][]rune, g.height)
	for i, row := range g.grid {
		runes[i] = make([]rune, g.width)
		for j, cell := range row {
			runes[i][j] = cell.Rune()
		}
	}
	return runes
}

func (p Pac) String() string {
	return fmt.Sprintf("MOVE %d %d %d", p.ID, p.Move.to.col, p.Move.to.row)
}

type Move struct {
	from Pos
	to   Pos
}

// code that I don't want to see anymore

func (g *Game) WithStdin(in io.Reader) *Game {
	g.Scanner = bufio.NewScanner(os.Stdin)
	g.Buffer(make([]byte, 1000000), 1000000)

	// width: size of the grid
	// height: top left corner is (x=0, y=0)
	g.Scan()
	fmt.Sscan(g.Text(), &g.width, &g.height)

	g.grid = make([][]Cell, g.height)
	for row := 0; row < g.height; row++ {
		g.grid[row] = make([]Cell, g.width)
		g.Scan()
		for col, c := range g.Text() { // one line of the grid: space " " is floor, pound "#" is wall
			g.grid[row][col] = Cell{Pos: Pos{row: row, col: col}, Type: celldecoder(c).Type()}
		}
	}
	return g
}

func (g *Game) RefreshState() {
	if g.turn > 0 {
		g.pastStates = append(g.pastStates, g.GameState)
	}
	g.turn++

	g.Scan()
	fmt.Sscan(g.Text(), &g.myScore, &g.opponentScore)

	g.Scan()
	fmt.Sscan(g.Text(), &g.visiblePacCount)

	g.pacs = make(map[Pos]*Pac, g.visiblePacCount)
	g.myPacs = make(map[Pos]*Pac, g.visiblePacCount)
	g.opponentPacs = make(map[Pos]*Pac, g.visiblePacCount)
	for i := 0; i < g.visiblePacCount; i++ {
		pac := new(Pac)
		var _mine int
		g.Scan()
		fmt.Sscan(g.Text(), &pac.ID, &_mine, &pac.row, &pac.col, &pac.typeID, &pac.speedTurnsLeft, &pac.abilityCooldown)
		pac.mine = _mine != 0
		g.pacs[pac.Pos] = pac
		if pac.mine {
			g.myPacs[pac.Pos] = pac
		} else {
			g.opponentPacs[pac.Pos] = pac
		}
	}

	g.Scan()
	fmt.Sscan(g.Text(), &g.visiblePelletCount)

	g.pellets = make(map[Pos]Pellet, g.visiblePelletCount)
	g.superPellets = make(map[Pos]Pellet, g.visiblePelletCount)
	for i := 0; i < g.visiblePelletCount; i++ {
		g.Scan()
		var pt Pellet
		fmt.Sscan(g.Text(), &pt.col, &pt.row, &pt.PelletValue)
		g.pellets[pt.Pos] = pt
		if pt.PelletValue == superPellet {
			g.superPellets[pt.Pos] = pt
		}
	}
}

type Grid struct {
	width  int
	height int
	grid   [][]Cell
}

func (g Grid) Size() int {
	return g.width * g.height
}

type Pos struct {
	row int
	col int
}

func (p Pos) ID() int {
	return p.row*g.width + p.col
}

func PosFromID(id int) Pos {
	p := Pos{row: id % g.width}
	p.col = id - p.row
	return p
}

type Cell struct {
	Type celltype
	Pos
}

func (c Cell) Rune() rune {
	switch c.Type {
	case wall:
		return '#'
	case ground:
		return ' '
	}
	panic(c.Type)
}

type celltype int

const (
	wall celltype = iota
	ground
)

type celldecoder rune

func (cd celldecoder) Type() celltype {
	switch cd {
	case '#':
		return wall
	case ' ':
		return ground
	}
	panic(cd)
}

type Pellet struct {
	PelletValue
	Pos
}

type PelletValue int

const (
	pellet      = PelletValue(1)
	superPellet = PelletValue(10)
)

func (p PelletValue) Rune() rune {
	switch p {
	case pellet:
		return '*'
	case superPellet:
		return 'X'
	}
	panic(p)
}

type Game struct {
	*bufio.Scanner
	Grid
	GameState
	pastStates []GameState
}

type GameState struct {
	turn               int
	myScore            int
	opponentScore      int
	visiblePacCount    int // all your pacs and enemy pacs in sight
	visiblePelletCount int // all pellets in sight
	pellets            map[Pos]Pellet
	superPellets       map[Pos]Pellet
	pacs               map[Pos]*Pac
	myPacs             map[Pos]*Pac
	opponentPacs       map[Pos]*Pac
}
