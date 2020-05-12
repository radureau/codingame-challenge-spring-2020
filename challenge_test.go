package main

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPos(t *testing.T) {
	G = new(Game)
	G.height = 3
	G.width = 5

	testCases := []struct {
		Pos
		sym Pos
	}{
		{xy(1, 0), xy(3, 0)},
		{xy(2, 1), xy(2, 1)},
		{xy(4, 2), xy(0, 2)},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%v | %v", tC.Pos, tC.sym), func(t *testing.T) {
			assert.Equal(t, tC.sym, tC.Pos.sym())
		})
	}
}

func TestDirections(t *testing.T) {
	G = new(Game)
	G.height = rand.Intn(3) + 3
	G.width = rand.Intn(3) + 3

	repr := []string{"UP", "DOWN", "RIGHT", "LEFT"}

	testCases := []struct {
		from Pos
		direction
		expected Pos
	}{
		{xy(1, 1), up, xy(1, 0)},
		{xy(1, 1), down, xy(1, 2)},
		{xy(1, 1), right, xy(2, 1)},
		{xy(1, 1), left, xy(0, 1)},

		{xy(1, 0), up, xy(1, G.height-1)},
		{xy(1, G.height-1), down, xy(1, 0)},
		{xy(G.width-1, 1), right, xy(0, 1)},
		{xy(0, 1), left, xy(G.width-1, 1)},
	}
	for i, tC := range testCases {
		t.Run(fmt.Sprint(i, "\t", tC.from, ".", repr[i%4]), func(t *testing.T) {
			actual := tC.from.ToDirection(tC.direction)
			if tC.expected != actual {
				t.Fatalf("expected: %v\tactual: %v", tC.expected, actual)
			}
		})
	}
}

func TestGraph(t *testing.T) {
	input, err := os.Open("simple.txt")
	assert.NoError(t, err)
	G = GameFromIoReader(input)
	G.buildGraph()

	testCases := []struct {
		Move
		Dist
	}{
		{Move{xy(0, 1), xy(0, 1)}, Dist(0)},
		{Move{xy(0, 1), xy(1, 1)}, Dist(1)},
		{Move{xy(0, 1), xy(4, 1)}, Dist(1)},
		{Move{xy(0, 1), xy(1, 2)}, Dist(2)},
		{Move{xy(0, 1), xy(3, 1)}, Dist(2)},
		{Move{xy(1, 1), xy(0, 1)}, Dist(1)},
		{Move{xy(1, 1), xy(4, 1)}, Dist(2)},
		{Move{xy(3, 2), xy(0, 1)}, Dist(3)},
		{Move{xy(3, 2), xy(3, 2)}, Dist(0)},
	}
	for i, tC := range testCases {
		t.Run(fmt.Sprint(i, tC), func(t *testing.T) {
			assert.Equal(t, tC.Dist, G.graph.dists[tC.Move])
		})
	}

	testCases2 := []struct {
		Pos
		nLinkedWith int
	}{
		{xy(2, 2), 2},
		{xy(1, 4), 6},
		{xy(1, 2), 5},
	}
	for i, tC := range testCases2 {
		t.Run(fmt.Sprint(i, tC), func(t *testing.T) {
			assert.Equal(t, tC.nLinkedWith, len(G.graph.cells[tC.Pos].linkedWith))
		})
	}

	C := G.graph.cells

	testCases3 := []struct {
		Pos
		speed
		expected influence
	}{
		{
			Pos: xy(2, 2), speed: speed1,
			// #####
			// 95#37
			// #201#
			// #6#4#
			// ²O#8¹
			// #####
			expected: influence{
				turn(0): {C[xy(2, 2)]}, // 0
				turn(1): {C[xy(2, 2)],
					C[xy(3, 2)], C[xy(1, 2)]}, // 1 2
				turn(2): {C[xy(2, 2)],
					C[xy(3, 2)], C[xy(1, 2)],
					C[xy(3, 1)], C[xy(3, 3)], C[xy(1, 1)], C[xy(1, 3)]}, // 3 4 5 6
				turn(3): {C[xy(3, 3)],
					C[xy(3, 2)], C[xy(1, 2)],
					C[xy(3, 1)], C[xy(3, 3)], C[xy(1, 1)], C[xy(1, 3)],
					C[xy(4, 1)], C[xy(3, 4)], C[xy(0, 1)], C[xy(1, 4)]}, // 7 8 9 O
			},
		},
		{
			Pos: xy(2, 2), speed: speed2,
			expected: influence{
				turn(0): {C[xy(2, 2)]}, // 0
				turn(1): {C[xy(2, 2)],
					C[xy(3, 2)], C[xy(1, 2)], // 1 2
					C[xy(3, 1)], C[xy(3, 3)], C[xy(1, 1)], C[xy(1, 3)]}, // 3 4 5 6
				turn(2): {C[xy(3, 3)],
					C[xy(3, 2)], C[xy(1, 2)],
					C[xy(3, 1)], C[xy(3, 3)], C[xy(1, 1)], C[xy(1, 3)],
					C[xy(4, 1)], C[xy(3, 4)], C[xy(0, 1)], C[xy(1, 4)], // 7 8 9 O
					C[xy(4, 4)], C[xy(0, 4)]}, // ¹ ²
			},
		},
	}
	for i, tC := range testCases3 {
		t.Run(fmt.Sprintf("%d:\tinfluence from %v with speed %d", i, tC.Pos, tC.speed), func(t *testing.T) {
			assert.Equal(t, tC.expected, G.graph.influences[tC.speed][tC.Pos])
		})
	}
}

func TestTrackPacFreshness(t *testing.T) {
	G = new(Game)
	G.height = 5
	G.width = 5

	testCases := []struct {
		current, before, expected map[freshness]map[Pos]*Pac
		expectedOldestFreshness   freshness
	}{
		{
			current: map[freshness]map[Pos]*Pac{
				0: {
					xy(0, 0): &Pac{PacID: PacID{ID: 1, ally: true}, Pos: xy(0, 0)},
					xy(0, 1): &Pac{PacID: PacID{ID: 1, ally: false}, Pos: xy(0, 1)},
				},
			},
			before: map[freshness]map[Pos]*Pac{
				0: {
					xy(0, 1): &Pac{PacID: PacID{ID: 1, ally: true}, Pos: xy(0, 1)},
				},
				1: {
					xy(1, 1): &Pac{PacID: PacID{ID: 1, ally: false}, Pos: xy(1, 1)},
				},
				2: {
					xy(4, 3): &Pac{PacID: PacID{ID: 2, ally: false}, Pos: xy(4, 3)},
				},
			},
			expected: map[freshness]map[Pos]*Pac{
				0: {
					xy(0, 0): &Pac{PacID: PacID{ID: 1, ally: true}, Pos: xy(0, 0)},
					xy(0, 1): &Pac{PacID: PacID{ID: 1, ally: false}, Pos: xy(0, 1)},
				},
				3: {
					xy(4, 3): &Pac{PacID: PacID{ID: 2, ally: false}, Pos: xy(4, 3)},
				},
			},
			expectedOldestFreshness: 3,
		},
		{
			current: map[freshness]map[Pos]*Pac{
				0: {
					xy(0, 0): &Pac{PacID: PacID{ID: 1, ally: true}, Pos: xy(0, 0)},
				},
			},
			before: map[freshness]map[Pos]*Pac{
				0: {
					xy(0, 0): &Pac{PacID: PacID{ID: 1, ally: true}, Pos: xy(0, 0)},
					xy(0, 1): &Pac{PacID: PacID{ID: 2, ally: true}, Pos: xy(0, 1)},
				},
			},
			expected: map[freshness]map[Pos]*Pac{
				0: {
					xy(0, 0): &Pac{PacID: PacID{ID: 1, ally: true}, Pos: xy(0, 0)},
				},
			},
			expectedOldestFreshness: 0,
		},
	}
	for i, tC := range testCases {
		t.Run(string(i), func(t *testing.T) {
			start := time.Now()
			oldest := trackPacFreshness(tC.current, tC.before)
			elapsed := time.Since(start)
			fmt.Printf("trackPacFreshness took %s\n", elapsed)
			assert.Equal(t, tC.expectedOldestFreshness, oldest)
			assert.Equal(t, tC.expected, tC.current)
		})
	}
}
