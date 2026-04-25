package skiplist

type Data interface {
	Len() int
}

type node struct {
	data Data
	next []*node
	span []int
}

func newNode(data Data, height int) *node {
	return &node{
		data: data,
		next: make([]*node, height),
		span: make([]int, height),
	}
}
