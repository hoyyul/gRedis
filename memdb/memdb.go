package memdb

import (
	"fmt"
	"gRedis/config"
	"gRedis/logger"
	"gRedis/resp"
	"strings"
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

func (db *MemDb) ExecCommand(cmd [][]byte) resp.RedisData {
	cmdName := strings.ToLower(string((cmd[0])))
	if command, ok := CmdTable[cmdName]; ok {
		return command.executor(db, cmd)
	}
	return resp.NewSimpleError(fmt.Sprintf("unknown command '%s'", cmdName))
}

// return true if expired
func (db *MemDb) CheckExpire(key string) bool {
	_expireTime, ok := db.expires.Get(key)

	// key is persistent
	if !ok {
		return false
	}

	// key is expired
	expireTime := _expireTime.(int64)
	now := time.Now().Unix()

	return now > expireTime
}

func (db *MemDb) SetExpire(key string, ttl int64) int {
	if _, ok := db.dict.Get(key); !ok {
		logger.Error("SetExpire: key doesn't exist")
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
