package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Grid struct {
	width  int
	height int
	grid   [][]Cell
}

type Pos struct {
	row int
	col int
}

type Cell struct {
	Type celltype
	Pos
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

type Pellet int

type Pac struct {
	ID   int  // pac number (unique within a team)
	mine bool // true if this pac is yours
	Pos
	typeID          int // unused in wood leagues
	speedTurnsLeft  int // unused in wood leagues
	abilityCooldown int // unused in wood leagues
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
	pacs               map[Pos]*Pac
	myPacs             map[Pos]*Pac
	opponentPacs       map[Pos]*Pac
}

/**
 * Grab the pellets as fast as you can!
 **/

func main() {
	g := new(Game).WithStdin(os.Stdin)
	for {
		g.RefreshState()

		// fmt.Fprintln(os.Stderr, "Debug messages...")
		fmt.Println("MOVE 0 15 10") // MOVE <pacId> <x> <y>
	}
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
	for i := 0; i < g.visiblePelletCount; i++ {
		g.Scan()
		var x, y int
		var value Pellet
		fmt.Sscan(g.Text(), &x, &y, &value)
		g.pellets[Pos{col: x, row: y}] = value
	}
}
