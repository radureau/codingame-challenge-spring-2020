package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// G Game
var G *Game

func main() {
	G = GameFromIoReader(os.Stdin)
	G.buildGraph()

	G.PlayFirstTurn()
	if os.Getenv("USER") == "gabrielradureau" {
		fmt.Println(G.GameState)
	}
	for !G.IsOver() {
		G.Play()
		if os.Getenv("USER") == "gabrielradureau" {
			fmt.Println(G.GameState)
		}
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
		for pos := range G.graph.nodes {
			if _, ok := G.pellets[0][pos]; !ok {
				if _, ok := G.pacs[0][pos]; !ok {
					G.pellets[0][pos] = &Pellet{Pos: pos, Value: NormalPellet}
				}
			}
		}
		G.oldestPelletFresness = 0
		G.scoreToReach = (ScorePoint(len(G.graph.nodes)-len(G.pacs[0])-nSuperPellet)*NormalPellet+
			ScorePoint(nSuperPellet)*SuperPellet)/
			2 + 1
		G.alliesCount = len(G.Allies())
		G.opponentCount = G.alliesCount
	} else {
		G.oldestPelletFresness = trackPelletFreshness(G.pellets, G.before.pellets)
		// evict consumed pellets
		for _, pac := range G.Allies() {
			node := G.graph.nodes[pac.Pos]
			untrackPelletAt(node.Pos)
			for pos := range node.linkedWith {
				if plt, ok := G.pellets[0][pos]; ok && plt.Value == Nought {
					untrackPelletAt(pos)
				}
			}
		}
	}
}

// PlayFirstTurn _
func (G *Game) PlayFirstTurn() {
	defer printElapsedTime("PlayFirstTurn")()
	G.Play()
}

// Play _
func (G *Game) Play() {
	G.ReadGameState()
	myPacs := G.Allies()
	moves := make([]string, len(myPacs))
	for i, ally := range myPacs {
		moves[i] = fmt.Sprintf("%s %d %v",
			"MOVE",
			ally.ID,
			"15 10",
		) // MOVE <pacID> <x> <y>
	}
	fmt.Println(strings.Join(moves, "|"))
}
