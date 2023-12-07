package memdb

import (
	"gRedis/logger"
	"gRedis/protocol"
	"time"
)

type MemDb struct {
	dict    *ConcurrentMap // memory cache db
	expires *ConcurrentMap // keys with expire time(seconds)
	locks   *LocksManager
}

func NewMemDb() *MemDb {
	return &MemDb{
		dict:    NewConcurrentMap(MaxSegSize),
		expires: NewConcurrentMap(MaxSegSize),
		locks:   NewLocksManager(2 * MaxSegSize),
	}
}

func (db *MemDb) ExecCommand(cmd [][]byte) protocol.RedisData {
	cmdName := string(cmd[0])
	if command, ok := CmdTable[cmdName]; ok {
		return command.executor(db, cmd)
	}
	return protocol.NewSimpleError("ERROR: cmd unsupported type")
}

// return true if expired
func (db *MemDb) CheckExpire(key string) bool {
	_expireTime, ok := db.expires.Get(key)

	// key not with expire time
	if !ok {
		return true
	}

	// key is expired
	expireTime := _expireTime.(int64)
	now := time.Now().Unix()

	return now > expireTime
}

func (db *MemDb) SetExpire(key string, expire int64) int {
	if _, ok := db.dict.Get(key); !ok {
		logger.Error("SetExpire: key doesn't exist")
		return 0
	}
	db.expires.Set(key, expire)
	return 1
}

func (db *MemDb) DeleteExpire(key string) int {
	return db.expires.Delete(key)
}

func (db *MemDb) DeleteExpireKey(key string) {
	db.locks.Lock(key)
	defer db.locks.UnLock(key)
	db.dict.Delete(key)
	db.expires.Delete(key)
}
