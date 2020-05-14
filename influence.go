package main

// influence: from a position, in how many turns can I get to how many cells (ordered by dist) ?
type influence map[turn][]*Node

func (nflc influence) addCells(at turn, cells ...*Node) {
	nflc[at] = append(nflc[at], cells...)
}
func (nflc influence) at(t turn) []*Node {
	return nflc[t]
}
func (nflc influence) containsCell(at turn, node *Node) bool {
	for _, n := range nflc[at] {
		if n.Pos == node.Pos {
			return true
		}
	}
	return false
}
