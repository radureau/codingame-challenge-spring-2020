package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
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
		g.RefreshState()

		outputs := make([]string, len(g.myPacs))
		for _, pac := range g.myPacs {
			outputs[pac.ID] = pac.Play(
				pac.ThinkAboutAMove(),
			)
		}
		fmt.Println(strings.Join(outputs, "|"))
	}
}

func abs(v int) int {
	if v >= 0 {
		return v
	}
	return -v
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func (p Pos) Dist(other Pos) int {
	return abs(p.col-other.col) + abs(p.row-other.row)
}

func (pac *Pac) LastMove() *Move {
	if pac.lastMove != nil {
		return pac.lastMove
	}
	if g.turn <= 1 {
		return nil
	}
	lastState := g.pastStates[0]
	for _, p := range lastState.myPacs {
		if pac.ID == p.ID {
			pac.lastMove = &p.Move
			break
		}
	}
	return pac.lastMove
}

// type Direction Pos

// var (
// 	Up    = Direction{-1, 0}
// 	Down  = Direction{1, 0}
// 	Right = Direction{0, 1}
// 	Left  = Direction{0, -1}
// 	Stay  = Direction{}
// )

func (mv Move) Score() float64 {
	worth := float64(g.pelletsMap.Get(mv.to).PelletValue)
	worth *= worth * worth
	dist := mv.from.Dist(mv.to)
	worth -= float64(dist * dist)
	return worth
}

func (pac *Pac) ThinkAboutAMove() Move {
	// if g.turn > 1 {
	// 	mv := pac.LastMove()
	// 	if mv.from != pac.Pos && mv.Score() > 0 {
	// 		return *mv
	// 	}
	// }
	var closestOpnt, closestAlly *Pac
	for _, opnt := range g.opponentPacs {
		closestOpnt = opnt
	}
	// for _, ally := range g.myPacs {
	minDistAlly := g.Size()
	for _, ally := range g.pacs {
		if ally.ID == pac.ID {
			continue
		}
		if dist := ally.Dist(pac.Pos); minDistAlly > dist {
			closestAlly = ally
			minDistAlly = dist
		}
	}
	_, _ = closestAlly, closestOpnt
	// dot := Pos{
	// 	row: g.width / len(g.myPacs) * pac.ID,
	// 	col: g.height / len(g.myPacs) * pac.ID,
	// }
	sorted := g.pelletsMap.Ordered(
		// ByDescRelativePelletValueTo(pac.Pos),
		ByDescPelletValue{},
		// ByRelativeDistTo(pac.Pos),
		// ByDescRelativeDistTo(closestOpnt.Pos),
		ByDescRelativeDistTo(closestAlly.Pos),
		// ByRelativeDistTo(dot),
		ByPosID{},
	).Sort()
	// for i := range sorted.sortedPos[:min(5, sorted.Len())] {
	// 	pt := sorted.get(i)
	// 	debug(fmt.Sprintf("%v", pt))
	// }
	// debug()
	mv := Move{from: pac.Pos}
SynchWitlAllies:
	for _, to := range sorted.sortedPos {
		mv.to = *to
		for _, ally := range g.myPacs {
			if ally.ID == pac.ID ||
				ally.Move.to == *to ||
				ally.LastMove() != nil && ally.LastMove().to == *to {
				continue SynchWitlAllies
			}
			return mv
		}
	}
	return mv
}

func (pac *Pac) Play(mv Move) string {
	pac.Move = mv
	return pac.String()
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
	lastMove *Move
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
	for _, pt := range g.pelletsMap.pellets {
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
	return fmt.Sprintf("MOVE %d %d %d %v", p.ID, p.Move.to.col, p.Move.to.row, p.Pos)
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
	lastGameState := g.GameState
	if lastGameState.turn > 0 {
		g.pastStates = append([]*GameState{&lastGameState}, g.pastStates...)
	}
	g.GameState = GameState{turn: lastGameState.turn + 1}

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

	g.pelletsMap.Init()
	for i := 0; i < g.visiblePelletCount; i++ {
		g.Scan()
		var pt Pellet
		fmt.Sscan(g.Text(), &pt.col, &pt.row, &pt.PelletValue)
		g.pelletsMap.Add(pt)
	}
	g.pelletsMap.SortByID()
}

type Grid struct {
	width  int
	height int
	grid   [][]Cell
}

func (g Grid) Size() int {
	return g.width * g.height
}

func (g Grid) Center() Pos {
	return Pos{col: g.width / 2, row: g.height / 2}
}

type Pos struct {
	row int
	col int
}

func (p *Pos) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p Pos) ID() int {
	return p.row*g.width + p.col
}

func PosFromID(id int) Pos {
	return Pos{row: id % g.width, col: id / g.width}
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
	pastStates []*GameState // from latest to oldest
}

type GameState struct {
	turn               int
	myScore            int
	opponentScore      int
	visiblePacCount    int // all your pacs and enemy pacs in sight
	visiblePelletCount int // all pellets in sight
	pelletsMap         PelletMapper
	pacs               map[Pos]*Pac
	myPacs             map[Pos]*Pac
	opponentPacs       map[Pos]*Pac
}

func (gs GameState) String() string {
	return fmt.Sprintf("GameState#%d", gs.turn)
}

type sortedPos []*Pos

func (positions sortedPos) SortByID() {
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].ID() < positions[j].ID()
	})
}

type PacMapper struct {
	pacs map[Pos]*Pac
	sortedPos
	sorts []PacComparator
}

func (pm PacMapper) Ordered(comparators ...PacComparator) *PacMapper {
	npm := pm.Copy()
	npm.sorts = comparators
	return npm
}

func (pm PacMapper) Sort() *PacMapper {
	sort.Stable(pm)
	return &pm
}

func (pm PacMapper) Less(i, j int) bool {
	pI, pJ := pm.get(i), pm.get(j)
	for _, cmp := range pm.sorts { // if i == j then use next comparator
		if cmp.Less(pI, pJ) {
			return true
		} else if cmp.Less(pJ, pI) {
			return false
		}
	}
	return false
}

func (pm *PacMapper) Init() *PacMapper {
	pm.pacs = make(map[Pos]*Pac)
	return pm
}
func (pm *PacMapper) Add(p *Pac) {
	pm.pacs[p.Pos] = p
	pm.sortedPos = append(pm.sortedPos, &p.Pos)
}

func (pm PacMapper) Get(pos Pos) *Pac {
	return pm.pacs[pos]
}

func (pm PacMapper) get(i int) *Pac {
	return pm.pacs[*pm.sortedPos[i]]
}

func (pm PacMapper) Copy() *PacMapper {
	sp := make(sortedPos, pm.Len())
	copy(sp, pm.sortedPos)
	return &PacMapper{
		pacs:      pm.pacs,
		sortedPos: sp,
	}
}

func (pm PacMapper) Len() int {
	return len(pm.pacs)
}

func (pm PacMapper) Swap(i, j int) {
	pm.sortedPos[i], pm.sortedPos[j] = pm.sortedPos[j], pm.sortedPos[i]
}

type PelletMapper struct {
	pellets map[Pos]Pellet
	sortedPos
	sorts []PelletComparator
}

func (pm PelletMapper) Ordered(comparators ...PelletComparator) *PelletMapper {
	npm := pm.Copy()
	npm.sorts = comparators
	return npm
}

func (pm PelletMapper) String() string {
	type v struct {
		ID int
		PelletValue
	}
	s := make([]v, len(pm.sortedPos))
	for i, pos := range pm.sortedPos {
		pt := pm.pellets[*pos]
		s[i] = v{pt.ID(), pt.PelletValue}
	}
	return fmt.Sprintf("%v", s[:min(5, len(pm.sortedPos))])
}

func (pm PelletMapper) Sort() *PelletMapper {
	sort.Stable(pm)
	return &pm
}

func (pm PelletMapper) Less(i, j int) bool {
	ptI, ptJ := pm.get(i), pm.get(j)
	for _, cmp := range pm.sorts { // if i == j then use next comparator
		if cmp.Less(&ptI, &ptJ) {
			return true
		} else if cmp.Less(&ptJ, &ptI) {
			return false
		}
	}
	return false
}

func (pm *PelletMapper) Init() *PelletMapper {
	pm.pellets = make(map[Pos]Pellet, g.visiblePelletCount)
	return pm
}
func (pm *PelletMapper) Add(pt Pellet) {
	pm.pellets[pt.Pos] = pt
	pm.sortedPos = append(pm.sortedPos, &pt.Pos)
}

func (pm PelletMapper) Get(pos Pos) Pellet {
	return pm.pellets[pos]
}

func (pm PelletMapper) get(i int) Pellet {
	return pm.pellets[*pm.sortedPos[i]]
}

func (pm PelletMapper) Copy() *PelletMapper {
	sp := make(sortedPos, pm.Len())
	copy(sp, pm.sortedPos)
	return &PelletMapper{
		pellets:   pm.pellets,
		sortedPos: sp,
	}
}

func (pm PelletMapper) Len() int {
	return len(pm.pellets)
}

func (pm PelletMapper) Swap(i, j int) {
	pm.sortedPos[i], pm.sortedPos[j] = pm.sortedPos[j], pm.sortedPos[i]
}

// ORDERING

type PacComparator interface {
	Less(pI, pJ *Pac) bool
}

type ByPacID struct{}

func (pm ByPacID) Less(pI, pJ *Pac) bool {
	return pI.ID < pJ.ID
}

// _

type PelletComparator interface {
	Less(ptI, ptJ *Pellet) bool
}

type ByDescPelletValue struct{}

func (pm ByDescPelletValue) Less(ptI, ptJ *Pellet) bool {
	return ptI.PelletValue > ptJ.PelletValue
}

type byRelativeDistTo struct {
	Pos
}

func ByRelativeDistTo(pos Pos) PelletComparator {
	return byRelativeDistTo{pos}
}

func (pm byRelativeDistTo) Less(ptI, ptJ *Pellet) bool {
	return ptI.Dist(pm.Pos) < ptJ.Dist(pm.Pos)
}

type byDescRelativeDistTo struct {
	Pos
}

func ByDescRelativeDistTo(pos Pos) PelletComparator {
	return byDescRelativeDistTo{pos}
}

func (pm byDescRelativeDistTo) Less(ptI, ptJ *Pellet) bool {
	return ptI.Dist(pm.Pos) >= ptJ.Dist(pm.Pos)
}

type ByPosID struct{}

func (pm ByPosID) Less(ptI, ptJ *Pellet) bool {
	return ptI.ID() < ptJ.ID()
}

type byDescRelativePelletValueTo struct {
	Pos
}

func ByDescRelativePelletValueTo(pos Pos) PelletComparator {
	return byDescRelativePelletValueTo{pos}
}

func (pm byDescRelativePelletValueTo) Less(ptI, ptJ *Pellet) bool {
	return Move{from: ptI.Pos, to: pm.Pos}.Score() >=
		Move{from: ptJ.Pos, to: pm.Pos}.Score()
}
