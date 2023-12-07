package memdb

import (
	"gRedis/util"
	"sort"
	"sync"
)

// LocksManager apply to ensure some atomic operations
type LocksManager struct {
	locks []*sync.RWMutex
}

func NewLocksManager(size int) *LocksManager {
	m := &LocksManager{}
	for i := 0; i < size; i++ {
		m.locks[i] = &sync.RWMutex{}
	}
	return m
}

func (m *LocksManager) GetKeyPos(key string) int {
	pos := util.Hash(key) // 可能比MaxSegSize大
	return pos % len(m.locks)
}

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

func (m *LocksManager) LockKeys(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].Lock()
	}
}

func (m *LocksManager) UnLockKeys(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].Unlock()
	}
}

func (m *LocksManager) RLockKeys(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].RLock()
	}
}

func (m *LocksManager) RUnLockKeys(keys []string) {
	order := m.getSortedLocks(keys)
	for i := range order {
		pos := order[i]
		m.locks[pos].RUnlock()
	}
}
