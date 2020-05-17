package main

type turn int
type freshness int

func (t turn) freshness() freshness { return freshness(G.turn - t) }
func (f freshness) turn() turn      { return G.turn - turn(f) }
func (f freshness) fresher() freshness {
	if f > 1 {
		return f - 1
	}
	return 0
}

// MaxTurn _
const MaxTurn = turn(200)

type speed int

const (
	speed1 speed = iota
	speed2
	// Nspeed number of speeds
	Nspeed
)

// ScorePoint _
type ScorePoint int

// Dist _
type Dist int

// Pos _
type Pos int

func xy(x, y int) Pos {
	return Pos(y*G.width + x)
}

func (p Pos) sym() Pos {
	return xy(-(int(p)%G.width)+G.width-1, int(p)/G.width)
}

func (p Pos) xy() (int, int) {
	return int(p) % G.width, int(p) / G.width
}

func (p Pos) dist(to Pos) Dist {
	return G.graph.dists[Move{p, to}]
}

type direction struct {
	x, y int
}

// Directions
var (
	up, down, right, left = direction{0, -1}, direction{0, 1}, direction{1, 0}, direction{-1, 0}
	Directions            = []direction{up, down, right, left} // order used when moving
)

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

func move(from, to *Node) Move {
	return Move{from.Pos, to.Pos}
}

type hotness int
