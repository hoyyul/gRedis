package memdb

import (
	"bytes"
	"fmt"
	"gRedis/resp"
	"strconv"
	"strings"
)

func lIndexList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	index, err := strconv.Atoi(string(cmd[2]))
	if err != nil {
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
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
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	node := l.Index(index)
	if node == nil {
		return resp.NewBulkString(nil)
	}

	return resp.NewBulkString(node.Val)
}

func lInsertList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 5 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(0)
	}

	var before bool
	flag := strings.ToLower(string(cmd[2]))
	if flag != "before" && flag != "after" {
		return resp.NewSimpleError("syntax error")
	} else {
		if flag == "before" {
			before = true
		}
	}

	pivot := cmd[3]
	val := cmd[4]

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	if before {
		if !l.InsertBefore(val, pivot) {
			// -1 when the pivot wasn't found.
			return resp.NewInteger(-1)
		}
	} else {
		if !l.InsertAfter(val, pivot) {
			// -1 when the pivot wasn't found.
			return resp.NewInteger(-1)
		}
	}

	return resp.NewInteger(int64(l.Len))
}

func lLenList(db *MemDb, cmd [][]byte) resp.RedisData {
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
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	return resp.NewInteger(int64(l.Len))
}

func lMoveList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 5 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	src := string(cmd[1])
	if db.DeleteExpiredKey(src) {
		return resp.NewBulkString(nil)
	}

	des := string(cmd[2])
	db.DeleteExpiredKey(des)
	srcDrc := strings.ToLower(string(cmd[3]))
	desDrc := strings.ToLower(string(cmd[4]))
	if (srcDrc != "left" && srcDrc != "right") || (desDrc != "left" && desDrc != "right") {
		return resp.NewSimpleError("syntax error")
	}

	db.locks.MLock([]string{src, des})
	defer db.locks.MUnLock([]string{src, des})

	// key not existed
	srcVal, ok := db.dict.Get(src)
	if !ok {
		return resp.NewBulkString(nil)
	}
	desVal, ok := db.dict.Get(des)
	if !ok {
		desVal = NewList()
		db.dict.Set(des, desVal)
	}

	// wrong type
	srcList, ok := srcVal.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	defer func() {
		if srcList.Len == 0 {
			db.dict.Delete(src)
			db.DeleteExpire(src)
		}
	}()

	desList, ok := desVal.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	var srcPop *ListNode
	if srcDrc == "left" {
		srcPop = srcList.LPop()
	} else {
		srcPop = srcList.RPop()
	}

	if desDrc == "left" {
		desList.LPush(srcPop.Val)
	} else {
		fmt.Println(desList, desList.Head, desList.Tail)
		desList.RPush(srcPop.Val)
	}

	return resp.NewBulkString(srcPop.Val)
}

func lPopList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	count := 1
	var err error
	if len(cmd) == 3 {
		count, err = strconv.Atoi(string(cmd[2]))
		if err != nil || count <= 0 {
			return resp.NewSimpleError("value is out of range, must be positive")
		}
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewBulkString(nil)
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	defer func() {
		if l.Len == 0 {
			db.dict.Delete(key)
			db.DeleteExpire(key)
		}
	}()

	if count == 1 {
		node := l.LPop()
		if node == nil {
			return resp.NewBulkString(nil)
		} else {
			return resp.NewBulkString(node.Val)
		}
	}

	if count > l.Len {
		count = l.Len
	}

	res := make([]resp.RedisData, 0, count)
	for i := 0; i < count; i++ {
		node := l.LPop()
		if node == nil {
			break
		}
		res = append(res, resp.NewBulkString(node.Val))
	}

	return resp.NewArray(res)
}

func lPosList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 || len(cmd)&1 != 1 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	element := cmd[2]

	var rank, count, maxLen bool
	var rankVal, countVal, maxLenVal int
	if len(cmd) > 3 {
		// parse options
		var err error
		for i := 3; i < len(cmd); i += 2 {
			switch strings.ToLower(string(cmd[i])) {
			case "rank":
				rank = true
				rankVal, err = strconv.Atoi(string(cmd[i+1]))
				if err != nil {
					return resp.NewSimpleError("value is not an integer")
				}
				if rankVal == 0 {
					return resp.NewSimpleError("RANK can't be zero: use 1 to start from the first match, 2 from the second ... or use negative to start from the end of the list")
				}
			case "count":
				count = true
				countVal, err = strconv.Atoi(string(cmd[i+1]))
				if err != nil {
					return resp.NewSimpleError("value is not an integer")
				}
				if countVal < 0 {
					return resp.NewSimpleError("COUNT can't be negative")
				}
			case "maxlen":
				maxLen = true
				maxLenVal, err = strconv.Atoi(string(cmd[i+1]))
				if err != nil {
					return resp.NewSimpleError("value is not an integer")
				}
				if maxLenVal < 0 {
					return resp.NewSimpleError("MAXLEN can't be negative")
				}
			default:
				return resp.NewSimpleError(fmt.Sprintf("Unsupported option %s", string(cmd[i])))
			}
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
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	var res []resp.RedisData
	var reverse bool
	var pos int

	// normally pos without options
	if !count && !rank && !maxLen {
		pos = l.Pos(element)
		if pos == -1 {
			return resp.NewBulkString(nil)
		} else {
			return resp.NewInteger(int64(pos))
		}
	}

	if count && countVal == 0 {
		countVal = l.Len
	}

	if maxLen && maxLenVal == 0 {
		maxLenVal = l.Len
	}

	var cur *ListNode
	if rank {
		if rankVal > 0 {
			pos = -1
			for cur = l.Head.Next; cur != l.Tail; cur = cur.Next {
				pos++
				if bytes.Equal(element, cur.Val) {
					rankVal--
				}
				if maxLen {
					maxLenVal--
					if maxLenVal == 0 {
						break
					}
				}
				if rankVal == 0 {
					break
				}
			}
		} else {
			reverse = true
			pos = l.Len
			for cur = l.Tail.Prev; cur != l.Head; cur = cur.Prev {
				pos--
				if bytes.Equal(element, cur.Val) {
					rankVal++
				}
				if maxLen {
					maxLenVal--
					if maxLenVal == 0 {
						break
					}
				}
				if rankVal == 0 {
					break
				}
			}
		}
	} else {
		cur = l.Head.Next
		pos = 0
		if maxLen {
			maxLenVal--
		}
	}

	// when rank is out of range, return nil
	if (rank && rankVal != 0) || cur == l.Head || cur == l.Tail {
		return resp.NewBulkString(nil)
	}

	if count {
		if !reverse {
			for ; cur != l.Tail; cur = cur.Next {
				if bytes.Equal(element, cur.Val) {
					res = append(res, resp.NewInteger(int64(pos)))
					countVal--
					if countVal == 0 {
						break
					}
				}
				if maxLen {
					if maxLenVal <= 0 {
						break
					}
					maxLenVal--
				}
				pos++
			}
		} else {
			for ; cur != l.Head; cur = cur.Prev {
				if bytes.Equal(element, cur.Val) {
					res = append(res, resp.NewInteger(int64(pos)))
					countVal--
					if countVal == 0 {
						break
					}
				}
				if maxLen {
					if maxLenVal <= 0 {
						break
					}
					maxLenVal--
				}
				pos--
			}
		}
	} else {
		// if count is not set, return first find pos inside maxLen range
		for ; cur != l.Tail; cur = cur.Next {
			if bytes.Equal(element, cur.Val) {
				return resp.NewInteger(int64(pos))
			}
			pos++
			if maxLen {
				if maxLenVal <= 0 {
					break
				}
				maxLenVal--
			}
		}
		return resp.NewBulkString(nil)
	}

	if len(res) == 0 {
		return resp.NewBulkString(nil)
	}
	return resp.NewArray(res)
}

func lPushList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		v = NewList()
		db.dict.Set(key, v)
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	for _, element := range cmd[2:] {
		l.LPush(element)
	}

	return resp.NewInteger(int64(l.Len))
}

func lPushXList(db *MemDb, cmd [][]byte) resp.RedisData {
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
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	for _, element := range cmd[2:] {
		l.LPush(element)
	}

	return resp.NewInteger(int64(l.Len))
}

func lRangeList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewArray(nil)
	}

	start, err := strconv.Atoi(string(cmd[2]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	end, err := strconv.Atoi(string(cmd[3]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewArray(nil)
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	vals := l.Range(start, end)
	res := make([]resp.RedisData, 0, len(vals))

	for i := 0; i < len(vals); i++ {
		res = append(res, resp.NewBulkString(vals[i]))
	}

	return resp.NewArray(res)
}

func lRemList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewInteger(0)
	}

	count, err := strconv.Atoi(string(cmd[2]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	element := cmd[3]

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewInteger(0)
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	defer func() {
		if l.Len == 0 {
			db.dict.Delete(key)
			db.DeleteExpire(key)
		}
	}()

	res := l.Remove(element, count)

	return resp.NewInteger(int64(res))
}

func lSetList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewSimpleError("no such key")
	}

	index, err := strconv.Atoi(string(cmd[2]))
	if err != nil {
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
		}
	}

	val := cmd[3]

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewSimpleError("no such key")
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	if !l.Set(val, index) {
		return resp.NewSimpleError("index out of range")
	}

	return resp.NewSimpleString("OK")
}

func lTrimList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewSimpleError("no such key")
	}

	start, err := strconv.Atoi(string(cmd[2]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	end, err := strconv.Atoi(string(cmd[3]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewSimpleError("no such key")
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	defer func() {
		if l.Len == 0 {
			db.dict.Delete(key)
			db.DeleteExpire(key)
		}
	}()

	l.Trim(start, end)

	return resp.NewSimpleString("OK")
}

func rPopList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	count := 1
	var err error
	if len(cmd) == 3 {
		count, err = strconv.Atoi(string(cmd[2]))
		if err != nil || count <= 0 {
			return resp.NewSimpleError("value is out of range, must be positive")
		}
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		return resp.NewBulkString(nil)
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	defer func() {
		if l.Len == 0 {
			db.dict.Delete(key)
			db.DeleteExpire(key)
		}
	}()

	if count == 1 {
		node := l.RPop()
		if node == nil {
			return resp.NewBulkString(nil)
		} else {
			return resp.NewBulkString(node.Val)
		}
	}

	if count > l.Len {
		count = l.Len
	}

	res := make([]resp.RedisData, 0, count)
	for i := 0; i < count; i++ {
		node := l.RPop()
		if node == nil {
			break
		}
		res = append(res, resp.NewBulkString(node.Val))
	}

	return resp.NewArray(res)
}

func rPushList(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// key not existed
	v, ok := db.dict.Get(key)
	if !ok {
		v = NewList()
		db.dict.Set(key, v)
	}

	// wrong type
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	for _, element := range cmd[2:] {
		l.RPush(element)
	}

	return resp.NewInteger(int64(l.Len))
}

func rPushXList(db *MemDb, cmd [][]byte) resp.RedisData {
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
	l, ok := v.(*List)
	if !ok {
		return resp.NewSimpleError("Operation against a key holding the wrong kind of value")
	}

	for _, element := range cmd[2:] {
		l.RPush(element)
	}

	return resp.NewInteger(int64(l.Len))
}

func RegisterListCommands() {
	RegisterCommand("lindex", lIndexList)
	RegisterCommand("linsert", lInsertList)
	RegisterCommand("llen", lLenList)
	RegisterCommand("lmove", lMoveList)
	RegisterCommand("lpop", lPopList)
	RegisterCommand("lpos", lPosList)
	RegisterCommand("lpush", lPushList)
	RegisterCommand("lpushx", lPushXList)
	RegisterCommand("lrange", lRangeList)
	RegisterCommand("lrem", lRemList)
	RegisterCommand("lset", lSetList)
	RegisterCommand("ltrim", lTrimList)
	RegisterCommand("rpop", rPopList)
	RegisterCommand("rpush", rPushList)
	RegisterCommand("rpushx", rPushXList)
}
