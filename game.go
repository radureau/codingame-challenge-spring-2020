package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

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

func (G *Game) scanWidthAndHeight() {
	G.Scan()
	fmt.Sscan(G.Text(), &G.width, &G.height)
}

// GameFromIoReader _
func GameFromIoReader(in io.Reader) *Game {
	G := new(Game)
	G.Scanner = bufio.NewScanner(in)
	G.Buffer(make([]byte, 1000000), 1000000)
	G.scanWidthAndHeight()
	return G
}

func (G *Game) buildGraph() {
	defer printElapsedTime("buildGraph")()
	G.graph = NewGraph(G.width * G.height)
	for y := 0; y < G.height; y++ {
		G.Scan()
		const floor = ' '
		for x, r := range G.Text() { // one line of the grid: space " " is floor, pound "#" is wall
			if r == floor {
				G.graph.createNode(x, y)
			}
		}
	}
	G.graph.linkTogether()
	G.graph.compute()
}

// G Game
var G *Game

// PlayFirstTurn _
func (G *Game) PlayFirstTurn() {
	defer printElapsedTime("PlayFirstTurn")()
	G.ReadGameState()
	G.alliesCount = len(G.Allies())
	G.opponentCount = G.alliesCount
	G.Play()
}

// Play _
func (G *Game) Play() {
	moves := make([]string, 0, 5)
	for _, ally := range G.Allies() {
		moves = append(moves, fmt.Sprintf("%s %d %v",
			"MOVE",
			ally.ID,
			"15 10",
		)) // MOVE <pacID> <x> <y>
	}
	fmt.Println(strings.Join(moves, "|"))
}

func main() {
	G = GameFromIoReader(os.Stdin)
	G.buildGraph()

	G.PlayFirstTurn()
	if os.Getenv("USER") == "gabrielradureau" {
		fmt.Println(G.GameState)
	}
	for {
		G.ReadGameState()
		G.Play()
		if os.Getenv("USER") == "gabrielradureau" {
			fmt.Println(G.GameState)
		}
	}
}
