// Package skiplist provides a weighted skiplist for offset-based data.
//
// Each stored value contributes Data.Len units to the list. Offsets used by
// InsertAt, Search, At, and DeleteAt are measured in those units rather than in
// node count. This makes the package suitable for text chunks, byte ranges,
// token streams, and other structures where callers need O(log n) lookup by a
// logical position.
package skiplist

import (
	"errors"
	"math/rand"
)

// Skiplist stores Data values and maintains weighted spans for O(log n)
// offset-based search, insertion, and deletion.
//
// Skiplist is not safe for concurrent use. Synchronize access externally if a
// list is shared across goroutines.
type Skiplist struct {
	head     *node
	level    int
	length   int
	count    int
	maxLevel int
}

// New returns an empty Skiplist with the given maximum tower height.
//
// Higher maxLevel values allow taller skiplist towers for large lists. If
// maxLevel is less than or equal to zero, New uses 1.
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

// Len returns the total logical length of all stored values.
//
// Len is the sum of Data.Len for every value in the list. Use Count to get the
// number of stored values.
func (sl *Skiplist) Len() int {
	if sl == nil {
		return 0
	}
	return sl.length
}

// Count returns the number of values stored in the list.
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
			sl.head.levels[i].span = sl.length
		}
		sl.level = lvl
	}

	node := newNode(value, lvl)
	for i := 0; i < lvl; i++ {
		node.levels[i].next = update[i].levels[i].next
		node.levels[i].span = update[i].levels[i].span - (rank[0] - rank[i])
		update[i].levels[i].next = node
		update[i].levels[i].span = (rank[0] - rank[i]) + width
	}

	for i := lvl; i < sl.level; i++ {
		update[i].levels[i].span += width
	}

	sl.length += width
	sl.count++
}

// InsertAt inserts value at logical offset.
//
// The offset must be in the range [0, Len()]. If offset falls inside an
// existing value, InsertAt inserts value before that existing value. Splitting
// values at inner offsets is the caller's responsibility.
//
// InsertAt returns an error if value is nil, value.Len() is not positive, or
// offset is out of range.
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

		for x.levels[i].next != nil && rank[i]+x.levels[i].span <= offset {
			rank[i] += x.levels[i].span
			x = x.levels[i].next
		}
		update[i] = x
	}

	sl.insertWithPath(value, update, rank)
	return nil
}

// At returns the value containing logical offset in O(log n).
//
// At is a convenience wrapper around Search that discards the inner offset.
func (sl *Skiplist) At(offset int) (Data, bool) {
	data, _, ok := sl.Search(offset)
	return data, ok
}

// Search returns the value containing logical offset and the offset within that
// value.
//
// For example, if the list contains values with lengths [5, 3] then Search(6)
// returns the second value with inner offset 1. Search returns ok=false if
// offset is outside [0, Len()).
func (sl *Skiplist) Search(offset int) (Data, int, bool) {
	if offset < 0 || offset >= sl.length {
		return nil, 0, false
	}

	traversed := 0
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.levels[i].next != nil && traversed+x.levels[i].span <= offset {
			traversed += x.levels[i].span
			x = x.levels[i].next
		}
	}

	if x.levels[0].next == nil {
		return nil, 0, false
	}
	return x.levels[0].next.data, offset - traversed, true
}

// DeleteAt deletes and returns the whole value containing logical offset in
// O(log n).
//
// DeleteAt does not split values. If offset falls inside a value, the entire
// value is removed. DeleteAt returns an error if offset is outside [0, Len()).
func (sl *Skiplist) DeleteAt(offset int) (Data, error) {
	if offset < 0 || offset >= sl.length {
		return nil, errors.New("offset out of range")
	}

	update := make([]*node, sl.maxLevel)
	traversed := 0
	x := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for x.levels[i].next != nil && traversed+x.levels[i].span <= offset {
			traversed += x.levels[i].span
			x = x.levels[i].next
		}
		update[i] = x
	}

	target := update[0].levels[0].next
	if target == nil {
		return nil, errors.New("offset out of range")
	}

	sl.deleteNode(target, update)
	return target.data, nil
}

func (sl *Skiplist) deleteNode(target *node, update []*node) {
	width := target.data.Len()
	for i := 0; i < sl.level; i++ {
		if update[i].levels[i].next == target {
			update[i].levels[i].span += target.spanAt(i) - width
			update[i].levels[i].next = target.nextAt(i)
		} else {
			update[i].levels[i].span -= width
		}
	}

	for sl.level > 1 && sl.head.levels[sl.level-1].next == nil {
		sl.level--
	}
	sl.length -= width
	sl.count--
}

func (n *node) nextAt(level int) *node {
	if n == nil || level < 0 || level >= len(n.levels) {
		return nil
	}
	return n.levels[level].next
}

func (n *node) spanAt(level int) int {
	if n == nil || level < 0 || level >= len(n.levels) {
		return 0
	}
	return n.levels[level].span
}

func (sl *Skiplist) randomLevel() int {
	level := 1
	for level < sl.maxLevel && rand.Int63()&1 == 0 {
		level++
	}
	return level
}
