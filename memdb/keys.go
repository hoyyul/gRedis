package memdb

import (
	"fmt"
	"gRedis/resp"
	"gRedis/util"
	"strconv"
	"strings"
	"time"
)

func pingKeys(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) > 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}
	if len(cmd) == 1 {
		return resp.NewSimpleString("PONG")
	}
	return resp.NewBulkString(cmd[1])
}

func delKey(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	deleted := 0
	for _, key := range cmd[1:] {
		if !db.DeleteExpiredKey(string(key)) {
			db.locks.Lock(string(key))
			deleted += db.dict.Delete(string(key))
			db.DeleteExpire(string(key))
			db.locks.UnLock(string(key))
		}
	}

	return resp.NewInteger(int64(deleted))
}

func existsKey(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	existed := 0
	for _, key := range cmd[1:] {
		if !db.DeleteExpiredKey(string(key)) {
			db.locks.RLock(string(key))
			if _, ok := db.dict.Get(string(key)); ok {
				existed++
			}
			db.locks.RUnLock(string(key))
		}
	}

	return resp.NewInteger(int64(existed))
}

func keysKey(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	pattern := string(cmd[1])
	res := make([]resp.RedisData, 0)
	keys := db.dict.Keys()
	for _, key := range keys {
		if !db.DeleteExpiredKey(key) {
			if util.PattenMatch(pattern, key) {
				res = append(res, resp.NewBulkString([]byte(key)))
			}
		}
	}

	return resp.NewArray(res)
}

/*
NX -- Set expiry only when the key has no expiry
XX -- Set expiry only when the key has an existing expiry
GT -- Set expiry only when the new expiry is greater than current one
LT -- Set expiry only when the new expiry is less than current one
*/
func expireKey(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 || len(cmd) > 4 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}
	var res int
	// set ttl
	v, err := strconv.ParseInt(string(cmd[2]), 10, 64)
	if err != nil {
		return resp.NewSimpleError("value is not an integer or out of range")
	}
	ttl := time.Now().Unix() + v

	// get option
	var option string
	if len(cmd) == 4 {
		option = strings.ToLower(string(cmd[3]))
	}

	// get key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(int64(0))
	}
	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	switch option {
	case "nx":
		if _, ok := db.expires.Get(key); !ok {
			res = db.SetExpire(key, ttl)
		}
	case "xx":
		if _, ok := db.expires.Get(key); ok {
			res = db.SetExpire(key, ttl)
		}
	case "gt":
		if v, ok := db.expires.Get(key); ok && ttl > v.(int64) {
			res = db.SetExpire(key, ttl)
		}
	case "lt":
		if v, ok := db.expires.Get(key); ok && ttl < v.(int64) {
			res = db.SetExpire(key, ttl)
		}
	default:
		if option != "" {
			return resp.NewSimpleError(fmt.Sprintf("Unsupported option %s", string(cmd[3])))
		}
		res = db.SetExpire(key, ttl)
	}
	return resp.NewInteger(int64(res))
}

func persistKey(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(int64(0))
	}
	db.locks.Lock(string(key))
	defer db.locks.UnLock(string(key))

	db.DeleteExpire(key)

	return resp.NewInteger(int64(1))
}

func ttlKey(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(int64(-2))
	}
	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	if _, ok := db.dict.Get(key); !ok {
		return resp.NewInteger(int64(-2))
	}

	now := time.Now().Unix()
	ttl, ok := db.expires.Get(key)
	if !ok {
		return resp.NewInteger(int64(-1))
	}

	return resp.NewInteger(ttl.(int64) - now)
}

func renameKey(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	newKey := string(cmd[2])
	oldKey := string(cmd[1])
	if db.DeleteExpiredKey(oldKey) {
		return resp.NewSimpleError("no such key")
	}

	db.locks.Lock(oldKey)
	defer db.locks.UnLock(oldKey)

	oldValue, ok := db.dict.Get(oldKey)
	if !ok {
		return resp.NewSimpleError("no such key")
	}

	oldTTL, ok := db.expires.Get(oldKey)

	db.dict.Delete(oldKey)
	db.expires.Delete(oldKey)
	// If newkey already exists it is overwritten
	db.dict.Delete(newKey)
	db.expires.Delete(newKey)
	db.dict.Set(newKey, oldValue)

	// If a key is renamed with RENAME, the associated time to live is transferred to the new key name.
	if ok {
		db.expires.Set(newKey, oldTTL)
	}

	return resp.NewSimpleString("OK")
}

func RegisterKeyCommands() {
	RegisterCommand("ping", pingKeys)
	RegisterCommand("del", delKey)
	RegisterCommand("exists", existsKey)
	RegisterCommand("keys", keysKey)
	RegisterCommand("expire", expireKey)
	RegisterCommand("persist", persistKey)
	RegisterCommand("ttl", ttlKey)
	RegisterCommand("rename", renameKey)
}
