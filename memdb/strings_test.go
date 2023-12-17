package memdb

import (
	"bytes"
	"testing"
)

func TestStringsCommand1(t *testing.T) {
	db := NewMemDb()

	// test SET and GET
	set1 := setString(db, [][]byte{[]byte("set"), []byte("mykey"), []byte("Hello")})
	if !bytes.Equal(set1.ToRedisFormat(), []byte("+OK\r\n")) {
		t.Error("SET is not correct")
	}
	get1 := getString(db, [][]byte{[]byte("get"), []byte("mykey")})
	if !bytes.Equal(get1.GetBytesData(), []byte("Hello")) {
		t.Error("GET is not correct")
	}

	// test SET EX and TTL
	set2 := setString(db, [][]byte{[]byte("set"), []byte("anotherkey"), []byte("will expire in a minute"), []byte("EX"), []byte("60")})
	if !bytes.Equal(set2.ToRedisFormat(), []byte("+OK\r\n")) {
		t.Error("SET is not correct")
	}
	ttl2 := ttlKey(db, [][]byte{[]byte("ttl"), []byte("anotherkey")})
	if !bytes.Equal(ttl2.ToRedisFormat(), []byte(":60\r\n")) {
		t.Error("TTL is not correct")
	}

	// test GET
	get3 := getString(db, [][]byte{[]byte("get"), []byte("nonexisting")})
	if !bytes.Equal(get3.GetBytesData(), nil) {
		t.Error("GET is not correct")
	}
}

func TestStringsCommand2(t *testing.T) {
	db := NewMemDb()

	// SET mykey This is a string
	set := setString(db, [][]byte{[]byte("set"), []byte("mykey"), []byte("This is a string")})
	if !bytes.Equal(set.ToRedisFormat(), []byte("+OK\r\n")) {
		t.Error("SET is not correct")
	}

	// GETRANGE mykey 0 3
	getRange1 := getRangeString(db, [][]byte{[]byte("getrange"), []byte("mykey"), []byte("0"), []byte("3")})
	if !bytes.Equal(getRange1.GetBytesData(), []byte("This")) {
		t.Error("GETRANGE is not correct")
	}

	// GETRANGE mykey -3 -1
	getRange2 := getRangeString(db, [][]byte{[]byte("getrange"), []byte("mykey"), []byte("-3"), []byte("-1")})
	if !bytes.Equal(getRange2.GetBytesData(), []byte("ing")) {
		t.Error("GETRANGE is not correct")
	}

	// GETRANGE mykey 0 -1
	getRange3 := getRangeString(db, [][]byte{[]byte("getrange"), []byte("mykey"), []byte("0"), []byte("-1")})
	if !bytes.Equal(getRange3.GetBytesData(), []byte("This is a string")) {
		t.Error("GETRANGE is not correct")
	}

	// GETRANGE mykey 10 100
	getRange4 := getRangeString(db, [][]byte{[]byte("getrange"), []byte("mykey"), []byte("10"), []byte("100")})
	if !bytes.Equal(getRange4.GetBytesData(), []byte("string")) {
		t.Error("GETRANGE is not correct")
	}
}

func TestStringsCommand3(t *testing.T) {
	db := NewMemDb()

	// SET key1 "Hello World"
	setString(db, [][]byte{[]byte("set"), []byte("key1"), []byte("Hello World")})

	// SETRANGE key1 6 "Redis"
	setRange1 := setRangeString(db, [][]byte{[]byte("setrange"), []byte("key1"), []byte("6"), []byte("Redis")})
	if !bytes.Equal(setRange1.ToRedisFormat(), []byte(":11\r\n")) {
		t.Error("SETRANGE is not correct")
	}

	// GET key1
	get1 := getString(db, [][]byte{[]byte("get"), []byte("key1")})
	if !bytes.Equal(get1.GetBytesData(), []byte("Hello Redis")) {
		t.Error("GET is not correct")
	}

	// SETRANGE key2 6 "Redis"
	setRange2 := setRangeString(db, [][]byte{[]byte("setrange"), []byte("key2"), []byte("6"), []byte("Redis")})
	if !bytes.Equal(setRange2.ToRedisFormat(), []byte(":11\r\n")) {
		t.Error("SETRANGE is not correct")
	}

	// GET key2; zero padding
	get2 := getString(db, [][]byte{[]byte("get"), []byte("key2")})
	if !bytes.Equal(get2.GetBytesData(), []byte{0, 0, 0, 0, 0, 0, 'R', 'e', 'd', 'i', 's'}) {
		t.Error("GET is not correct")
	}
}

func TestStringsCommand4(t *testing.T) {
	db := NewMemDb()

	// SET key1 "Hello"
	setString(db, [][]byte{[]byte("set"), []byte("key1"), []byte("Hello")})

	// SET key2 "World"
	setString(db, [][]byte{[]byte("set"), []byte("key2"), []byte("World")})

	// MGET key1 key2 nonexisting
	mget := mGetString(db, [][]byte{[]byte("mget"), []byte("key1"), []byte("key2"), []byte("nonexisting")})
	if mget.String() != "Hello World nil" {
		t.Error("MGET is not correct")
	}
}

func TestStringsCommand5(t *testing.T) {
	db := NewMemDb()

	// MSET key1 "Hello" key2 "World"
	mSetString(db, [][]byte{[]byte("mset"), []byte("key1"), []byte("Hello"), []byte("key2"), []byte("World")})

	// GET key1
	get1 := getString(db, [][]byte{[]byte("get"), []byte("key1")})
	if !bytes.Equal(get1.GetBytesData(), []byte("Hello")) {
		t.Error("GET is not correct")
	}

	// GET key2
	get2 := getString(db, [][]byte{[]byte("get"), []byte("key2")})
	if !bytes.Equal(get2.GetBytesData(), []byte("World")) {
		t.Error("GET is not correct")
	}
}

func TestStringsCommand6(t *testing.T) {
	db := NewMemDb()

	// SETEX mykey 10 "Hello"
	setExString(db, [][]byte{[]byte("setex"), []byte("mykey"), []byte("10"), []byte("Hello")})

	// TTL mykey
	ttl := ttlKey(db, [][]byte{[]byte("ttl"), []byte("mykey")})
	if !bytes.Equal(ttl.ToRedisFormat(), []byte(":10\r\n")) {
		t.Error("TTL is not correct")
	}

	// GET mykey
	get := getString(db, [][]byte{[]byte("get"), []byte("mykey")})
	if !bytes.Equal(get.GetBytesData(), []byte("Hello")) {
		t.Error("GET is not correct")
	}
}

func TestStringsCommand7(t *testing.T) {
	db := NewMemDb()

	// SETNX mykey "Hello"
	setnx := setNxString(db, [][]byte{[]byte("setnx"), []byte("mykey"), []byte("Hello")})
	if !bytes.Equal(setnx.ToRedisFormat(), []byte(":1\r\n")) {
		t.Error("SETNX is not correct")
	}

	// SETNX mykey "World"
	setnx = setNxString(db, [][]byte{[]byte("setnx"), []byte("mykey"), []byte("World")})
	if !bytes.Equal(setnx.ToRedisFormat(), []byte(":0\r\n")) {
		t.Error("SETNX is not correct")
	}

	// GET mykey
	get := getString(db, [][]byte{[]byte("get"), []byte("mykey")})
	if !bytes.Equal(get.GetBytesData(), []byte("Hello")) {
		t.Error("GET is not correct")
	}

	// STRLEN mykey
	strlen := strLenString(db, [][]byte{[]byte("strlen"), []byte("mykey")})
	if !bytes.Equal(strlen.ToRedisFormat(), []byte(":5\r\n")) {
		t.Error("STRLEN is not correct")
	}

	// STRLEN nonexisting
	strlen = strLenString(db, [][]byte{[]byte("strlen"), []byte("nonexisting")})
	if !bytes.Equal(strlen.ToRedisFormat(), []byte(":0\r\n")) {
		t.Error("STRLEN is not correct")
	}
}

func TestStringsCommand8(t *testing.T) {
	db := NewMemDb()

	// SET mykey "10"
	setString(db, [][]byte{[]byte("set"), []byte("mykey"), []byte("10")})

	// INCR mykey
	incr := incrString(db, [][]byte{[]byte("incr"), []byte("mykey")})
	if !bytes.Equal(incr.ToRedisFormat(), []byte(":11\r\n")) {
		t.Error("INCR is not correct")
	}

	// INCR mykey
	decr := decrString(db, [][]byte{[]byte("decr"), []byte("mykey")})
	if !bytes.Equal(decr.ToRedisFormat(), []byte(":10\r\n")) {
		t.Error("DECR is not correct")
	}

	// INCRBY mykey 5
	incrby := incrByString(db, [][]byte{[]byte("incr"), []byte("mykey"), []byte("5")})
	if !bytes.Equal(incrby.ToRedisFormat(), []byte(":15\r\n")) {
		t.Error("INCRBY is not correct")
	}
}

func TestStringsCommand9(t *testing.T) {
	db := NewMemDb()

	// SET mykey 10.50
	setString(db, [][]byte{[]byte("set"), []byte("mykey"), []byte("10.50")})

	// INCRBYFLOAT mykey 0.1
	incrbyf := incrByFloatString(db, [][]byte{[]byte("incrbyfloat"), []byte("mykey"), []byte("0.1")})
	if !bytes.Equal(incrbyf.GetBytesData(), []byte("10.6")) {
		t.Error("INCRBYFLOAT is not correct")
	}

	// SET mykey 5.0e3
	setString(db, [][]byte{[]byte("set"), []byte("mykey"), []byte("5.0e3")})
	incrbyf = incrByFloatString(db, [][]byte{[]byte("incrbyfloat"), []byte("mykey"), []byte("2.0e2")})
	if !bytes.Equal(incrbyf.GetBytesData(), []byte("5200")) {
		t.Error("INCRBYFLOAT is not correct")
	}

}

func TestStringsCommand10(t *testing.T) {
	db := NewMemDb()

	// APPEND mykey "Hello"
	apd := appendString(db, [][]byte{[]byte("append"), []byte("mykey"), []byte("Hello")})
	if !bytes.Equal(apd.GetBytesData(), []byte("5")) {
		t.Error("APPEND is not correct")
	}

	// APPEND mykey " World"
	apd = appendString(db, [][]byte{[]byte("append"), []byte("mykey"), []byte(" World")})
	if !bytes.Equal(apd.GetBytesData(), []byte("11")) {
		t.Error("APPEND is not correct")
	}

	// GET mykey
	get := getString(db, [][]byte{[]byte("get"), []byte("mykey")})
	if !bytes.Equal(get.GetBytesData(), []byte("Hello World")) {
		t.Error("GET is not correct")
	}

}
