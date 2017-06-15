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

func (m *Makross) NewPriorityQueue() *pueue.PriorityQueue {
	return pueue.New()
}

func (c *Context) NewPriorityQueue() *pueue.PriorityQueue {
	return c.makross.NewPriorityQueue()
}

func (m *Makross) NewPriorityQueuesMap() map[string]*pueue.PriorityQueue {
	return map[string]*pueue.PriorityQueue{}
}

func (c *Context) NewPriorityQueuesMap() map[string]*pueue.PriorityQueue {
	return c.makross.NewPriorityQueuesMap()
}

func (m *Makross) RemoveFilterHook(key string) {
	if m.HasFilterHook(key) {
		delete(m.FiltersMap, key)
	}
}

func (c *Context) RemoveFilterHook(key string, globals ...bool) {
	var global = false
	if len(globals) > 0 {
		global = globals[0]
	}
	if global {
		c.makross.RemoveFilterHook(key)
	} else if c.HasFilterHook(key) {
		delete(c.FiltersMap, key)
	}
}

func (m *Makross) RemoveActionHook(key string) {
	if m.HasActionHook(key) {
		delete(m.QueuesMap, key)
		delete(m.FiltersMap, key)
	}
}

func (c *Context) RemoveActionHook(key string, globals ...bool) {
	var global = false
	if len(globals) > 0 {
		global = globals[0]
	}
	if global {
		c.makross.RemoveActionHook(key)
	} else if c.HasActionHook(key) {
		delete(c.makross.QueuesMap, key)
		delete(c.FiltersMap, key)
	}
}

func (m *Makross) RemoveActionsHook() {
	m.QueuesMap = nil
}

func (c *Context) RemoveActionsHook() {
	c.makross.RemoveActionsHook()
}

func (m *Makross) HasFilterHook(key string) bool {
	if _, okay := m.FiltersMap[key]; okay {
		return true
	}
	return false
}

func (c *Context) HasFilterHook(key string, globals ...bool) bool {
	var global = false
	if len(globals) > 0 {
		global = globals[0]
	}
	if global {
		return c.makross.HasFilterHook(key)
	} else if _, okay := c.FiltersMap[key]; okay {
		return true
	}
	return false
}

func (m *Makross) HasActionHook(key string) bool {
	if _, okay := m.QueuesMap[key]; okay {
		if _, okay := m.QueuesMap[key].Top().(*Node); okay {
			return true
		}
	}
	return false
}

func (c *Context) HasActionHook(key string) bool {
	return c.makross.HasActionHook(key)
}

func (m *Makross) AddActionHook(key string, function func(), priorities ...uint64) {
	m.AddFilterHook(key, func([]byte) []byte {
		function()
		return nil
	}, priorities...)
}

func (c *Context) AddActionHook(key string, function func(), priorities ...uint64) {
	c.AddFilterHook(key, func([]byte) []byte {
		function()
		return nil
	}, priorities...)
}

func (m *Makross) AddFilterHook(key string, function func([]byte) []byte, priorities ...uint64) {
	if !m.HasActionHook(key) {
		if m.QueuesMap == nil {
			m.QueuesMap = m.NewPriorityQueuesMap()
		}
		m.QueuesMap[key] = m.NewPriorityQueue()
	}

	var priority uint64
	if len(priorities) > 0 {
		priority = priorities[0]
	} else {
		//atomic.AddUint64(&ops, 1)
		priority = ops
	}

	m.QueuesMap[key].Push(&Node{priority: priority, key: key, callback: function})
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

// DoActionHook Global DoActionHook
func (m *Makross) DoActionHook(key string) {
	m.DoFilterHook(key, nil)
}

func (c *Context) DoActionHook(key string, globals ...bool) {
	c.DoFilterHook(key, nil, globals...)
}

// DoFilterHook Global DoFilterHook
func (m *Makross) DoFilterHook(key string, function func() []byte) []byte {

	if !m.HasActionHook(key) {
		m.AddFilterHook(key, func(b []byte) []byte {
			if function == nil {
				return b
			}
			return function()
		})
		return m.DoFilterHook(key, function)
	}

	if !m.HasFilterHook(key) {
		if m.FiltersMap == nil {
			m.FiltersMap = make(map[string][]byte)
		}
	}

	if m.HasActionHook(key) {
		for m.QueuesMap[key].Len() > 0 {
			n, okay := m.QueuesMap[key].Pop().(*Node)
			if !okay {
				continue
			}

			if function != nil { //for Global Filter
				if m.FiltersMap[key] != nil {
					m.FiltersMap[key] = n.callback(m.FiltersMap[key])
				} else {
					m.FiltersMap[key] = n.callback(function())
				}
			} else { //for Global Action
				m.FiltersMap[key] = n.callback(nil)
			}
		}
	}

	return m.FiltersMap[key]

}

func (c *Context) DoFilterHook(key string, function func() []byte, globals ...bool) []byte {

	var global = false
	if len(globals) > 0 {
		global = globals[0]
	}

	var filterBytes []byte

	if !c.HasActionHook(key) {
		c.AddFilterHook(key, func(b []byte) []byte {
			if function == nil {
				return b
			}
			return function()
		})
		return c.DoFilterHook(key, function, globals...)
	}

	if global {
		if !c.makross.HasFilterHook(key) {
			if c.makross.FiltersMap == nil {
				c.makross.FiltersMap = make(map[string][]byte)
			}
		}
	} else {
		if !c.HasFilterHook(key) {
			if c.FiltersMap == nil {
				c.FiltersMap = make(map[string][]byte)
			}
		}
	}

	if c.HasActionHook(key) {
		for c.makross.QueuesMap[key].Len() > 0 {
			n, okay := c.makross.QueuesMap[key].Pop().(*Node)
			if !okay {
				continue
			}
			if global {
				if function != nil { //for Global Filter
					if c.makross.FiltersMap[key] != nil {
						c.makross.FiltersMap[key] = n.callback(c.makross.FiltersMap[key])
					} else {
						c.makross.FiltersMap[key] = n.callback(function())
					}
				} else { //for Global Action
					c.makross.FiltersMap[key] = n.callback(nil)
				}
				filterBytes = c.makross.FiltersMap[key]
			} else {
				if function != nil { //for Filter
					if c.FiltersMap[key] != nil {
						c.FiltersMap[key] = n.callback(c.FiltersMap[key])
					} else {
						c.FiltersMap[key] = n.callback(function())
					}
				} else { //for Action
					c.FiltersMap[key] = n.callback(nil)
				}
				filterBytes = c.FiltersMap[key]
			}

		}
	}

	return filterBytes

}
