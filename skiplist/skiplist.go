// Package skiplist basic general skiplist meant for text
package skiplist

import (
	"errors"
	"math/rand"
)

type Skiplist struct {
	head     *Node
	level    int
	length   int
	maxLevel int
	compare  func(a, b Data) int
}

func New(maxLevel int, compare func(a, b Data) int) *Skiplist {
	if maxLevel <= 0 {
		maxLevel = 1
	}

	return &Skiplist{
		head:     NewNode(nil, maxLevel),
		level:    1,
		length:   0,
		maxLevel: maxLevel,
		compare:  compare,
	}
}

func (sl *Skiplist) Len() int {
	if sl == nil {
		return 0
	}
	return sl.length
}

// uses comparable on the data and puts the value at where a < value < b
func (sl *Skiplist) Insert(value Data) {
	if value == nil {
		panic("cannot insert nil value")
	}
	if sl.compare == nil {
		panic("compare function is required for Insert")
	}

	update := make([]*Node, sl.maxLevel)
	rank := make([]int, sl.maxLevel)

	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.next[i] != nil && sl.compare(x.next[i].data, value) < 0 {
			rank[i] += x.span[i]
			x = x.next[i]
		}
		update[i] = x
	}

	sl.insertWithPath(value, update, rank)
}

// traverses using a arg comparable a < value < b returns traversed levels
func (sl *Skiplist) Traverse(data Data, compare func(a, b Data) int) []*Node {
	if compare == nil {
		compare = sl.compare
	}
	if compare == nil {
		return nil
	}

	update := make([]*Node, sl.maxLevel)
	ptr := sl.head
	for level := sl.level - 1; level >= 0; level-- {
		for ptr.next[level] != nil && compare(ptr.next[level].data, data) < 0 {
			ptr = ptr.next[level]
		}
		update[level] = ptr
	}
	for level := sl.level; level < sl.maxLevel; level++ {
		update[level] = sl.head
	}
	return update
}

func (sl *Skiplist) insertWithPath(value Data, update []*Node, rank []int) {
	lvl := sl.randomLevel()

	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			rank[i] = 0
			update[i] = sl.head
			sl.head.span[i] = sl.length
		}
		sl.level = lvl
	}

	node := NewNode(value, lvl)
	for i := 0; i < lvl; i++ {
		node.next[i] = update[i].next[i]
		node.span[i] = update[i].span[i] - (rank[0] - rank[i])
		update[i].next[i] = node
		update[i].span[i] = (rank[0] - rank[i]) + 1
	}

	for i := lvl; i < sl.level; i++ {
		update[i].span[i]++
	}

	sl.length++
}

// InsertAt inserts value before current index, where index in [0, Len()].
func (sl *Skiplist) InsertAt(index int, value Data) error {
	if value == nil {
		return errors.New("cannot insert nil value")
	}
	if index < 0 || index > sl.length {
		return errors.New("index out of range")
	}

	update := make([]*Node, sl.maxLevel)
	rank := make([]int, sl.maxLevel)

	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.next[i] != nil && rank[i]+x.span[i] <= index {
			rank[i] += x.span[i]
			x = x.next[i]
		}
		update[i] = x
	}

	sl.insertWithPath(value, update, rank)
	return nil
}

// At returns the value at index in O(log n).
func (sl *Skiplist) At(index int) (Data, bool) {
	if index < 0 || index >= sl.length {
		return nil, false
	}

	targetRank := index + 1
	traversed := 0
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && traversed+x.span[i] <= targetRank {
			traversed += x.span[i]
			x = x.next[i]
		}
		if traversed == targetRank {
			return x.data, true
		}
	}

	return nil, false
}

// Delete finds first instance of node then deletes it, requires compare.
func (sl *Skiplist) Delete(data Data, compare func(a, b Data) int) error {
	if compare == nil {
		compare = sl.compare
	}
	if compare == nil {
		return errors.New("compare function is required")
	}

	update := sl.Traverse(data, compare)
	target := update[0].next[0]
	if target == nil || compare(target.data, data) != 0 {
		return nil
	}

	sl.deleteNode(target, update)
	return nil
}

// DeleteAt deletes the value at index in O(log n).
func (sl *Skiplist) DeleteAt(index int) (Data, error) {
	if index < 0 || index >= sl.length {
		return nil, errors.New("index out of range")
	}

	update := make([]*Node, sl.maxLevel)
	traversed := 0
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && traversed+x.span[i] <= index {
			traversed += x.span[i]
			x = x.next[i]
		}
		update[i] = x
	}

	target := update[0].next[0]
	if target == nil {
		return nil, errors.New("index out of range")
	}

	sl.deleteNode(target, update)
	return target.data, nil
}

func (sl *Skiplist) deleteNode(target *Node, update []*Node) {
	for i := 0; i < sl.level; i++ {
		if update[i].next[i] == target {
			update[i].span[i] += target.spanAt(i) - 1
			update[i].next[i] = target.nextAt(i)
		} else {
			update[i].span[i]--
		}
	}

	for sl.level > 1 && sl.head.next[sl.level-1] == nil {
		sl.level--
	}
	sl.length--
}

func (n *Node) nextAt(level int) *Node {
	if n == nil || level < 0 || level >= len(n.next) {
		return nil
	}
	return n.next[level]
}

func (n *Node) spanAt(level int) int {
	if n == nil || level < 0 || level >= len(n.span) {
		return 0
	}
	return n.span[level]
}

func (sl *Skiplist) randomLevel() int {
	level := 1
	for level < sl.maxLevel && rand.Int63()&1 == 0 {
		level++
	}
	return level
}
