package main

// Pac _
type Pac struct {
	PacID
	Pos
	Shifumi
	speedTurnsLeft  turn
	abilityCooldown turn
}

// PacID _
type PacID struct {
	ID   int  // pacID: pac number (unique within a team)
	ally bool // true if this pac is yours
}

// Shifumi _
type Shifumi int

// Pac type
const (
	ROCK = Shifumi(iota)
	PAPER
	SCISSORS
)
