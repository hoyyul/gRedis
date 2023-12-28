package memdb

import (
	"gRedis/resp"
	"strconv"
	"strings"
)

func hDelHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(0)
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	// if hash table is empty, delete it
	defer func() {
		if h.IsEmpty() {
			db.dict.Delete(key)
			db.DeleteExpire(key)
		}
	}()

	res := 0
	for _, _field := range cmd[2:] {
		field := string(_field)
		res += h.Del(field)
	}

	return resp.NewInteger(int64(res))
}

func hExistsHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(0)
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	field := string(cmd[2])
	if h.Exist(field) {
		return resp.NewInteger(int64(1))
	}

	return resp.NewInteger(int64(0))
}

func hGetHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewBulkString(nil)
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	field := string(cmd[2])
	val := h.Get(field)

	return resp.NewBulkString(val)
}

func hGetAllHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	res := make([]resp.RedisData, 0, h.Len())
	for k, v := range h.table {
		res = append(res, resp.NewBulkString([]byte(k)))
		res = append(res, resp.NewBulkString(v))
	}

	return resp.NewArray(res)
}

func hIncrByHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// If key does not exist, a new key holding a hash is created.
	v, ok := db.dict.Get(key)
	if !ok {
		db.dict.Set(key, NewHash())
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	field := string(cmd[2])
	increment, err := strconv.Atoi(string(cmd[3]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	val, ok := h.IncrBy(field, increment)
	if !ok {
		return resp.NewSimpleError("hash value is not an integer")
	}

	return resp.NewInteger(int64(val))
}

func hIncrByFloatHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// If key does not exist, a new key holding a hash is created.
	v, ok := db.dict.Get(key)
	if !ok {
		db.dict.Set(key, NewHash())
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	field := string(cmd[2])
	increment, err := strconv.ParseFloat(string(cmd[3]), 64)
	if err != nil {
		return resp.NewSimpleError("value is not an float")
	}

	val, ok := h.IncrByFloat(field, increment)
	if !ok {
		return resp.NewSimpleError("hash value is not an float")
	}

	return resp.NewInteger(int64(val))
}

func hKeysHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	keys := h.Keys()
	res := make([]resp.RedisData, 0, h.Len())
	for _, key := range keys {
		res = append(res, resp.NewBulkString([]byte(key)))
	}

	return resp.NewArray(res)
}

func hLenHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(0)
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	return resp.NewInteger(int64(h.Len()))
}

func hMGetHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	res := make([]resp.RedisData, 0, len(cmd)-2)
	for _, _field := range cmd[2:] {
		field := string(_field)
		v := h.Get(field)
		res = append(res, resp.NewBulkString(v))
	}

	return resp.NewArray(res)
}

func hMSetHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 4 || len(cmd)&1 == 1 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// If key does not exist, a new key holding a hash is created.
	v, ok := db.dict.Get(key)
	if !ok {
		db.dict.Set(key, NewHash())
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	for i := 2; i < len(cmd); i += 2 {
		field := string(cmd[i])
		val := cmd[i+1]
		h.Set(field, val)
	}

	return resp.NewSimpleString("OK")
}

func hSetHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 4 || len(cmd)&1 == 1 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// If key does not exist, a new key holding a hash is created.
	v, ok := db.dict.Get(key)
	if !ok {
		db.dict.Set(key, NewHash())
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	res := 0
	for i := 2; i < len(cmd); i += 2 {
		field := string(cmd[i])
		val := cmd[i+1]
		h.Set(field, val)
		res++
	}

	return resp.NewInteger(int64(res))
}

func hSetNxHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// If key does not exist, a new key holding a hash is created.
	v, ok := db.dict.Get(key)
	if !ok {
		db.dict.Set(key, NewHash())
	} else {
		// 0 if the field already exists in the hash and no operation was performed.
		return resp.NewInteger(0)
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	field := string(cmd[3])
	val := cmd[4]
	h.Set(field, val)

	return resp.NewInteger(1)
}

func hValsHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		// an empty list when key does not exist.
		return resp.NewArray([]resp.RedisData{})
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	vals := h.Values()
	res := make([]resp.RedisData, 0, h.Len())
	for _, val := range vals {
		res = append(res, resp.NewBulkString(val))
	}

	return resp.NewArray(res)
}

func hStrLenHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(0)
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	field := string(cmd[2])
	return resp.NewInteger(int64(h.StrLen(field)))
}

func hRandFieldHash(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 2 || len(cmd) > 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}
	var withvalues bool
	var err error
	count := 1

	if len(cmd) >= 3 {
		count, err = strconv.Atoi(string(cmd[2]))
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
		}
	}

	if len(cmd) == 4 {
		if strings.ToLower(string(cmd[3])) != "withvalues" {
			return resp.NewSimpleError("syntax error")
		} else {
			withvalues = true
		}
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewBulkString(nil)
	}

	// wrong type
	h, ok := v.(*Hash)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	//  a single, randomly selected field when the count option is not used
	if count == 1 {
		fields := h.Random(1)
		return resp.NewBulkString([]byte(fields[0]))
	}

	var res []resp.RedisData
	if withvalues {
		fields, values := h.RandomWithValue(count)
		res = make([]resp.RedisData, 0, len(fields)*2)
		for i := 0; i < len(fields); i += 2 {
			res = append(res, resp.NewBulkString([]byte(fields[i])))
			res = append(res, resp.NewBulkString(values[i]))
		}
	} else {
		fields := h.Random(count)
		res = make([]resp.RedisData, 0, len(fields))
		for i := 0; i < len(fields); i++ {
			res = append(res, resp.NewBulkString([]byte(fields[i])))
		}
	}
	return resp.NewArray(res)
}

func RegisterHashCommands() {
	RegisterCommand("hdel", hDelHash)
	RegisterCommand("hexists", hExistsHash)
	RegisterCommand("hget", hGetHash)
	RegisterCommand("hgetall", hGetAllHash)
	RegisterCommand("hincrby", hIncrByHash)
	RegisterCommand("hincrbyfloat", hIncrByFloatHash)
	RegisterCommand("hkeys", hKeysHash)
	RegisterCommand("hlen", hLenHash)
	RegisterCommand("hmget", hMGetHash)
	RegisterCommand("hmset", hMSetHash)
	RegisterCommand("hset", hSetHash)
	RegisterCommand("hsetnx", hSetNxHash)
	RegisterCommand("hvals", hValsHash)
	RegisterCommand("hstrlen", hStrLenHash)
	RegisterCommand("hrandfield", hRandFieldHash)
}
