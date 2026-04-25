package skiplist

import "testing"

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

	want := []intData{10, 25, 30, 40}
	if sl.Len() != len(want) {
		t.Fatalf("expected len %d, got %d", len(want), sl.Len())
	}
	for i, w := range want {
		if got := mustAt(t, sl, i); got != w {
			t.Fatalf("index %d: expected %d, got %d", i, w, got)
		}
	}
}
