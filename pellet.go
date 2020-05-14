package main

// Pellets possible score point values
const (
	NormalPellet = ScorePoint(1)
	SuperPellet  = ScorePoint(10)
	Nought       = ScorePoint(0)
)

// Pellet _
type Pellet struct {
	Pos
	Value ScorePoint
}
