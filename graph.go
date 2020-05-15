package main

// Graph _
type Graph struct {
	nodes      map[Pos]*Node
	positions  []Pos                     // sorted
	dists      map[Move]Dist             // without any bumps
	paths      map[Move]path             // without any bumps
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
	compute := func(node *Node, _ Dist, visited []*Node) {
		wasVisited := func(node *Node) bool {
			for _, n := range visited {
				if n.Pos == node.Pos {
					return true
				}
			}
			return false
		}
		g.paths[move(node, node)] = make(path, 0)
		for _, neighbour := range node.neighbours {
			g.dists[move(node, neighbour)] = Dist(1)
			g.dists[move(neighbour, node)] = Dist(1)
			if _, ok := g.paths[move(node, neighbour)]; !ok {
				g.paths[move(node, neighbour)] = path{neighbour}
			}
			if _, ok := g.paths[move(neighbour, node)]; !ok {
				g.paths[move(neighbour, node)] = path{node}
			}
			if wasVisited(neighbour) {
				for _, n := range visited {
					if n.Pos != node.Pos && n.Pos != neighbour.Pos {
						nDist := g.dists[move(neighbour, n)] + 1
						g.dists[move(node, n)] = nDist
						g.dists[move(n, node)] = nDist
						if _, ok := g.paths[move(node, n)]; !ok {
							if nPath, ok := g.paths[move(neighbour, n)]; ok {
								g.paths[move(node, n)] = append(path{neighbour}, nPath...)
								g.paths[move(n, node)] = append(g.paths[move(n, neighbour)], node)
							}
						}
					}
				}
			}
		}
	}
	g.breadthFirstSearch(nil, compute)
	for _, node := range g.nodes {
		lastDist := Dist(1)
		g.influences[speed1][node.Pos] = influence{0: {node}, 1: append([]*Node{node}, node.neighbours...)}
		g.influences[speed2][node.Pos] = influence{0: {node}}
		g.breadthFirstSearch(node, func(n *Node, dist Dist, visited []*Node) {
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
