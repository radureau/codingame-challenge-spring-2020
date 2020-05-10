package main

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
}
