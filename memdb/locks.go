package memdb

import (
	"fmt"
	"gRedis/util"
	"sort"
	"sync"
)

// LocksManager apply to ensure some atomic operations
type LocksManager struct {
	locks []*sync.RWMutex
}

func NewLocksManager(size int) *LocksManager {
	locks := make([]*sync.RWMutex, size)
	for i := 0; i < size; i++ {
		locks[i] = &sync.RWMutex{}
	}
	return &LocksManager{locks: locks}
}

func (m *LocksManager) GetKeyPos(key string) int {
	key = fmt.Sprintf("@%s&", key)
	pos := util.Hash(key) // 可能比MaxSegSize大
	return pos % len(m.locks)
}

// 即使映射到同一pos，也是前一个锁释放了，后一个才结束阻塞并且上锁，保证安全性。
func (m *LocksManager) Lock(key string) {
	pos := m.GetKeyPos(key)
	m.locks[pos].Lock()
}

func (m *LocksManager) UnLock(key string) {
	pos := m.GetKeyPos(key)
	m.locks[pos].Unlock()
}

func (m *LocksManager) RLock(key string) {
	pos := m.GetKeyPos(key)
	m.locks[pos].RLock()
}

func (m *LocksManager) RUnLock(key string) {
	pos := m.GetKeyPos(key)
	m.locks[pos].RUnlock()
}

// force to lock/unlock in the same order, to avoid dead lock
func (m *LocksManager) getSortedLocks(keys []string) []int {
	set := make(map[int]struct{})
	for i := range keys {
		pos := m.GetKeyPos(keys[i])
		set[pos] = struct{}{}
	}

	order := make([]int, len(set))
	i := 0
	for pos := range set {
		order[i] = pos
		i++
	}

	sort.Ints(order)
	return order
}

func (m *LocksManager) MLock(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].Lock()
	}
}

func (m *LocksManager) MUnLock(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].Unlock()
	}
}

func (m *LocksManager) MRLock(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].RLock()
	}
}

func (m *LocksManager) MRUnLock(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].RUnlock()
	}
}
