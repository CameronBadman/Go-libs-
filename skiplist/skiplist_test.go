package skiplist

import (
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"
)

type intData int

func (d intData) Len() int { return 1 }

type weightedData struct {
	value int
	width int
}

func (d weightedData) Len() int { return d.width }

func assertInvariants(t *testing.T, sl *Skiplist) {
	t.Helper()

	if sl == nil {
		return
	}
	if sl.maxLevel <= 0 {
		t.Fatalf("maxLevel must be positive, got %d", sl.maxLevel)
	}
	if sl.level < 1 || sl.level > sl.maxLevel {
		t.Fatalf("level %d out of range [1,%d]", sl.level, sl.maxLevel)
	}
	if sl.head == nil {
		t.Fatal("missing head node")
	}
	if len(sl.head.levels) != sl.maxLevel {
		t.Fatalf("head height: expected %d, got %d", sl.maxLevel, len(sl.head.levels))
	}

	afterOffset := map[*node]int{sl.head: 0}
	count := 0
	length := 0
	for x := sl.head.levels[0].next; x != nil; x = x.levels[0].next {
		if x.data == nil {
			t.Fatal("data node has nil data")
		}
		width := x.data.Len()
		if width <= 0 {
			t.Fatalf("data node has non-positive length %d", width)
		}
		count++
		length += width
		afterOffset[x] = length
	}

	if length != sl.length {
		t.Fatalf("stored length mismatch: expected %d, got %d", length, sl.length)
	}
	if count != sl.count {
		t.Fatalf("stored count mismatch: expected %d, got %d", count, sl.count)
	}
	if count == 0 && sl.level != 1 {
		t.Fatalf("empty list should have level 1, got %d", sl.level)
	}

	for level := 0; level < sl.level; level++ {
		for x := sl.head; ; {
			if level >= len(x.levels) {
				t.Fatalf("node at level %d has height %d", level, len(x.levels))
			}

			next := x.levels[level].next
			after, ok := afterOffset[x]
			if !ok {
				t.Fatalf("level %d references node missing from level 0", level)
			}

			wantSpan := sl.length - after
			if next != nil {
				nextAfter, ok := afterOffset[next]
				if !ok {
					t.Fatalf("level %d references node missing from level 0", level)
				}
				wantSpan = nextAfter - after
			}

			if x.levels[level].span != wantSpan {
				t.Fatalf("level %d span mismatch: expected %d, got %d", level, wantSpan, x.levels[level].span)
			}
			if next == nil {
				break
			}
			x = next
		}
	}

	for level := sl.level; level < sl.maxLevel; level++ {
		if sl.head.levels[level].next != nil {
			t.Fatalf("inactive level %d has head link", level)
		}
	}
}

func mustAt(t *testing.T, sl *Skiplist, offset int) intData {
	t.Helper()
	v, ok := sl.At(offset)
	if !ok {
		t.Fatalf("expected value at offset %d", offset)
	}
	return v.(intData)
}

func snapshot(t *testing.T, sl *Skiplist) []intData {
	t.Helper()
	out := make([]intData, 0, sl.Len())
	for i := 0; i < sl.Len(); i++ {
		out = append(out, mustAt(t, sl, i))
	}
	return out
}

func assertState(t *testing.T, sl *Skiplist, want []intData) {
	t.Helper()
	assertInvariants(t, sl)
	if sl.Len() != len(want) {
		t.Fatalf("expected len %d, got %d", len(want), sl.Len())
	}
	if sl.Count() != len(want) {
		t.Fatalf("expected count %d, got %d", len(want), sl.Count())
	}
	got := snapshot(t, sl)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("state mismatch\nwant: %v\ngot:  %v", want, got)
	}
}

func TestOffsetEdits(t *testing.T) {
	sl := New(10)

	for i, v := range []intData{10, 20, 30, 40} {
		if err := sl.InsertAt(i, v); err != nil {
			t.Fatalf("insert %d failed: %v", i, err)
		}
	}

	if err := sl.InsertAt(2, intData(25)); err != nil {
		t.Fatalf("insert at middle failed: %v", err)
	}

	if got := mustAt(t, sl, 2); got != 25 {
		t.Fatalf("expected inserted value at index 2, got %d", got)
	}

	deleted, err := sl.DeleteAt(1)
	if err != nil {
		t.Fatalf("delete at failed: %v", err)
	}
	if deleted.(intData) != 20 {
		t.Fatalf("expected deleted 20, got %d", deleted.(intData))
	}

	assertState(t, sl, []intData{10, 25, 30, 40})
}

func TestWeightedOffsets(t *testing.T) {
	sl := New(10)

	for _, item := range []weightedData{
		{value: 10, width: 3},
		{value: 20, width: 4},
		{value: 30, width: 2},
	} {
		if err := sl.InsertAt(sl.Len(), item); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
		assertInvariants(t, sl)
	}

	if sl.Len() != 9 {
		t.Fatalf("expected logical len 9, got %d", sl.Len())
	}
	if sl.Count() != 3 {
		t.Fatalf("expected count 3, got %d", sl.Count())
	}

	tests := []struct {
		offset int
		want   int
		inner  int
	}{
		{0, 10, 0},
		{2, 10, 2},
		{3, 20, 0},
		{6, 20, 3},
		{7, 30, 0},
		{8, 30, 1},
	}
	for _, tt := range tests {
		got, inner, ok := sl.Search(tt.offset)
		if !ok {
			t.Fatalf("expected value at offset %d", tt.offset)
		}
		if got.(weightedData).value != tt.want {
			t.Fatalf("offset %d: expected %d, got %d", tt.offset, tt.want, got.(weightedData).value)
		}
		if inner != tt.inner {
			t.Fatalf("offset %d: expected inner offset %d, got %d", tt.offset, tt.inner, inner)
		}
	}

	if _, ok := sl.At(9); ok {
		t.Fatal("expected At(9) to fail")
	}

	deleted, err := sl.DeleteAt(4)
	if err != nil {
		t.Fatalf("delete at weighted offset failed: %v", err)
	}
	assertInvariants(t, sl)
	if deleted.(weightedData).value != 20 {
		t.Fatalf("expected deleted value 20, got %d", deleted.(weightedData).value)
	}
	if sl.Len() != 5 {
		t.Fatalf("expected logical len 5 after delete, got %d", sl.Len())
	}
	if sl.Count() != 2 {
		t.Fatalf("expected count 2 after delete, got %d", sl.Count())
	}

	got, ok := sl.At(3)
	if !ok {
		t.Fatal("expected value at offset 3 after delete")
	}
	if got.(weightedData).value != 30 {
		t.Fatalf("expected value 30 at offset 3 after delete, got %d", got.(weightedData).value)
	}
}

func TestRandomWeightedOffsetOpsAgainstModel(t *testing.T) {
	sl := New(16)
	model := make([]weightedData, 0)
	rng := rand.New(rand.NewSource(99))

	for step := 0; step < 1000; step++ {
		doInsert := len(model) == 0 || rng.Intn(100) < 70
		if doInsert {
			offset := rng.Intn(totalWeightedLen(model) + 1)
			item := weightedData{
				value: step,
				width: rng.Intn(8) + 1,
			}
			if err := sl.InsertAt(offset, item); err != nil {
				t.Fatalf("step %d: InsertAt(%d, %+v) failed: %v", step, offset, item, err)
			}
			idx := insertIndexForOffset(model, offset)
			model = append(model, weightedData{})
			copy(model[idx+1:], model[idx:])
			model[idx] = item
		} else {
			offset := rng.Intn(totalWeightedLen(model))
			idx, _, ok := weightedIndexForOffset(model, offset)
			if !ok {
				t.Fatalf("step %d: model missing offset %d", step, offset)
			}
			deleted, err := sl.DeleteAt(offset)
			if err != nil {
				t.Fatalf("step %d: DeleteAt(%d) failed: %v", step, offset, err)
			}
			if deleted.(weightedData) != model[idx] {
				t.Fatalf("step %d: DeleteAt(%d) expected %+v got %+v", step, offset, model[idx], deleted)
			}
			model = append(model[:idx], model[idx+1:]...)
		}

		assertWeightedState(t, sl, model)
	}
}

func TestMaxLevelEdges(t *testing.T) {
	for _, maxLevel := range []int{-3, 0, 1, 2, 64} {
		t.Run(strconv.Itoa(maxLevel), func(t *testing.T) {
			sl := New(maxLevel)
			assertInvariants(t, sl)

			if err := sl.InsertAt(0, weightedData{value: 1, width: 3}); err != nil {
				t.Fatalf("insert first failed: %v", err)
			}
			if err := sl.InsertAt(sl.Len(), weightedData{value: 2, width: 5}); err != nil {
				t.Fatalf("insert second failed: %v", err)
			}
			assertWeightedState(t, sl, []weightedData{
				{value: 1, width: 3},
				{value: 2, width: 5},
			})

			if _, err := sl.DeleteAt(0); err != nil {
				t.Fatalf("delete first failed: %v", err)
			}
			if _, err := sl.DeleteAt(sl.Len() - 1); err != nil {
				t.Fatalf("delete second failed: %v", err)
			}
			assertInvariants(t, sl)
			if sl.Len() != 0 || sl.Count() != 0 {
				t.Fatalf("expected empty list, got len=%d count=%d", sl.Len(), sl.Count())
			}
		})
	}
}

func TestOffsetBounds(t *testing.T) {
	sl := New(6)
	assertInvariants(t, sl)

	if _, ok := sl.At(0); ok {
		t.Fatal("expected At(0) on empty list to fail")
	}
	if err := sl.InsertAt(-1, intData(1)); err == nil {
		t.Fatal("expected InsertAt(-1) to fail")
	}
	if err := sl.InsertAt(1, intData(1)); err == nil {
		t.Fatal("expected InsertAt(1) on empty list to fail")
	}
	if _, err := sl.DeleteAt(0); err == nil {
		t.Fatal("expected DeleteAt(0) on empty list to fail")
	}

	if err := sl.InsertAt(0, intData(42)); err != nil {
		t.Fatalf("insert at 0 failed: %v", err)
	}
	assertInvariants(t, sl)
	if _, ok := sl.At(-1); ok {
		t.Fatal("expected At(-1) to fail")
	}
	if _, ok := sl.At(1); ok {
		t.Fatal("expected At(1) to fail with len=1")
	}
	if _, err := sl.DeleteAt(1); err == nil {
		t.Fatal("expected DeleteAt(1) with len=1 to fail")
	}
}

func TestRandomOffsetOpsAgainstSliceModel(t *testing.T) {
	sl := New(12)
	model := make([]intData, 0)
	rng := rand.New(rand.NewSource(7))

	for step := 0; step < 500; step++ {
		doInsert := len(model) == 0 || rng.Intn(100) < 65
		if doInsert {
			idx := rng.Intn(len(model) + 1)
			val := intData(rng.Intn(1000))
			if err := sl.InsertAt(idx, val); err != nil {
				t.Fatalf("step %d: insertAt(%d,%d) failed: %v", step, idx, val, err)
			}
			model = append(model, 0)
			copy(model[idx+1:], model[idx:])
			model[idx] = val
		} else {
			idx := rng.Intn(len(model))
			got, err := sl.DeleteAt(idx)
			if err != nil {
				t.Fatalf("step %d: deleteAt(%d) failed: %v", step, idx, err)
			}
			if got.(intData) != model[idx] {
				t.Fatalf("step %d: deleteAt(%d) expected %d got %d", step, idx, model[idx], got.(intData))
			}
			model = append(model[:idx], model[idx+1:]...)
		}

		assertState(t, sl, model)
	}
}

func stressCount() int {
	if raw := os.Getenv("SKIPLIST_STRESS_N"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return 200000
}

func TestMassivePointLoadAndAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping massive stress test in short mode")
	}

	n := stressCount()
	sl := New(20)

	for i := 0; i < n; i++ {
		if err := sl.InsertAt(sl.Len(), intData(i)); err != nil {
			t.Fatalf("append at %d failed: %v", i, err)
		}
	}

	if sl.Len() != n {
		t.Fatalf("expected len %d, got %d", n, sl.Len())
	}
	assertInvariants(t, sl)

	checkpoints := []int{0, n / 4, n / 2, (3 * n) / 4, n - 1}
	for _, idx := range checkpoints {
		got := mustAt(t, sl, idx)
		if got != intData(idx) {
			t.Fatalf("checkpoint %d: expected %d got %d", idx, idx, got)
		}
	}

	rng := rand.New(rand.NewSource(12345))
	for i := 0; i < 2000; i++ {
		idx := rng.Intn(n)
		got := mustAt(t, sl, idx)
		if got != intData(idx) {
			t.Fatalf("random check %d at idx %d: expected %d got %d", i, idx, idx, got)
		}
	}

	deleteHead := 500
	if deleteHead > sl.Len() {
		deleteHead = sl.Len()
	}
	for i := 0; i < deleteHead; i++ {
		deleted, err := sl.DeleteAt(0)
		if err != nil {
			t.Fatalf("head delete %d failed: %v", i, err)
		}
		if deleted.(intData) != intData(i) {
			t.Fatalf("head delete %d: expected %d got %d", i, i, deleted.(intData))
		}
	}

	deleteTail := 500
	if deleteTail > sl.Len() {
		deleteTail = sl.Len()
	}
	for i := 0; i < deleteTail; i++ {
		idx := sl.Len() - 1
		deleted, err := sl.DeleteAt(idx)
		if err != nil {
			t.Fatalf("tail delete %d failed: %v", i, err)
		}
		want := intData(n - 1 - i)
		if deleted.(intData) != want {
			t.Fatalf("tail delete %d: expected %d got %d", i, want, deleted.(intData))
		}
	}

	expectedLen := n - deleteHead - deleteTail
	if sl.Len() != expectedLen {
		t.Fatalf("expected len after deletes %d, got %d", expectedLen, sl.Len())
	}
	assertInvariants(t, sl)
}

func assertWeightedState(t *testing.T, sl *Skiplist, want []weightedData) {
	t.Helper()
	assertInvariants(t, sl)

	if sl.Count() != len(want) {
		t.Fatalf("expected count %d, got %d", len(want), sl.Count())
	}
	if sl.Len() != totalWeightedLen(want) {
		t.Fatalf("expected len %d, got %d", totalWeightedLen(want), sl.Len())
	}

	for offset := 0; offset < sl.Len(); offset++ {
		idx, inner, ok := weightedIndexForOffset(want, offset)
		if !ok {
			t.Fatalf("model missing offset %d", offset)
		}
		got, gotInner, ok := sl.Search(offset)
		if !ok {
			t.Fatalf("skiplist missing offset %d", offset)
		}
		if got.(weightedData) != want[idx] {
			t.Fatalf("offset %d: expected %+v, got %+v", offset, want[idx], got)
		}
		if gotInner != inner {
			t.Fatalf("offset %d: expected inner offset %d, got %d", offset, inner, gotInner)
		}
	}
}

func totalWeightedLen(items []weightedData) int {
	total := 0
	for _, item := range items {
		total += item.Len()
	}
	return total
}

func insertIndexForOffset(items []weightedData, offset int) int {
	traversed := 0
	for i, item := range items {
		traversed += item.Len()
		if traversed > offset {
			return i
		}
	}
	return len(items)
}

func weightedIndexForOffset(items []weightedData, offset int) (int, int, bool) {
	traversed := 0
	for i, item := range items {
		next := traversed + item.Len()
		if offset < next {
			return i, offset - traversed, true
		}
		traversed = next
	}
	return 0, 0, false
}
