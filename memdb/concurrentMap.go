package memdb

import (
	"gRedis/util"
	"sync"
)

const MaxSegSize int = int(1<<32 - 1) // max signed int32

type ConcurrentMap struct {
	table []*segmentation
	size  int // table size
	count int // key counts
}

type segmentation struct {
	ht   map[string]any // hash table
	rwMu *sync.RWMutex  //  The lock can be held by an arbitrary number of readers or a single writer.
}

func NewConcurrentMap(size int) *ConcurrentMap {
	if size <= 0 || size >= MaxSegSize {
		size = MaxSegSize
	}

	m := &ConcurrentMap{
		table: make([]*segmentation, size),
		size:  size,
		count: 0,
	}

	for i := 0; i < size; i++ {
		m.table[i] = &segmentation{ht: make(map[string]any), rwMu: &sync.RWMutex{}}
	}

	return m
}

func (m *ConcurrentMap) getKeyPos(key string) int {
	hash := util.Hash(key)
	return hash % m.size
}

// 设置int输出而不是bool 是为了记录新增key数量
func (m *ConcurrentMap) Set(key string, value any) int {
	added := 0
	pos := m.getKeyPos(key)

	segmentation := m.table[pos]
	segmentation.rwMu.Lock()
	defer segmentation.rwMu.Unlock()

	if _, ok := segmentation.ht[key]; !ok {
		added = 1
		m.count++
	}
	segmentation.ht[key] = value
	return added
}

func (m *ConcurrentMap) Delete(key string) int {
	pos := m.getKeyPos(key)

	segmentation := m.table[pos]
	segmentation.rwMu.Lock()
	defer segmentation.rwMu.Unlock()

	if _, ok := segmentation.ht[key]; ok {
		delete(segmentation.ht, key)
		m.count--
		return 1
	}

	return 0
}

func (m *ConcurrentMap) Get(key string) (any, bool) {
	pos := m.getKeyPos(key)
	segmentation := m.table[pos]
	segmentation.rwMu.RLock()
	defer segmentation.rwMu.RUnlock()

	value, ok := segmentation.ht[key]
	return value, ok
}

func (m *ConcurrentMap) Size() int {
	return m.size
}

func (m *ConcurrentMap) Count() int {
	return m.count
}

func (m *ConcurrentMap) Clear() {
	*m = *NewConcurrentMap(m.size) // 这里改的是指针
}

// 这里拿到的keys有可能有过期的，需要lazy deletion
func (m *ConcurrentMap) Keys() []string {
	keys := make([]string, m.count)
	k := 0
	for i := range m.table {
		m.table[i].rwMu.Lock()
		for key := range m.table[i].ht {
			keys[k] = key
			k++
		}
		m.table[i].rwMu.Unlock()
	}
	return keys
}
