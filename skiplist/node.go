package skiplist

type Data interface {
	len() int
}

type Node struct {
	data Data
	next []*Node
	span []int
}

func NewNode(data Data, height int) *Node {
	return &Node{
		data: data,
		next: make([]*Node, height),
		span: make([]int, height),
	}
}

func (n *Node) height() int {
	if n == nil {
		return 0
	}
	return len(n.next)
}
