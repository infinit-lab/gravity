package model

import (
	"sort"
	"sync"
)

type BeforeInsertLayer func(id int, resource interface{})
type BeforeEraseLayer func(id int, resource interface{})

type Cache interface {
	Insert(id int, resource interface{}, layer BeforeInsertLayer)
	Erase(id int, layer BeforeEraseLayer)
	Get(id int) (resource interface{}, ok bool)
	GetAll() (resources []interface{})
	Clear()
}

func NewCache() Cache {
	c := new(cache)
	c.cache = make(map[int]interface{})
	return c
}

type cache struct {
	cache map[int]interface{}
	mutex sync.Mutex
}

func (c *cache) Insert(id int, resource interface{}, layer BeforeInsertLayer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if layer != nil {
		layer(id, resource)
	}
	c.cache[id] = resource
}

func (c *cache) Erase(id int, layer BeforeEraseLayer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	resource, ok := c.cache[id]
	if !ok {
		return
	}
	if layer != nil {
		layer(id, resource)
	}
	delete(c.cache, id)
}

func (c *cache) Get(id int) (resource interface{}, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	resource, ok = c.cache[id]
	return
}

func (c *cache) GetAll() (resources []interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var keys []int
	for key := range c.cache {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		resources = append(resources, c.cache[key])
	}
	return
}

func (c *cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[int]interface{})
}
