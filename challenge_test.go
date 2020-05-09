package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func _fakeFame() *Game {
	return &Game{Grid: Grid{height: 100, width: 100}}
}

func TestSortedPos_SortByID(t *testing.T) {
	g = _fakeFame()
	positions := sortedPos{
		{2, 3},
		{7, 1},
		{1, 2},
		{3, 2},
	}

	positions.SortByID()

	if *positions[0] != (Pos{1, 2}) {
		fmt.Println(positions)
		t.Fail()
	}

}

func ExamplePelletMapper_String() {
	g = _fakeFame()
	pellets := []Pellet{
		{Pos: Pos{2, 3}},
		{Pos: Pos{7, 1}},
		{Pos: Pos{1, 2}},
		{Pos: Pos{3, 2}},
	}
	pm := PelletMapper{}
	pm.Init()
	for _, pt := range pellets {
		pm.Add(pt)
	}
	pm.SortByID()
	fmt.Println(pm)

	fmt.Println()

	for i := range pm.sortedPos {
		pt := pm.get(i)
		fmt.Printf("%v ", pt)
	}

	//Output:
	// [{102 0} {203 0} {302 0} {701 0}]
	//
	// {0 {1 2}} {0 {2 3}} {0 {3 2}} {0 {7 1}}
}

func ExamplePelletMapper_Sort() {
	g = _fakeFame()
	pellets := []Pellet{
		{Pos: Pos{2, 3}, PelletValue: pellet},
		{Pos: Pos{7, 1}, PelletValue: superPellet},
		{Pos: Pos{1, 2}, PelletValue: superPellet},
		{Pos: Pos{3, 2}, PelletValue: pellet},
	}
	pm := PelletMapper{}
	pm.Init()
	for _, pt := range pellets {
		pm.Add(pt)
	}
	pm.SortByID()
	fmt.Println(pm)

	sorted := pm.Ordered(ByDescPelletValue{}).Sort()

	for i := range sorted.sortedPos {
		pt := sorted.get(i)
		fmt.Printf("%v", pt)
	}
	fmt.Println()

	sorted = pm.Ordered(
		ByDescPelletValue{},
		ByRelativeDistTo(Pos{100, 100}),
	).Sort()

	for i := range sorted.sortedPos {
		pt := sorted.get(i)
		_ = pt
		fmt.Printf("%v", pt)
	}
	fmt.Println()

	dot := Pos{5, 5}
	sorted = pm.Ordered(
		ByRelativeDistTo(dot),
		ByPosID{},
	).Sort()
	for i := range sorted.sortedPos {
		pt := sorted.get(i)
		_ = pt
		fmt.Printf("(%d %v)", pt.Dist(dot), pt.Pos)
	}
	fmt.Println()

	dot = Pos{3, 4}
	sorted = pm.Ordered(
		ByDescRelativeDistTo(dot),
		ByPosID{},
	).Sort()
	for i := range sorted.sortedPos {
		pt := sorted.get(i)
		_ = pt
		fmt.Printf("(%d %v)", pt.Dist(dot), pt.Pos)
	}

	// Output:
	// [{102 10} {203 1} {302 1} {701 10}]
	// {10 {1 2}}{10 {7 1}}{1 {2 3}}{1 {3 2}}
	// {10 {7 1}}{10 {1 2}}{1 {2 3}}{1 {3 2}}
	// (5 {2 3})(5 {3 2})(6 {7 1})(7 {1 2})
	// (7 {7 1})(4 {1 2})(2 {2 3})(2 {3 2})
}

func Test_OrderByPos(t *testing.T) {
	g = new(Game)
	g.width = 4
	g.height = 4
	pellets := make([]Pellet, g.Size())
	for i := range pellets {
		pellets[i] = Pellet{Pos: PosFromID(i)}
	}

	pm := new(PelletMapper).Init()
	for _, pt := range pellets {
		pm.Add(pt)
	}

	pm = pm.Ordered(
		ByRelativeDistTo(g.Center()),
		ByPosID{},
	).Sort()
	expected := []*Pos{
		{2, 2},
		{1, 2}, {2, 1}, {2, 3}, {3, 2},
		{0, 2}, {1, 1}, {1, 3}, {2, 0}, {3, 1}, {3, 3},
		{0, 1}, {0, 3}, {1, 0}, {3, 0},
		{0, 0},
	}
	// lastDist := 0
	// for i := range pm.sortedPos {
	// 	pt := pm.get(i)
	// 	dist := pt.Dist(g.Center())
	// 	if lastDist != dist {
	// 		fmt.Println()
	// 	}
	// 	lastDist = dist
	// 	fmt.Printf("(%d %v)", dist, pt.Pos)
	// }
	// fmt.Println()
	assert.EqualValues(t, expected, []*Pos(pm.sortedPos))

	// ----------------

	pm = pm.Ordered(
		ByDescRelativeDistTo(g.Center()),
		ByPosID{},
	).Sort()
	expected = []*Pos{
		{0, 0},
		{0, 1}, {0, 3}, {1, 0}, {3, 0},
		{0, 2}, {1, 1}, {1, 3}, {2, 0}, {3, 1}, {3, 3},
		{1, 2}, {2, 1}, {2, 3}, {3, 2},
		{2, 2},
	}
	// lastDist = 0
	// for i := range pm.sortedPos {
	// 	pt := pm.get(i)
	// 	dist := pt.Dist(g.Center())
	// 	if lastDist != dist {
	// 		fmt.Println()
	// 	}
	// 	lastDist = dist
	// 	fmt.Printf("(%d %v)", dist, pt.Pos)
	// }
	// fmt.Println()
	assert.EqualValues(t, expected, []*Pos(pm.sortedPos))
}
