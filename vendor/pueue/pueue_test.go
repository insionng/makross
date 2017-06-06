package pueue

import (
	"math/rand"
	"sort"
	"testing"
)

type Int int

func (this Int) Less(other interface{}) bool {
	return this < other.(Int)
}

type IntSorter []Int

func (s *IntSorter) Len() int {
	return len(*s)
}

func (s *IntSorter) Less(i, j int) bool {
	return (*s)[i] < (*s)[j]
}

func (s *IntSorter) Swap(i, j int) {
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

func TestInt(t *testing.T) {
	q := New()

	if q.Len() != 0 {
		t.Fatal()
	}

	q.Push(Int(-1))
	for i := 0; i < 998; i++ {
		q.Push(Int(rand.Intn(100)))
	}
	q.Push(Int(5201314))

	s := new(IntSorter)
	n := 1000
	for q.Len() > 0 {
		if q.Len() != n {
			t.Fatal()
		}
		n--

		x := q.Top()
		y := q.Pop()
		if x != y || x.(Int) != y.(Int) {
			t.Fatal()
		}

		*s = append(*s, x.(Int))
	}

	if (*s)[0] != -1 || (*s)[999] != 5201314 {
		t.Fatal()
	}

	if !sort.IsSorted(s) {
		t.Fatal()
	}

	q = New()

	if q.Len() != 0 {
		t.Fatal()
	}
}

func TestFixAndRemove(t *testing.T) {
	q := New()
	q.Push(Int(1))
	q.Push(Int(3))
	q.Push(Int(2))
	q.Push(Int(4))

	if (*q.s)[0].(Int) != 1 || (*q.s)[1].(Int) != 3 || (*q.s)[2].(Int) != 2 || (*q.s)[3].(Int) != 4 {
		t.Fatal()
	}

	q.Fix(Int(5), 1)

	if (*q.s)[0].(Int) != 1 || (*q.s)[1].(Int) != 4 || (*q.s)[2].(Int) != 2 || (*q.s)[3].(Int) != 5 {
		t.Fatal()
	}

	a, b, c, d := q.Pop(), q.Pop(), q.Pop(), q.Pop()
	if a.(Int) != 1 || b.(Int) != 2 || c.(Int) != 4 || d.(Int) != 5 {
		t.Fatal()
	}

	if q.Len() != 0 {
		t.Fatal()
	}

	q.Push(Int(8))
	q.Push(Int(6))
	q.Push(Int(7))
	q.Push(Int(9))

	// println((*q.s)[0].(Int), (*q.s)[1].(Int), (*q.s)[2].(Int), (*q.s)[3].(Int))

	if (*q.s)[0].(Int) != 6 || (*q.s)[1].(Int) != 8 || (*q.s)[2].(Int) != 7 || (*q.s)[3].(Int) != 9 {
		t.Fatal()
	}

	if q.Top().(Int) != 6 {
		t.Fatal()
	}

	q.Remove(0)

	if q.Top().(Int) != 7 {
		t.Fatal()
	}

	a, b, c = q.Pop(), q.Pop(), q.Pop()
	if a.(Int) != 7 || b.(Int) != 8 || c.(Int) != 9 {
		t.Fatal()
	}
}

type Node struct {
	priority int
	value    int
}

func NewNode(p, v int) *Node {
	return &Node{priority: p}
}

func (this *Node) Less(other interface{}) bool {
	return this.priority < other.(*Node).priority
}

func TestStruct(t *testing.T) {
	q := New()
	for i := 0; i < 1000; i++ {
		q.Push(NewNode(rand.Intn(100000), i))
	}

	n := 1000
	for q.Len() > 0 {
		if q.Len() != n {
			t.Fatal()
		}
		n--

		x := q.Top().(*Node)
		y := q.Pop().(*Node)

		if x.priority != y.priority || x.value != y.value {
			t.Fatal()
		}
	}
}

func BenchmarkPush1(b *testing.B) {
	b.StopTimer()
	q := New()
	b.StartTimer()
	for i := 0; i < 100000; i++ {
		q.Push(NewNode(rand.Intn(100), i))
	}
}

func BenchmarkPush2(b *testing.B) {
	b.StopTimer()
	q := New()
	b.StartTimer()
	for i := 0; i < 500000; i++ {
		q.Push(NewNode(rand.Intn(100), i))
	}
}

func BenchmarkPush3(b *testing.B) {
	b.StopTimer()
	q := New()
	b.StartTimer()
	for i := 0; i < 1000000; i++ {
		q.Push(NewNode(rand.Intn(100), i))
	}
}

func BenchmarkPop1(b *testing.B) {
	b.StopTimer()
	q := New()
	for i := 0; i < 100000; i++ {
		q.Push(NewNode(rand.Intn(100), i))
	}

	b.StartTimer()
	for q.Len() > 0 {
		q.Pop()
	}
}

func BenchmarkPop2(b *testing.B) {
	b.StopTimer()
	q := New()
	for i := 0; i < 500000; i++ {
		q.Push(NewNode(rand.Intn(100), i))
	}

	b.StartTimer()
	for q.Len() > 0 {
		q.Pop()
	}
}

func BenchmarkPop3(b *testing.B) {
	b.StopTimer()
	q := New()
	for i := 0; i < 1000000; i++ {
		q.Push(NewNode(rand.Intn(100), i))
	}

	b.StartTimer()
	for q.Len() > 0 {
		q.Pop()
	}
}
