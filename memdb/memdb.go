package memdb

import (
	"gRedis/config"
	"time"
)

type MemDb struct {
	dict    *ConcurrentMap // memory cache db
	expires *ConcurrentMap // keys with expire time(seconds)
	locks   *LocksManager
}

func NewMemDb() *MemDb {
	return &MemDb{
		dict:    NewConcurrentMap(config.Conf.SegNum),
		expires: NewConcurrentMap(config.Conf.SegNum),
		locks:   NewLocksManager(2 * config.Conf.SegNum),
	}
}

// return true if expired
func (db *MemDb) CheckExpire(key string) bool {
	_expireTime, ok := db.expires.Get(key)

	// key is persistent or not in dict table
	if !ok {
		return false
	}

	expireTime := _expireTime.(int64)
	now := time.Now().Unix()

	// true if expired; false if key not expired (in both expire table and dict)
	return now > expireTime
}

// this guarantee if key can be fount in expire table,
// it must to be in dict table
func (db *MemDb) SetExpire(key string, ttl int64) int {
	if _, ok := db.dict.Get(key); !ok {
		//logger.Error("SetExpire: key doesn't exist")
		return 0
	}
	db.expires.Set(key, ttl)
	return 1
}

func (db *MemDb) DeleteExpire(key string) int {
	return db.expires.Delete(key)
}

// lazy deletion
func (db *MemDb) DeleteExpiredKey(key string) bool {
	if db.CheckExpire(key) {
		db.locks.Lock(key)
		defer db.locks.UnLock(key)
		db.dict.Delete(key)
		db.expires.Delete(key)
		return true
	}
	return false
}
