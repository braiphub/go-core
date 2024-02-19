package braipfilter

import "sync"

type filterCache struct {
	sync.RWMutex
	configs map[string][]filterConfig
}

func newCache() *filterCache {
	return &filterCache{
		RWMutex: sync.RWMutex{},
		configs: make(map[string][]filterConfig),
	}
}

func (c *filterCache) set(name string, val []filterConfig) {
	c.Lock()
	defer c.Unlock()

	c.configs[name] = val
}

func (c *filterCache) get(name string) []filterConfig {
	c.RLock()
	defer c.RUnlock()

	if val, ok := c.configs[name]; ok {
		return val
	}

	return nil
}
