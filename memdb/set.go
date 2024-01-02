package memdb

import (
	"gRedis/resp"
	"strconv"
)

func sAddSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

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

	db.locks.MRLock(keys)
	defer db.locks.MRUnLock(keys)

	v, ok := db.dict.Get(keys[0])
	if !ok {
		return resp.NewArray(nil)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	var res []resp.RedisData
	if len(keys) == 1 {
		members := s.Members()
		for _, mem := range members {
			res = append(res, resp.NewBulkString([]byte(mem)))
		}
		return resp.NewArray(res)
	}

	sets := make([]*Set, 0, len(keys)-1)
	for _, key := range keys[1:] {
		_set, ok := db.dict.Get(key)
		if ok {
			set, ok := _set.(*Set)
			if !ok {
				return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
			}
			sets = append(sets, set)
		}
	}

	diff := s.Difference(sets...)

	for _, mem := range diff.Members() {
		res = append(res, resp.NewBulkString([]byte(mem)))
	}

	return resp.NewArray(res)
}

func sInterStoreSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// check destination
	dest := string(cmd[1])
	db.DeleteExpiredKey(dest)

	// check set operation keys
	keys := make([]string, 0, len(cmd)-2)
	for _, _key := range cmd[2:] {
		key := string(_key)
		if !db.DeleteExpiredKey(key) {
			// can't guarantee the key must to be existed
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return resp.NewInteger(0)
	}

	// get primary key
	key := keys[0]

	// lock destination + keys
	lockKeys := append([]string{dest}, keys...)

	db.locks.MLock(lockKeys)
	defer db.locks.MUnLock(lockKeys)

	// check primary key
	v, ok := db.dict.Get(key)
	if !ok {
		// if primary doesn't existed, treat it as a empty set
		v = NewSet()
	}

	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	// check destination key
	destVal, ok := db.dict.Get(dest)
	if !ok {
		destVal = NewSet()
	}

	destSet, ok := destVal.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	// if destination set has at least one member, than save to dict tabe before return
	defer func() {
		if destSet.Len() > 0 {
			db.dict.Set(dest, destSet)
		}
	}()

	// only primary key
	if len(keys) == 1 {
		members := s.Members()
		for _, mem := range members {
			destSet.Add(mem)
		}
		return resp.NewInteger(int64(destSet.Len()))
	}

	// calculate set operation
	sets := make([]*Set, 0, len(keys)-1)
	for _, key := range keys[2:] {
		_set, ok := db.dict.Get(key)
		if ok {
			set, ok := _set.(*Set)
			if !ok {
				return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
			}
			sets = append(sets, set)
		}
	}

	inter := s.Intersect(sets...)

	for _, mem := range inter.Members() {
		destSet.Add(mem)
	}

	return resp.NewInteger(int64(destSet.Len()))
}

func sInterSet(db *MemDb, cmd [][]byte) resp.RedisData {
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

	db.locks.MRLock(keys)
	defer db.locks.MRUnLock(keys)

	v, ok := db.dict.Get(keys[0])
	if !ok {
		return resp.NewArray(nil)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	var res []resp.RedisData
	if len(keys) == 1 {
		members := s.Members()
		for _, mem := range members {
			res = append(res, resp.NewBulkString([]byte(mem)))
		}
		return resp.NewArray(res)
	}

	sets := make([]*Set, 0, len(keys)-1)
	for _, key := range keys[1:] {
		_set, ok := db.dict.Get(key)
		if ok {
			set, ok := _set.(*Set)
			if !ok {
				return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
			}
			sets = append(sets, set)
		}
	}

	inter := s.Intersect(sets...)

	for _, mem := range inter.Members() {
		res = append(res, resp.NewBulkString([]byte(mem)))
	}

	return resp.NewArray(res)
}

func sDiffStoreSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// check destination
	dest := string(cmd[1])
	db.DeleteExpiredKey(dest)

	// check set operation keys
	keys := make([]string, 0, len(cmd)-2)
	for _, _key := range cmd[2:] {
		key := string(_key)
		if !db.DeleteExpiredKey(key) {
			// can't guarantee the key must to be existed
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return resp.NewInteger(0)
	}

	// get primary key
	key := keys[0]

	// lock destination + keys
	lockKeys := append([]string{dest}, keys...)

	db.locks.MLock(lockKeys)
	defer db.locks.MUnLock(lockKeys)

	// check primary key
	v, ok := db.dict.Get(key)
	if !ok {
		// if primary doesn't existed, treat it as a empty set
		v = NewSet()
	}

	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	// check destination key
	destVal, ok := db.dict.Get(dest)
	if !ok {
		destVal = NewSet()
	}

	destSet, ok := destVal.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	// if destination set has at least one member, than save to dict tabe before return
	defer func() {
		if destSet.Len() > 0 {
			db.dict.Set(dest, destSet)
		}
	}()

	// only primary key
	if len(keys) == 1 {
		members := s.Members()
		for _, mem := range members {
			destSet.Add(mem)
		}
		return resp.NewInteger(int64(destSet.Len()))
	}

	// calculate set operation
	sets := make([]*Set, 0, len(keys)-1)
	for _, key := range keys[2:] {
		_set, ok := db.dict.Get(key)
		if ok {
			set, ok := _set.(*Set)
			if !ok {
				return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
			}
			sets = append(sets, set)
		}
	}

	diff := s.Difference(sets...)

	for _, mem := range diff.Members() {
		destSet.Add(mem)
	}

	return resp.NewInteger(int64(destSet.Len()))
}

func sIsMemberSet(db *MemDb, cmd [][]byte) resp.RedisData {
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

	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	if !s.Has(key) {
		return resp.NewInteger(0)
	}

	return resp.NewInteger(1)
}

func sMembersSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewArray(nil)
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewArray(nil)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	var res []resp.RedisData
	members := s.Members()
	for _, mem := range members {
		res = append(res, resp.NewBulkString([]byte(mem)))
	}

	return resp.NewArray(res)
}

func sMoveSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	src := string(cmd[1])
	if db.DeleteExpiredKey(src) {
		return resp.NewInteger(0)
	}

	des := string(cmd[2])
	if db.DeleteExpiredKey(des) {
		return resp.NewInteger(0)
	}

	member := string(cmd[3])

	db.locks.MLock([]string{src, des})
	defer db.locks.MUnLock([]string{src, des})

	srcVal, ok := db.dict.Get(src)
	if !ok {
		return resp.NewInteger(0)
	}

	srcSet, ok := srcVal.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	defer func() {
		if srcSet.Len() == 0 {
			db.dict.Delete(src)
			db.DeleteExpire(src)
		}
	}()

	desVal, ok := db.dict.Get(des)
	if !ok {
		return resp.NewInteger(0)
	}

	desSet, ok := desVal.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	res := srcSet.Move(desSet, member)

	return resp.NewInteger(int64(res))
}

func sPopSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 && len(cmd) != 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	countVal := 1
	var err error
	var count bool
	if len(cmd) == 3 {
		count = true
		countVal, err = strconv.Atoi(string(cmd[2]))
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
		}
	}

	if countVal <= 0 {
		return resp.NewSimpleError("value must be positive")
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewBulkString(nil)
	}

	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	defer func() {
		if s.Len() == 0 {
			db.dict.Delete(key)
			db.DeleteExpire(key)
		}
	}()

	if count {
		pop := s.Pop()
		if pop == "" {
			return resp.NewBulkString(nil)
		}
		return resp.NewBulkString([]byte(pop))
	}

	if countVal > s.Len() {
		countVal = s.Len()
	}

	res := make([]resp.RedisData, 0, countVal)
	for i := 0; i < countVal; i++ {
		res = append(res, resp.NewBulkString([]byte(s.Pop())))
	}

	return resp.NewArray(res)
}

func sRandMemberSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 && len(cmd) != 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	countVal := 1
	var err error
	var count bool
	if len(cmd) == 3 {
		count = true
		countVal, err = strconv.Atoi(string(cmd[2]))
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
		}
	}

	if countVal == 0 {
		return resp.NewBulkString(nil)
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewBulkString(nil)
	}

	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	var rand []string
	if count {
		rand = s.Random(1)
		if rand[0] == "" {
			return resp.NewBulkString(nil)
		}
		return resp.NewBulkString([]byte(rand[0]))
	}

	if countVal > s.Len() {
		countVal = s.Len()
	}

	res := make([]resp.RedisData, 0, countVal)
	rand = s.Random(countVal)
	for i := 0; i < countVal; i++ {
		res = append(res, resp.NewBulkString([]byte(rand[i])))
	}

	return resp.NewArray(res)
}

func sRemSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

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
		res += s.Remove(key)
	}

	return resp.NewInteger(int64(res))
}

func sUnionSet(db *MemDb, cmd [][]byte) resp.RedisData {
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

	db.locks.MRLock(keys)
	defer db.locks.MRUnLock(keys)

	v, ok := db.dict.Get(keys[0])
	if !ok {
		return resp.NewArray(nil)
	}

	// wrong type
	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	var res []resp.RedisData
	if len(keys) == 1 {
		members := s.Members()
		for _, mem := range members {
			res = append(res, resp.NewBulkString([]byte(mem)))
		}
		return resp.NewArray(res)
	}

	sets := make([]*Set, 0, len(keys)-1)
	for _, key := range keys[1:] {
		_set, ok := db.dict.Get(key)
		if ok {
			set, ok := _set.(*Set)
			if !ok {
				return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
			}
			sets = append(sets, set)
		}
	}

	union := s.Union(sets...)

	for _, mem := range union.Members() {
		res = append(res, resp.NewBulkString([]byte(mem)))
	}

	return resp.NewArray(res)
}

func sUnionStoreSet(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// check destination
	dest := string(cmd[1])
	db.DeleteExpiredKey(dest)

	// check set operation keys
	keys := make([]string, 0, len(cmd)-2)
	for _, _key := range cmd[2:] {
		key := string(_key)
		if !db.DeleteExpiredKey(key) {
			// can't guarantee the key must to be existed
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return resp.NewInteger(0)
	}

	// get primary key
	key := keys[0]

	// lock destination + keys
	lockKeys := append([]string{dest}, keys...)

	db.locks.MLock(lockKeys)
	defer db.locks.MUnLock(lockKeys)

	// check primary key
	v, ok := db.dict.Get(key)
	if !ok {
		// if primary doesn't existed, treat it as a empty set
		v = NewSet()
	}

	s, ok := v.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	// check destination key
	destVal, ok := db.dict.Get(dest)
	if !ok {
		destVal = NewSet()
	}

	destSet, ok := destVal.(*Set)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	// if destination set has at least one member, than save to dict tabe before return
	defer func() {
		if destSet.Len() > 0 {
			db.dict.Set(dest, destSet)
		}
	}()

	// only primary key
	if len(keys) == 1 {
		members := s.Members()
		for _, mem := range members {
			destSet.Add(mem)
		}
		return resp.NewInteger(int64(destSet.Len()))
	}

	// calculate set operation
	sets := make([]*Set, 0, len(keys)-1)
	for _, key := range keys[2:] {
		_set, ok := db.dict.Get(key)
		if ok {
			set, ok := _set.(*Set)
			if !ok {
				return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
			}
			sets = append(sets, set)
		}
	}

	union := s.Union(sets...)

	for _, mem := range union.Members() {
		destSet.Add(mem)
	}

	return resp.NewInteger(int64(destSet.Len()))
}

func RegisterSetCommands() {
	RegisterCommand("sadd", sAddSet)
	RegisterCommand("scard", sCardSet)
	RegisterCommand("sdiff", sDiffSet)
	RegisterCommand("sdiffstore", sDiffStoreSet)
	RegisterCommand("sinter", sInterSet)
	RegisterCommand("sinterstore", sInterStoreSet)
	RegisterCommand("sismember", sIsMemberSet)
	RegisterCommand("smembers", sMembersSet)
	RegisterCommand("smove", sMoveSet)
	RegisterCommand("spop", sPopSet)
	RegisterCommand("srandmember", sRandMemberSet)
	RegisterCommand("srem", sRemSet)
	RegisterCommand("sunion", sUnionSet)
	RegisterCommand("sunionstore", sUnionStoreSet)
}
