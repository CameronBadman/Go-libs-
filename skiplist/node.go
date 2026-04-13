package skiplist

type Data interface {
	len() int
}

type Node struct {
	data Data
	next []*Node
}

func NewNode(data Data, cap int) *Node {
	return &Node{
		data: data,
		next: make([]*Node, 0, cap),
	}
}
