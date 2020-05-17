package main

import "sort"

// Graph _
type Graph struct {
	nodes      map[Pos]*Node
	positions  []Pos                     // sorted
	dists      map[Move]Dist             // without any bumps
	paths      map[Move]path             // shortest path without any bumps
	influences [Nspeed]map[Pos]influence // influences with speed1 at index 0 and influences with speed2 at index 1
}

// Node _
type Node struct {
	Pos
	neighbours []*Node       // sorted by Direction order
	linkedWith map[Pos]*Node // nodes with same row or column without any wall in between
}

// NewGraph _
func NewGraph(capacity int) *Graph {
	g := new(Graph)
	g.nodes = make(map[Pos]*Node, capacity)
	g.positions = make([]Pos, 0, capacity)
	g.dists = make(map[Move]Dist, capacity)
	g.paths = make(map[Move]path, capacity)
	g.influences = [Nspeed]map[Pos]influence{make(map[Pos]influence, capacity), make(map[Pos]influence, capacity)}
	return g
}
func (g *Graph) createNode(x, y int) {
	nC := &Node{
		Pos:        xy(x, y),
		neighbours: make([]*Node, 0, 4),
		linkedWith: make(map[Pos]*Node),
	}
	g.nodes[nC.Pos] = nC
	g.positions = append(g.positions, nC.Pos)
}
func (g Graph) linkTogether() {
	for _, node := range g.nodes {
		for _, dir := range Directions {
			if n, ok := g.nodes[node.ToDirection(dir)]; ok {
				node.neighbours = append(node.neighbours, n)
			}
			current, ok := g.nodes[node.ToDirection(dir)]
			for ok {
				if current.Pos == node.Pos {
					break
				}
				node.linkedWith[current.Pos] = current
				current, ok = g.nodes[current.ToDirection(dir)]
			}
		}
	}
}
func (g Graph) breadthFirstSearch(from *Node, compute func(node *Node, dist Dist, visited []*Node)) {
	visited := make([]*Node, 0, len(g.nodes))
	markAsVisited := func(n *Node) {
		visited = append(visited, n)
	}
	wasVisited := func(node *Node) bool {
		for _, n := range visited {
			if n.Pos == node.Pos {
				return true
			}
		}
		return false
	}
	start := from
	if start == nil {
		start = g.nodes[Pos(g.positions[0])]
	}
	markAsVisited(start)
	type toVisit struct {
		*Node
		Dist
	}
	visits := []toVisit{{start, 0}}
	for len(visits) > 0 {
		var current toVisit
		current, visits = visits[0], visits[1:]
		compute(current.Node, current.Dist, visited)
		for _, n := range current.neighbours {
			if !wasVisited(n) {
				markAsVisited(n)
				visits = append(visits, toVisit{n, current.Dist + 1})
			}
		}
	}
}
func (g Graph) compute() {
	for _, node := range g.nodes {
		g.dists[move(node, node)] = 0
		lastDist := Dist(1)
		g.influences[speed1][node.Pos] = influence{0: {node}, 1: append([]*Node{node}, node.neighbours...)}
		g.influences[speed2][node.Pos] = influence{0: {node}}
		g.breadthFirstSearch(node, func(n *Node, dist Dist, visited []*Node) {
			g.dists[move(node, n)] = dist
			if lastDist == dist-1 {
				lastDist = dist
				t := turn(dist)
				if t%2 == 0 {
					g.influences[speed2][node.Pos][t/2] = append(visited)
				}
				g.influences[speed1][node.Pos][t] = append(visited)
			}
		})
	}
}

// SortByDistanceToPos _
type SortByDistanceToPos struct {
	Pos
	positions []Pos
}

func (s SortByDistanceToPos) Len() int { return len(s.positions) }
func (s SortByDistanceToPos) Swap(i, j int) {
	s.positions[i], s.positions[j] = s.positions[j], s.positions[i]
}
func (s SortByDistanceToPos) Less(i, j int) bool {
	return s.positions[i].dist(s.Pos) < s.positions[j].dist(s.Pos)
}

// Sort _
func (s SortByDistanceToPos) Sort() {
	sort.Sort(s)
}
