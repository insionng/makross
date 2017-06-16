package pueue

import (
	"container/heap"
	"sync"
)

var mutex sync.RWMutex

type Interface interface {
	Less(other interface{}) bool
}

type sorter []Interface

// Implement heap.Interface: Push, Pop, Len, Less, Swap
func (s *sorter) Push(x interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	*s = append(*s, x.(Interface))
}

func (s *sorter) Pop() interface{} {
	mutex.RLock()
	defer mutex.RUnlock()

	n := len(*s)
	if n > 0 {
		x := (*s)[n-1]
		*s = (*s)[0 : n-1]
		return x
	}
	return nil
}

func (s *sorter) Len() int {
	mutex.RLock()
	defer mutex.RUnlock()
	return len(*s)
}

func (s *sorter) Less(i, j int) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return (*s)[i].Less((*s)[j])
}

func (s *sorter) Swap(i, j int) {
	mutex.Lock()
	defer mutex.Unlock()
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

// Define priority queue struct
type PriorityQueue struct {
	s     *sorter
	mutex sync.RWMutex
}

func New() *PriorityQueue {
	q := &PriorityQueue{s: new(sorter)}
	q.mutex.Lock()
	heap.Init(q.s)
	q.mutex.Unlock()
	return q
}

func (q *PriorityQueue) Push(x Interface) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	heap.Push(q.s, x)
}

func (q *PriorityQueue) Pop() Interface {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return heap.Pop(q.s).(Interface)
}

func (q *PriorityQueue) Top() Interface {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	if len(*q.s) > 0 {
		return (*q.s)[0].(Interface)
	}
	return nil
}

func (q *PriorityQueue) Fix(x Interface, i int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	(*q.s)[i] = x
	heap.Fix(q.s, i)
}

func (q *PriorityQueue) Remove(i int) Interface {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return heap.Remove(q.s, i).(Interface)
}

func (q *PriorityQueue) Len() int {
	q.mutex.RLock()
	l := q.s.Len()
	q.mutex.RUnlock()
	return l
}
