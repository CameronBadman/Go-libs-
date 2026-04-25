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

func TestOffsetBounds(t *testing.T) {
	sl := New(6)

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
}
