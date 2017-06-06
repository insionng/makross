非线程安全的优先队列（golang标准库中的heap是小根堆）

~~~ go
package main

import (
	"fmt"
	"pueue"
)

type Node struct {
	priority int
	value    int
}

func (this *Node) Less(other interface{}) bool {
	return this.priority < other.(*Node).priority
}

func main() {
	q := priority_queue.New()

	q.Push(&Node{priority: 8, value: 1})
	q.Push(&Node{priority: 7, value: 2})
	q.Push(&Node{priority: 9, value: 3})

	x := q.Top().(*Node)
	fmt.Println(x.priority, x.value)

	for q.Len() > 0 {
		x = q.Pop().(*Node)
		fmt.Println(x.priority, x.value)
	}

	// output:
	// 7 2

	// 7 2
	// 8 1
	// 9 3
}

~~~


##LICENSE

MIT