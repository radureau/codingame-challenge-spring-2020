package main

import "fmt"

func (G *Game) String() string {
	s := G.GameState.String()
	for _, gs := range G.pastStates {
		s += "\n" + gs.String()
	}
	return s
}

func (gs GameState) String() string {
	s := fmt.Sprintf("Game: %.2f\tScore: %.2f\tOpnt: %.2f\n",
		gs.GameProgress(), gs.MyProgress(), gs.OpntProgress())
	s += fmt.Sprintf("Allies: %v\nOpnts: %v\n\n", gs.Allies(), gs.Opnts())
	if gs.turn > 1 {
		s += fmt.Sprintf("Visibles Pellets: %v\n", gs.pellets[0])
	}
	for y := 0; y < G.height; y++ {
		runes := make([]rune, G.width)
		for x := 0; x < G.width; x++ {
			if node, ok := G.graph.nodes[xy(x, y)]; ok {
				r := ' '
				for f, m := range gs.pellets {
					if pl, ok := m[node.Pos]; ok {
						if f == 0 {
							r = pl.Rune()
						} else {
							r = '?'
						}
					}
				}
				if pac, ok := gs.pacs[0][node.Pos]; ok {
					r = pac.Rune()
				}
				runes[x] = r
			} else {
				runes[x] = '#'
			}
		}
		s += string(runes) + "\n"
	}
	return s
}

func (p path) Debug(from Pos) string {
	if len(p) < 1 {
		return fmt.Sprintf("~path~\n%v :\n\n", Move{from, from})
	}
	s := fmt.Sprintf("~path~\n%v :\n", Move{from, p[len(p)-1].Pos})
	for y := 0; y < G.height; y++ {
		runes := make([]rune, G.width)
		for x := 0; x < G.width; x++ {
			if node, ok := G.graph.nodes[xy(x, y)]; ok {
				r := ' '
				if p.contains(node.Pos) {
					r = rune(fmt.Sprintf("%d", G.graph.dists[move(p[0], node)]%10)[0])
				} else if node.Pos == from {
					r = 'X'
				}
				runes[x] = r
			} else {
				runes[x] = '#'
			}
		}
		s += string(runes) + "\n"
	}
	return s
}

// Rune UPPERCASE means ally
func (pac Pac) Rune() rune {
	r := []rune{'𝑃', '𝐹', '𝐶'}[pac.Shifumi]
	if !pac.ally {
		r -= 26
	}
	return r
}

// String UPPERCASE
func (pac Pac) String() string { // shifumi Rune pos freh
	s := fmt.Sprintf("%s %v", string(pac.Shifumi.Rune()), pac.Pos)
	if !pac.ally {
		s += fmt.Sprintf(":%d", pac.freshness)
	}
	return s
}

func (p Pos) String() string {
	x, y := p.xy()
	return fmt.Sprintf("(%d,%d)", x, y)
}

func (mv Move) String() string {
	return fmt.Sprintf("{%v -> %v}", mv.from, mv.to)
}

func (s Shifumi) String() string {
	return []string{"ROCK", "PAPER", "SCISSORS"}[s]
}

// Rune for debug purpose
func (s Shifumi) Rune() rune {
	return []rune{'✊', '✋', '✌'}[s]
}

// Rune _
func (plt Pellet) Rune() rune {
	switch plt.Value {
	case SuperPellet:
		return 'x'
		// return 0x2318
	default:
		return '·'
	}
}

func (plt Pellet) String() string {
	return fmt.Sprintf("%v", plt.hotnessForPac[G.closestAlliesToPos(plt.Pos)[0].PacID])
}
