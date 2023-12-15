package memdb

import (
	"bytes"
	"gRedis/config"
	"gRedis/resp"
	"testing"
	"time"
)

func init() {
	config.Conf = &config.Config{SegNum: 100}
}

func TestDelKey(t *testing.T) {
	db := NewMemDb()
	db.dict.Set("key1", "Hello")
	db.dict.Set("key2", "World")

	del := delKey(db, [][]byte{[]byte("del"), []byte("key1"), []byte("key2"), []byte("key3")})

	if !bytes.Equal(del.ToRedisFormat(), []byte(":2\r\n")) {
		t.Error("del num is not correct")
	}

	_, ok1 := db.dict.Get("key1")
	_, ok2 := db.dict.Get("key2")
	_, ok3 := db.dict.Get("key3")
	if ok1 || ok2 || ok3 {
		t.Error("del wrong")
	}
}

func TestExistsKey(t *testing.T) {
	db := NewMemDb()
	db.dict.Set("key1", "Hello")

	exists1 := existsKey(db, [][]byte{[]byte("exists"), []byte("key1")})
	if !bytes.Equal(exists1.ToRedisFormat(), []byte(":1\r\n")) {
		t.Error("exists num is not correct")
	}

	exists2 := existsKey(db, [][]byte{[]byte("exists"), []byte("nosuchkey")})
	if !bytes.Equal(exists2.ToRedisFormat(), []byte(":0\r\n")) {
		t.Error("exists num is not correct")
	}

	db.dict.Set("key2", "World")

	exists3 := existsKey(db, [][]byte{[]byte("exists"), []byte("key1"), []byte("key2"), []byte("nosuchkey")})
	if !bytes.Equal(exists3.ToRedisFormat(), []byte(":2\r\n")) {
		t.Error("exists num is not correct")
	}
}

func TestKeysKey(t *testing.T) {
	db := NewMemDb()
	db.dict.Set("firstname", "Jack")
	db.dict.Set("lastname", "Stuntman")
	db.dict.Set("age", 35)

	// case *name*
	_keys1 := keysKey(db, [][]byte{[]byte("keys"), []byte("*name*")})
	keys1 := _keys1.(*resp.RedisArray).GetData()

	var match1, match2 bool
	for _, data := range keys1 {
		if bytes.Equal(data.GetBytesData(), []byte("lastname")) {
			match1 = true
		}
		if bytes.Equal(data.GetBytesData(), []byte("firstname")) {
			match2 = true
		}
	}

	if !match1 || !match2 || len(keys1) != 2 {
		t.Error("keys not match")
	}

	// case a??
	_keys2 := keysKey(db, [][]byte{[]byte("keys"), []byte("a??")})
	keys2 := _keys2.(*resp.RedisArray).GetData()

	var match3 bool
	for _, data := range keys2 {
		if bytes.Equal(data.GetBytesData(), []byte("age")) {
			match3 = true
		}
	}

	if !match3 || len(keys2) != 1 {
		t.Error("keys not match")
	}

	// case *
	_keys3 := keysKey(db, [][]byte{[]byte("keys"), []byte("*")})
	keys3 := _keys3.(*resp.RedisArray).GetData()

	match1, match2, match3 = false, false, false
	for _, data := range keys3 {
		if bytes.Equal(data.GetBytesData(), []byte("lastname")) {
			match1 = true
		}
		if bytes.Equal(data.GetBytesData(), []byte("firstname")) {
			match2 = true
		}
		if bytes.Equal(data.GetBytesData(), []byte("age")) {
			match3 = true
		}
	}

	if !match1 || !match2 || !match3 || len(keys3) != 3 {
		t.Error("keys not match")
	}
}

func TestExpireKey(t *testing.T) {
	db := NewMemDb()
	db.dict.Set("mykey", "Hello")
	db.dict.Set("mykey1", "Hello")

	// no option
	expire1 := expireKey(db, [][]byte{[]byte("expire"), []byte("mykey"), []byte("10")})
	if !bytes.Equal(expire1.ToRedisFormat(), []byte(":1\r\n")) {
		t.Error("expire reply is not correct")
	}
	ttl1, _ := db.expires.Get("mykey")
	if ttl1.(int64)-time.Now().Unix() != 10 {
		t.Error("expire incorrect")
	}

	// "XX" for persistent key
	expire2 := expireKey(db, [][]byte{[]byte("expire"), []byte("mykey1"), []byte("10"), []byte("XX")})
	if !bytes.Equal(expire2.ToRedisFormat(), []byte(":0\r\n")) {
		t.Error("expire reply is not correct")
	}
	_, ok2 := db.expires.Get("mykey1")
	if ok2 {
		t.Error("expire incorrect")
	}

	// "NX" for persistent key
	expire3 := expireKey(db, [][]byte{[]byte("expire"), []byte("mykey1"), []byte("10"), []byte("NX")})
	if !bytes.Equal(expire3.ToRedisFormat(), []byte(":1\r\n")) {
		t.Error("expire reply is not correct")
	}
	ttl3, _ := db.expires.Get("mykey1")
	if ttl3.(int64)-time.Now().Unix() != 10 {
		t.Error("expire incorrect")
	}
}

func TestTtlKey(t *testing.T) {
	db := NewMemDb()
	db.dict.Set("mykey", "Hello")

	expireKey(db, [][]byte{[]byte("expire"), []byte("mykey"), []byte("10")})
	ttl := ttlKey(db, [][]byte{[]byte("ttl"), []byte("mykey")})
	if !bytes.Equal(ttl.ToRedisFormat(), []byte(":10\r\n")) {
		t.Error("ttl incorrect")
	}
}

func TestPersistKey(t *testing.T) {
	db := NewMemDb()
	db.dict.Set("mykey", "Hello")

	expireKey(db, [][]byte{[]byte("expire"), []byte("mykey"), []byte("10")})
	ttl1 := ttlKey(db, [][]byte{[]byte("ttl"), []byte("mykey")})
	if !bytes.Equal(ttl1.ToRedisFormat(), []byte(":10\r\n")) {
		t.Error("persist incorrect")
	}

	persist := persistKey(db, [][]byte{[]byte("persist"), []byte("mykey")})
	if !bytes.Equal(persist.ToRedisFormat(), []byte(":1\r\n")) {
		t.Error("persist incorrect")
	}
	ttl2 := ttlKey(db, [][]byte{[]byte("ttl"), []byte("mykey")})
	if !bytes.Equal(ttl2.ToRedisFormat(), []byte(":-1\r\n")) {
		t.Error("persist incorrect")
	}

}

func TestRenameKey(t *testing.T) {
	db := NewMemDb()
	db.dict.Set("mykey", "Hello")

	rename := renameKey(db, [][]byte{[]byte("rename"), []byte("mykey"), []byte("myotherkey")})
	if !bytes.Equal(rename.ToRedisFormat(), []byte("+OK\r\n")) {
		t.Error("rename incorrect")
	}

	v, _ := db.dict.Get("myotherkey")
	if v.(string) != "Hello" {
		t.Error("rename incorrect")
	}
}
