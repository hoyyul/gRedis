package memdb

import "gRedis/resp"

func sAddSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
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
		v = NewSet()
		db.dict.Set(key, v)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	res := 0
	for _, _key := range cmd[2:] {
		key := string(_key)
		res += s.Add(key)
	}

	return resp.NewInteger(int64(res))
}

func sCardSet(db *MemDb, cmd [][]byte) resp.RedisData {
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

	// If key does not exist, a new key holding a hash is created.
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	return resp.NewInteger(int64(s.Len()))
}

func sDiffSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	keys := make([]string, 0, len(cmd)-1)
	for _, _key := range cmd[1:] {
		key := string(_key)
		if !db.DeleteExpiredKey(key) {
			// can't guarantee the key must to be existed
			keys = append(keys, key)
		}
	}

	if

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// If key does not exist, a new key holding a hash is created.
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	return resp.NewInteger(int64(s.Len()))
}
