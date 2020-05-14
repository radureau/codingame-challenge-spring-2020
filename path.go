package main

type path []*Node

func (p path) index(pos Pos) (idx int) {
	for i, node := range p {
		if node.Pos == pos {
			return i
		}
	}
	return -1
}

func (p path) contains(pos Pos) bool {
	for _, node := range p {
		if node.Pos == pos {
			return true
		}
	}
	return false
}
