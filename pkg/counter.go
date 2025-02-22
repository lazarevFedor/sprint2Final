package pkg

import "sync"

type Counter struct {
	value int // значение счетчика
	mu    sync.RWMutex
}

func (c *Counter) Increment() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *Counter) GetValue() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func (c *Counter) Decrement() {
	c.mu.Lock()
	c.value--
	c.mu.Unlock()
}

func (c *Counter) SetValue(value int) {
	c.mu.Lock()
	c.value = value
	c.mu.Unlock()
}
