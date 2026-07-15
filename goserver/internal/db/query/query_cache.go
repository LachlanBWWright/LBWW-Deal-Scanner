package query

import (
	"sync"

	"dealscanner/internal/models"
)

type savedQueryCache struct {
	mu      sync.RWMutex
	queries map[string][]QueryItem
}

var defaultSavedQueryCache = &savedQueryCache{
	queries: make(map[string][]QueryItem),
}

type globalsCache struct {
	mu      sync.RWMutex
	globals *models.Globals
}

var defaultGlobalsCache = &globalsCache{}

func (c *savedQueryCache) get(queryType string) ([]QueryItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	queries, ok := c.queries[queryType]
	if !ok {
		return nil, false
	}
	return cloneQueryItems(queries), true
}

func (c *savedQueryCache) set(queryType string, queries []QueryItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queries[queryType] = cloneQueryItems(queries)
}

func (c *savedQueryCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queries = make(map[string][]QueryItem)
}

func cloneQueryItems(items []QueryItem) []QueryItem {
	if items == nil {
		return []QueryItem{}
	}

	cloned := make([]QueryItem, len(items))
	for i, item := range items {
		cloned[i] = cloneQueryItem(item)
	}
	return cloned
}

func cloneQueryItem(item QueryItem) QueryItem {
	cloned := item
	cloned.Url = cloneStringPtr(item.Url)
	cloned.Name = cloneStringPtr(item.Name)
	cloned.DisplayUrl = cloneStringPtr(item.DisplayUrl)
	cloned.MaxPrice = cloneFloat64Ptr(item.MaxPrice)
	cloned.LastPrice = cloneFloat64Ptr(item.LastPrice)
	cloned.MinPrice = cloneFloat64Ptr(item.MinPrice)
	cloned.MinFloat = cloneFloat64Ptr(item.MinFloat)
	cloned.MaxFloat = cloneFloat64Ptr(item.MaxFloat)
	cloned.RequiredPhrases = cloneStringPtr(item.RequiredPhrases)
	cloned.RequiredMatchMode = cloneStringPtr(item.RequiredMatchMode)
	cloned.ExcludePhrases = cloneStringPtr(item.ExcludePhrases)
	cloned.ExcludeMatchMode = cloneStringPtr(item.ExcludeMatchMode)
	cloned.RequiredInDescription = cloneStringPtr(item.RequiredInDescription)
	cloned.ExcludeInDescription = cloneStringPtr(item.ExcludeInDescription)
	cloned.ScanMode = cloneStringPtr(item.ScanMode)
	return cloned
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneFloat64Ptr(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func (c *globalsCache) get() (*models.Globals, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.globals == nil {
		return nil, false
	}
	return cloneGlobals(c.globals), true
}

func (c *globalsCache) set(globals *models.Globals) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.globals = cloneGlobals(globals)
}

func (c *globalsCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.globals = nil
}

func cloneGlobals(globals *models.Globals) *models.Globals {
	if globals == nil {
		return nil
	}
	cloned := *globals
	return &cloned
}
