package main

import (
	"fmt"
	"math/rand"
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
			assert.Equal(t, tC.expected, tC.from.ToDirection(tC.direction))
		})
	}
}
