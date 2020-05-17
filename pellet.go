package main

// Pellets possible score point values
const (
	NormalPellet = ScorePoint(1)
	SuperPellet  = ScorePoint(10)
	Nought       = ScorePoint(0)
)

// Pellet _
type Pellet struct {
	*Node
	Value         ScorePoint
	hotnessForPac map[PacID]hotness
}

func (plt *Pellet) initHotness() {
	plt.hotnessForPac = make(map[PacID]hotness)
	for _, pac := range G.pacs[0] {
		plt.hotnessForPac[pac.PacID] = 1
		if plt.Value == SuperPellet {
			plt.hotnessForPac[pac.PacID] = 1000
		}
	}
}

type SortedPellet struct {
	*Pac
	plts []*Pellet
}

func (s SortedPellet) Len() int      { return len(s.plts) }
func (s SortedPellet) Swap(i, j int) { s.plts[i], s.plts[j] = s.plts[j], s.plts[i] }
func (s SortedPellet) Less(i, j int) bool {
	pi, pj := s.plts[i], s.plts[j]
	if pi.Value < pj.Value { // SuperPellet top priority
		return true
	}
	piClosest := G.closestPacToPos(pi.Pos, nil)
	pjClosest := G.closestPacToPos(pj.Pos, nil)
	if pjClosest[0].dist(s.Pos) < piClosest[0].dist(s.Pos) {
		return true
	}

	return s.Pac.dist(pi.Pos) < s.Pac.dist(pj.Pos) ||
		pi.Pos < pj.Pos
}
