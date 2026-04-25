package skiplist

import (
	"math/rand"
	"reflect"
	"testing"
)

type intData int

func (d intData) len() int { return 1 }

func cmpInt(a, b Data) int {
	av := a.(intData)
	bv := b.(intData)
	if av < bv {
		return -1
	}
	if av > bv {
		return 1
	}
	return 0
}

func mustAt(t *testing.T, sl *Skiplist, idx int) intData {
	t.Helper()
	v, ok := sl.At(idx)
	if !ok {
		t.Fatalf("expected value at index %d", idx)
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
	got := snapshot(t, sl)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("state mismatch\nwant: %v\ngot:  %v", want, got)
	}
}

func TestInsertSortedAndAt(t *testing.T) {
	sl := New(8, cmpInt)
	for _, v := range []intData{5, 1, 3, 2, 4} {
		sl.Insert(v)
	}

	if sl.Len() != 5 {
		t.Fatalf("expected len 5, got %d", sl.Len())
	}

	for i := 0; i < 5; i++ {
		got := mustAt(t, sl, i)
		want := intData(i + 1)
		if got != want {
			t.Fatalf("index %d: expected %d, got %d", i, want, got)
		}
	}
}

func TestIndexEdits(t *testing.T) {
	sl := New(10, cmpInt)

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

func TestIndexBounds(t *testing.T) {
	sl := New(6, cmpInt)

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

func TestDeleteByValueWithDuplicates(t *testing.T) {
	sl := New(8, cmpInt)
	for _, v := range []intData{1, 2, 2, 2, 3} {
		sl.Insert(v)
	}

	if err := sl.Delete(intData(2), nil); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	assertState(t, sl, []intData{1, 2, 2, 3})

	if err := sl.Delete(intData(2), nil); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	assertState(t, sl, []intData{1, 2, 3})

	if err := sl.Delete(intData(9), nil); err != nil {
		t.Fatalf("delete missing value should not fail: %v", err)
	}
	assertState(t, sl, []intData{1, 2, 3})
}

func TestRandomIndexOpsAgainstSliceModel(t *testing.T) {
	sl := New(12, cmpInt)
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
