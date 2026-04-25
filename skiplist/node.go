package skiplist

// Data is a value stored in a Skiplist.
//
// Len defines the logical width of the value. Callers choose what that width
// means: bytes, runes, tokens, text chunks, or any other offset unit.
//
// Len must return a positive value for inserted data and must remain stable
// while the value is stored in a Skiplist. If a value's logical length changes,
// remove it and insert the updated value so spans can be rebuilt correctly.
type Data interface {
	Len() int
}

type node struct {
	data   Data
	levels []level
}

type level struct {
	next *node
	span int
}

func newNode(data Data, height int) *node {
	return &node{
		data:   data,
		levels: make([]level, height),
	}
}
