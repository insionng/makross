package makross

import (
	"pueue"
)

var (
	ops uint64 = 0
)

type Node struct {
	priority uint64
	key      string
	callback func([]byte) []byte
}

func (n *Node) Less(other interface{}) bool {
	return n.priority < other.(*Node).priority
}

func (c *Context) NewPriorityQueue() *pueue.PriorityQueue {
	return pueue.New()
}

func (c *Context) NewPriorityQueuesMap() map[string]*pueue.PriorityQueue {
	return map[string]*pueue.PriorityQueue{}
}

func (c *Context) RemoveFilterHook(key string) {
	if c.HasFilterHook(key) {
		delete(c.FiltersMap, key)
	}
}

func (c *Context) RemoveActionHook(key string) {
	if c.HasActionHook(key) {
		delete(c.makross.QueuesMap, key)
	}
}

func (c *Context) RemoveActionsHook() {
	c.makross.QueuesMap = nil
}

func (c *Context) HasFilterHook(key string) bool {
	if _, okay := c.FiltersMap[key]; okay {
		return true
	}
	return false
}

func (c *Context) HasActionHook(key string) bool {
	if _, okay := c.makross.QueuesMap[key]; okay {
		return true
	}
	return false
}

func (c *Context) AddActionHook(key string, function func(), priorities ...uint64) {
	c.AddFilterHook(key, func([]byte) []byte {
		function()
		return nil
	}, priorities...)
}

func (c *Context) AddFilterHook(key string, function func([]byte) []byte, priorities ...uint64) {
	if !c.HasActionHook(key) {
		if c.makross.QueuesMap == nil {
			c.makross.QueuesMap = c.NewPriorityQueuesMap()
		}
		c.makross.QueuesMap[key] = c.NewPriorityQueue()
	}

	var priority uint64
	if len(priorities) > 0 {
		priority = priorities[0]
	} else {
		//atomic.AddUint64(&ops, 1)
		priority = ops
	}

	c.makross.QueuesMap[key].Push(&Node{priority: priority, key: key, callback: function})
}

func (c *Context) CurrentFilter() string {
	return c.CurrentFilterKey
}

func (c *Context) DoActionHook(key string) {
	c.DoFilterHook(key, nil)
}

func (c *Context) DoFilterHook(key string, function func() []byte) []byte {
	if !c.HasActionHook(key) {
		c.AddFilterHook(key, func([]byte) []byte {
			return function()
		})
		return c.DoFilterHook(key, function)
	}

	if !c.HasFilterHook(key) {
		if c.FiltersMap == nil {
			c.FiltersMap = make(map[string][]byte)
		}
	}

	c.CurrentFilterKey = key
	if c.HasActionHook(key) {
		for c.makross.QueuesMap[key].Len() > 0 {
			n, okay := c.makross.QueuesMap[key].Pop().(*Node)
			if !okay {
				continue
			}

			if function != nil { //for Filter
				if c.FiltersMap[key] != nil {
					c.FiltersMap[key] = n.callback(c.FiltersMap[key])
				} else {
					c.FiltersMap[key] = n.callback(function())
				}
			} else { //for Action
				c.FiltersMap[key] = n.callback(nil)
			}
		}
	}
	return c.FiltersMap[key]
}
