package query

import "sync"

type savedQueryCache struct {
	mu      sync.RWMutex
	queries map[string][]QueryItem
}

var defaultSavedQueryCache = &savedQueryCache{
	queries: make(map[string][]QueryItem),
}

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
	cloned.MinPrice = cloneFloat64Ptr(item.MinPrice)
	cloned.MinFloat = cloneFloat64Ptr(item.MinFloat)
	cloned.MaxFloat = cloneFloat64Ptr(item.MaxFloat)
	cloned.RequiredPhrases = cloneStringPtr(item.RequiredPhrases)
	cloned.ExcludePhrases = cloneStringPtr(item.ExcludePhrases)
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
