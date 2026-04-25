// Package skiplist provides a weighted skiplist for offset-based data.
package skiplist

import (
	"errors"
	"math/rand"
)

type Skiplist struct {
	head     *node
	level    int
	length   int
	count    int
	maxLevel int
}

func New(maxLevel int) *Skiplist {
	if maxLevel <= 0 {
		maxLevel = 1
	}

	return &Skiplist{
		head:     newNode(nil, maxLevel),
		level:    1,
		length:   0,
		count:    0,
		maxLevel: maxLevel,
	}
}

// Len returns the total logical length of all values using Data.Len.
func (sl *Skiplist) Len() int {
	if sl == nil {
		return 0
	}
	return sl.length
}

// Count returns the number of values stored in the skiplist.
func (sl *Skiplist) Count() int {
	if sl == nil {
		return 0
	}
	return sl.count
}

func (sl *Skiplist) insertWithPath(value Data, update []*node, rank []int) {
	lvl := sl.randomLevel()
	width := value.Len()

	if lvl > sl.level {
		for i := sl.level; i < lvl; i++ {
			rank[i] = 0
			update[i] = sl.head
			sl.head.span[i] = sl.length
		}
		sl.level = lvl
	}

	node := newNode(value, lvl)
	for i := 0; i < lvl; i++ {
		node.next[i] = update[i].next[i]
		node.span[i] = update[i].span[i] - (rank[0] - rank[i])
		update[i].next[i] = node
		update[i].span[i] = (rank[0] - rank[i]) + width
	}

	for i := lvl; i < sl.level; i++ {
		update[i].span[i] += width
	}

	sl.length += width
	sl.count++
}

// InsertAt inserts value at logical offset, where offset is in [0, Len()].
// If offset falls inside an existing value, value is inserted before that item.
func (sl *Skiplist) InsertAt(offset int, value Data) error {
	if value == nil {
		return errors.New("cannot insert nil value")
	}
	if value.Len() <= 0 {
		return errors.New("cannot insert value with non-positive length")
	}
	if offset < 0 || offset > sl.length {
		return errors.New("offset out of range")
	}

	update := make([]*node, sl.maxLevel)
	rank := make([]int, sl.maxLevel)

	x := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.next[i] != nil && rank[i]+x.span[i] <= offset {
			rank[i] += x.span[i]
			x = x.next[i]
		}
		update[i] = x
	}

	sl.insertWithPath(value, update, rank)
	return nil
}

// At returns the value containing logical offset in O(log n).
func (sl *Skiplist) At(offset int) (Data, bool) {
	data, _, ok := sl.Search(offset)
	return data, ok
}

// Search returns the value containing logical offset and the offset within that value.
func (sl *Skiplist) Search(offset int) (Data, int, bool) {
	if offset < 0 || offset >= sl.length {
		return nil, 0, false
	}

	traversed := 0
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && traversed+x.span[i] <= offset {
			traversed += x.span[i]
			x = x.next[i]
		}
	}

	if x.next[0] == nil {
		return nil, 0, false
	}
	return x.next[0].data, offset - traversed, true
}

// DeleteAt deletes the value containing logical offset in O(log n).
func (sl *Skiplist) DeleteAt(offset int) (Data, error) {
	if offset < 0 || offset >= sl.length {
		return nil, errors.New("offset out of range")
	}

	update := make([]*node, sl.maxLevel)
	traversed := 0
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.next[i] != nil && traversed+x.span[i] <= offset {
			traversed += x.span[i]
			x = x.next[i]
		}
		update[i] = x
	}

	target := update[0].next[0]
	if target == nil {
		return nil, errors.New("offset out of range")
	}

	sl.deleteNode(target, update)
	return target.data, nil
}

func (sl *Skiplist) deleteNode(target *node, update []*node) {
	width := target.data.Len()
	for i := 0; i < sl.level; i++ {
		if update[i].next[i] == target {
			update[i].span[i] += target.spanAt(i) - width
			update[i].next[i] = target.nextAt(i)
		} else {
			update[i].span[i] -= width
		}
	}

	for sl.level > 1 && sl.head.next[sl.level-1] == nil {
		sl.level--
	}
	sl.length -= width
	sl.count--
}

func (n *node) nextAt(level int) *node {
	if n == nil || level < 0 || level >= len(n.next) {
		return nil
	}
	return n.next[level]
}

func (n *node) spanAt(level int) int {
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
