package memdb

import (
	"fmt"
	"gRedis/resp"
	"strconv"
	"strings"
	"time"
)

/*
EX seconds -- Set the specified expire time, in seconds.
PX milliseconds -- Set the specified expire time, in milliseconds.
EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds.
PXAT timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds.
NX -- Only set the key if it does not already exist.
XX -- Only set the key if it already exists.
KEEPTTL -- Retain the time to live associated with the key.
GET -- Return the old string stored at key, or nil if key did not exist. An error is returned and SET aborted if the value stored at key is not a string.
*/
func setString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	val := cmd[2]

	// parse options
	var nx, xx, get, ex, px, exat, pxat, keepttl bool
	var exval, pxval, exatval, pxatval int64
	var err error
	optNum := 0
	for i := 3; i < len(cmd); i++ {
		switch strings.ToLower(string(cmd[i])) {
		case "nx":
			nx = true
		case "xx":
			xx = true
		case "get":
			get = true
		case "ex":
			ex = true
			optNum++
			i++
			if i >= len(cmd) {
				resp.NewSimpleError("syntax error")
			}
			exval, err = strconv.ParseInt(string(cmd[i]), 10, 64)
			if err != nil {
				resp.NewSimpleError("value is not an integer")
			}
		case "px":
			px = true
			optNum++
			i++
			if i >= len(cmd) {
				resp.NewSimpleError("syntax error")
			}
			pxval, err = strconv.ParseInt(string(cmd[i]), 10, 64)
			if err != nil {
				return resp.NewSimpleError("value is not an integer")
			}
		case "exat":
			exat = true
			optNum++
			if i >= len(cmd) {
				return resp.NewSimpleError("syntax error")
			}
			exatval, err = strconv.ParseInt(string(cmd[i]), 10, 64)
			if err != nil {
				return resp.NewSimpleError("value is not an integer")
			}
		case "pxat":
			pxat = true
			optNum++
			if i >= len(cmd) {
				return resp.NewSimpleError("syntax error")
			}
			pxatval, err = strconv.ParseInt(string(cmd[i]), 10, 64)
			if err != nil {
				return resp.NewSimpleError("value is not an integer")
			}
		case "keepttl":
			keepttl = true
			optNum++
		default:
			return resp.NewSimpleError(fmt.Sprintf("Unsupported option %s", string(cmd[i])))
		}
	}

	if nx && xx || optNum > 1 {
		return resp.NewSimpleError("syntax error")
	}

	// set
	var res resp.RedisData
	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	oldVal, oldOk := db.dict.Get(key)
	if oldOk {
		if _, ok := oldVal.([]byte); !ok {
			return resp.NewSimpleError("The value stored at key is not a string")
		}
	}

	if nx || xx {
		if nx {
			if !oldOk {
				db.dict.Set(key, val)
				res = resp.NewSimpleString("OK")
			} else {
				res = resp.NewBulkString(nil)
			}
		} else {
			if oldOk {
				db.dict.Set(key, val)
				res = resp.NewSimpleString("OK")
			} else {
				res = resp.NewBulkString(nil)
			}
		}
	} else {
		db.dict.Set(key, val)
		res = resp.NewSimpleString("OK")
	}

	if get {
		if oldOk {
			res = resp.NewBulkString(oldVal.([]byte))
		} else {
			res = resp.NewBulkString(nil)
		}
	}

	if !keepttl {
		db.DeleteExpire(key)
	}

	if ex {
		db.SetExpire(key, time.Now().Unix()+exval)
	}
	if px {
		db.SetExpire(key, time.Now().Unix()+pxval/1000)
	}
	if exat {
		db.SetExpire(key, exatval)
	}
	if pxat {
		db.SetExpire(key, pxatval/1000)
	}

	return res
}

func getString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString(nil)
	}

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	val, ok := db.dict.Get(key)
	if ok {
		if v, ok := val.([]byte); !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		} else {
			return resp.NewBulkString(v)
		}
	} else {
		return resp.NewBulkString(nil)
	}
}

func getRangeString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	if db.DeleteExpiredKey(key) {
		return resp.NewBulkString([]byte{})
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

	val, ok := db.dict.Get(key)
	if !ok {
		return resp.NewBulkString([]byte{})
	}
	v, ok := val.([]byte)
	if !ok {
		return resp.NewSimpleError("ERR the value stored at key is not a string")
	}

	if start < 0 {
		start = len(v) + start
	}
	if end < 0 {
		end = len(v) + end
	}
	end++

	if start > end || start >= len(v) || end < 0 {
		return resp.NewBulkString([]byte{})
	}

	if start < 0 {
		start = 0
	}

	if end > len(v) {
		end = len(v)
	}

	return resp.NewBulkString(v[start:end])
}

func setRangeString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// parse cmd
	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	offset, err := strconv.Atoi(string(cmd[2]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}
	if offset < 0 {
		return resp.NewSimpleError("offset is out of range")
	}

	value := cmd[3]

	var oldVal, newVal []byte
	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	// check if key existed
	val, ok := db.dict.Get(key)
	if ok {
		oldVal, ok = val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}
	} else {
		oldVal = make([]byte, 0)
	}

	//  If the offset is larger than the current length of the string at key, the string is padded with zero-bytes to make offset fit.
	if offset > len(oldVal) {
		newVal = oldVal
		zeroes := make([]byte, offset-len(oldVal)) // pre-allocating memory
		newVal = append(newVal, zeroes...)
	} else {
		newVal = oldVal[:offset]
	}

	newVal = append(newVal, value...)
	db.dict.Set(key, newVal)
	return resp.NewInteger(int64(len(newVal)))
}

func mGetString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	res := make([]resp.RedisData, 0)

	for _, k := range cmd[1:] {
		key := string(k)
		if !db.DeleteExpiredKey(key) {
			db.locks.RLock(key)

			if val, ok := db.dict.Get(key); ok {
				if v, ok := val.([]byte); ok {
					res = append(res, resp.NewBulkString(v))
				} else {
					res = append(res, resp.NewBulkString(nil)) // not string
				}
			} else {
				res = append(res, resp.NewBulkString(nil))
			}

			db.locks.RUnLock(key)
		}
	}

	return resp.NewArray(res)
}

// MSET is atomic, so all given keys are set at once.
func mSetString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) < 3 || len(cmd)&1 != 1 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	keys := make([]string, 0)
	vals := make([][]byte, 0)
	for i := 1; i < len(cmd); i += 2 {
		keys = append(keys, string(cmd[i]))
		vals = append(vals, cmd[i+1])
	}

	db.locks.MLock(keys)
	defer db.locks.MUnLock(keys)

	for i := 0; i < len(keys); i++ {
		db.DeleteExpiredKey(keys[i])
		db.dict.Set(keys[i], vals[i])
	}

	return resp.NewSimpleString("OK")
}

func setExString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 4 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	sec, err := strconv.ParseInt(string(cmd[2]), 10, 64)
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	val := cmd[3]

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	db.dict.Set(key, val)
	db.SetExpire(key, time.Now().Unix()+sec)

	return resp.NewSimpleString("OK")
}

func setNxString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	if _, ok := db.dict.Get(key); ok {
		return resp.NewInteger(0)
	}

	val := cmd[2]

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	db.dict.Set(key, val)

	return resp.NewInteger(1)
}

func strLenString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.RLock(key)
	defer db.locks.RUnLock(key)

	if val, ok := db.dict.Get(key); ok {
		v, ok := val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}
		return resp.NewInteger(int64(len(v)))
	} else {
		return resp.NewInteger(0)
	}
}

func incrString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	if val, ok := db.dict.Get(key); ok {
		v, ok := val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}

		nV, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
		}

		nV++
		nVal := strconv.FormatInt(nV, 10)
		db.dict.Set(key, []byte(nVal))

		return resp.NewInteger(nV)
	} else {
		return resp.NewInteger(0)
	}
}

func incrByString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	increment, err := strconv.ParseInt(string(cmd[2]), 10, 64)
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	if val, ok := db.dict.Get(key); ok {
		v, ok := val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}

		nV, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
		}

		nV += increment
		nVal := strconv.FormatInt(nV, 10)
		db.dict.Set(key, []byte(nVal))

		return resp.NewInteger(nV)
	} else {
		return resp.NewInteger(0)
	}
}

func decrString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	if val, ok := db.dict.Get(key); ok {
		v, ok := val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}

		nV, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return resp.NewSimpleError("value is not an integer")
		}

		nV--
		nVal := strconv.FormatInt(nV, 10)
		db.dict.Set(key, []byte(nVal))

		return resp.NewInteger(nV)
	} else {
		return resp.NewInteger(0)
	}
}

func decrByString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	decrement, err := strconv.ParseInt(string(cmd[2]), 10, 64)
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	if val, ok := db.dict.Get(key); ok {
		v, ok := val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}

		nV, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return resp.NewSimpleError("(error) value is not an integer")
		}

		nV -= decrement
		nVal := strconv.FormatInt(nV, 10)
		db.dict.Set(key, []byte(nVal))

		return resp.NewInteger(nV)
	} else {
		return resp.NewInteger(0)
	}
}

func incrByFloatString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	increment, err := strconv.ParseFloat(string(cmd[2]), 64)
	if err != nil {
		return resp.NewSimpleError("value is not an float")
	}

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	if val, ok := db.dict.Get(key); ok {
		v, ok := val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}

		fV, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return resp.NewSimpleError("value is not an float")
		}

		fV += increment
		fVal := strconv.FormatFloat(fV, 'f', -1, 64)
		db.dict.Set(key, []byte(fVal))

		return resp.NewBulkString([]byte(fVal))
	} else {
		//  If the key does not exist, it is set to 0 before performing the operation.
		f := []byte(strconv.FormatFloat(increment, 'f', -1, 64))
		db.dict.Set(key, f)
		return resp.NewBulkString(f)
	}
}

func appendString(db *MemDb, cmd [][]byte) resp.RedisData {
	if len(cmd) != 3 {
		return resp.NewSimpleError("ERR wrong number of arguments for command")
	}

	// passive delete expired key
	key := string(cmd[1])
	db.DeleteExpiredKey(key)

	apd := cmd[2]

	db.locks.Lock(key)
	defer db.locks.UnLock(key)

	if val, ok := db.dict.Get(key); ok {
		v, ok := val.([]byte)
		if !ok {
			return resp.NewSimpleError("ERR the value stored at key is not a string")
		}
		v = append(v, apd...)
		db.dict.Set(key, v)
		return resp.NewInteger(int64(len(v)))
	} else {
		db.dict.Set(key, apd)
		return resp.NewInteger(int64(len(apd)))
	}

}

func RegisterStringCommands() {
	RegisterCommand("set", setString)
	RegisterCommand("get", getString)
	RegisterCommand("getrange", getRangeString)
	RegisterCommand("setrange", setRangeString)
	RegisterCommand("mget", mGetString)
	RegisterCommand("mset", mSetString)
	RegisterCommand("setex", setExString)
	RegisterCommand("setnx", setNxString)
	RegisterCommand("strlen", strLenString)
	RegisterCommand("incr", incrString)
	RegisterCommand("incrby", incrByString)
	RegisterCommand("decr", decrString)
	RegisterCommand("decrby", decrByString)
	RegisterCommand("incrbyfloat", incrByFloatString)
	RegisterCommand("append", appendString)
}
