package collection

import "sync"

type Collection struct {
	RWMutex sync.RWMutex
	Data    map[string]string
}

func (c *Collection) has(key string) bool {
	if _, ok := c.Data[key]; ok {
		return true
	}

	return false
}

func New() *Collection {
	return &Collection{
		Data:    make(map[string]string),
		RWMutex: sync.RWMutex{},
	}
}

func (c *Collection) Has(key string) bool {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()

	return c.has(key)
}

func (c *Collection) Add(key, value string) *Collection {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if c.has(key) {
		return c
	}

	c.Data[key] = value

	return c
}

func (c *Collection) Delete(key string) bool {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if !c.has(key) {
		return false
	}

	delete(c.Data, key)

	return true
}

func (c *Collection) Get(key string) (string, bool) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()

	if !c.has(key) {
		return "", false
	}

	value, ok := c.Data[key]

	return value, ok
}
