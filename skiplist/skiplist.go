// Package skiplist basic general skiplist meant for text
package skiplist

import (
	"math/rand"
)

type Skiplist struct {
	head     *Node
	level    int
	maxLevel int
	compare  func(a, b Data) int
}

func New(maxLevel int, compare func(a, b Data) int) *Skiplist {
	return &Skiplist{
		head:     NewNode(nil, maxLevel),
		level:    0,
		maxLevel: maxLevel,
		compare:  compare,
	}
}

// uses comparable on the data and puts the value at where a < value < b
func (sl *Skiplist) Insert(value Data) {
	if value == nil {
		panic("cannot insert nil value")
	}
	update := sl.Traverse(value, sl.compare)
	level := sl.randomLevel()
	node := NewNode(value, level+1)
	sl.insertNode(node, update[:level+1])
}

// traverses using a arg comparable a < value < b returns traversed levels
func (sl *Skiplist) Traverse(data Data, compare func(a, b Data) int) []*Node {
	update := make([]*Node, sl.maxLevel)
	ptr := sl.head
	for level := sl.maxLevel - 1; level >= 0; level-- {
		for ptr.next[level] != nil && compare(ptr.next[level].data, data) < 0 {
			ptr = ptr.next[level]
		}
		update[level] = ptr
	}
	return update
}

// inserts node at insertPoints, internal method
func (sl *Skiplist) insertNode(node *Node, insertPoints []*Node) error {
	for index, n := range insertPoints {
		node.next[index] = n.next[index]
		n.next[index] = node
	}

	return nil
}

// Delete finds first insance of node them deletes it, requires compare
func (sl *Skiplist) Delete(data Data, compare func(a, b Data) int) error {
	update := sl.Traverse(data, compare)
	target := update[0].next[0]

	if target == nil || compare(target.data, data) != 0 {
		return nil
	}

	for i := 0; i < sl.maxLevel; i++ {
		if update[i].next[i] != target {
			break
		}
		update[i].next[i] = target.next[i]
	}

	return nil
}

func (sl *Skiplist) randomLevel() int {
	level := 0
	for level < sl.maxLevel && rand.Int63()&1 == 0 {
		level++
	}
	return level
}
