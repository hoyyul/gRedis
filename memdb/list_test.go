package memdb

import (
	"bytes"
	"gRedis/config"
	"gRedis/resp"
	"testing"
)

func init() {
	config.Conf = &config.Config{
		SegNum: 100,
	}
}

func TestLPosList(t *testing.T) {
	m := NewMemDb()
	lPushList(m, [][]byte{[]byte("lpush"), []byte("l1"), []byte("d"), []byte("b"), []byte("a"), []byte("c"), []byte("b"), []byte("a")})

	var res resp.RedisData
	//    test normal pos
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("a")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewInteger(0).ToRedisFormat()) {
		t.Error("normal lpos error")
	}
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("d")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewInteger(5).ToRedisFormat()) {
		t.Error("normal lpos error")
	}

	// test rank option
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("a"), []byte("rank"), []byte("2")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewInteger(3).ToRedisFormat()) {
		t.Error("positive rank lpos error")
	}
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("b"), []byte("rank"), []byte("-2")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewInteger(1).ToRedisFormat()) {
		t.Error("negative rank lpos error")
	}

	//     test count option
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("a"), []byte("count"), []byte("2")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewArray([]resp.RedisData{resp.NewInteger(0), resp.NewInteger(3)}).ToRedisFormat()) {
		t.Error("count lpos error")
	}
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("c"), []byte("count"), []byte("1")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewArray([]resp.RedisData{resp.NewInteger(2)}).ToRedisFormat()) {
		t.Error("count lpos error")
	}
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("b"), []byte("count"), []byte("1"), []byte("rank"), []byte("-1")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewArray([]resp.RedisData{resp.NewInteger(4)}).ToRedisFormat()) {
		t.Error("count lpos error")
	}

	//    test maxlen option
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("a"), []byte("maxlen"), []byte("2"), []byte("count"), []byte("0")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewArray([]resp.RedisData{resp.NewInteger(0)}).ToRedisFormat()) {
		t.Error("maxlen lpos error")
	}
	res = lPosList(m, [][]byte{[]byte("lpos"), []byte("l1"), []byte("d"), []byte("maxlen"), []byte("3")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewBulkString(nil).ToRedisFormat()) {
		t.Error("maxlen lpos error")
	}
}

func TestLRemList(t *testing.T) {
	m := NewMemDb()
	rPushList(m, [][]byte{[]byte("rpush"), []byte("l1"), []byte("0"), []byte("1"), []byte("1"), []byte("1"), []byte("2"), []byte("2")})

	var res resp.RedisData
	res = lRemList(m, [][]byte{[]byte("lrem"), []byte("l1"), []byte("0"), []byte("0")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewInteger(1).ToRedisFormat()) {
		t.Error("lrem error")
	}
	res = lRemList(m, [][]byte{[]byte("lrem"), []byte("l1"), []byte("2"), []byte("1")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewInteger(2).ToRedisFormat()) {
		t.Error("lrem error")
	}
	res = lRemList(m, [][]byte{[]byte("lrem"), []byte("l1"), []byte("0"), []byte("2")})
	if !bytes.Equal(res.ToRedisFormat(), resp.NewInteger(2).ToRedisFormat()) {
		t.Error("lrem error")
	}
}
