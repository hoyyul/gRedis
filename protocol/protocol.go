package protocol

// reference: https://redis.io/docs/reference/protocol-spec/
/*
To communicate with the Redis server,
Redis clients use a protocol called REdis Serialization Protocol (RESP).
*/

import (
	"fmt"
	"strconv"
)

var (
	CRLF string = "\r\n"
)

type RedisData interface {
	GetBytesData() []byte
	ToRedisFormat() []byte
}

// non-binary strings
type SimpleString struct {
	data string // "OK"
}

type SimpleError struct {
	data string // "ERROR"
}

type Integer struct {
	data int64 // -15, 20
}

// binary strings
type BulkString struct {
	data []byte // bytes("OK")
}

type Array struct {
	data []RedisData // [SimpleString("OK"), SimpleError("ERROR"), Integer(-15), BulkString(bytes("OK"))]
}

// SimpleString
func NewSimpleString(data string) *SimpleString {
	return &SimpleString{
		data: data,
	}
}

func (s *SimpleString) GetData() string {
	return s.data
}

func (s *SimpleString) GetBytesData() []byte {
	return []byte(s.data)
}

func (s *SimpleString) ToRedisFormat() []byte {
	return []byte(fmt.Sprintf("+%s%s", s.data, CRLF)) // +OK\r\n
}

// SimpleError
func NewSimpleError(data string) *SimpleError {
	return &SimpleError{
		data: data,
	}
}

func (e *SimpleError) GetData() string {
	return e.data
}

func (e *SimpleError) GetBytesData() []byte {
	return []byte(e.data)
}

func (e *SimpleError) ToRedisFormat() []byte {
	return []byte(fmt.Sprintf("-%s%s", e.data, CRLF)) // -Error message\r\n
}

// Integer
func NewInteger(data int64) *Integer {
	return &Integer{
		data: data,
	}
}

func (i *Integer) GetData() int64 {
	return i.data
}

func (i *Integer) GetBytesData() []byte {
	return []byte(strconv.FormatInt(i.data, 10)) // +42 -> "42", -42 -> "-42"
}

func (i *Integer) ToRedisFormat() []byte {
	return []byte(fmt.Sprintf(":%s%s", strconv.FormatInt(i.data, 10), CRLF)) // [<+|->]<value>\r\n
}

// Bulk String
func NewBulkString(data []byte) *BulkString {
	return &BulkString{
		data: data,
	}
}

func (bs *BulkString) GetData() []byte {
	return bs.data
}

func (bs *BulkString) GetBytesData() []byte {
	return bs.data
}

func (bs *BulkString) ToRedisFormat() []byte {
	return bs.data
}

// Array
func NewArray(data []RedisData) *Array {
	return &Array{
		data: data,
	}
}

func (a *Array) GetData() []RedisData {
	return a.data
}

func (a *Array) GetBytesData() []byte {
	arr := make([]byte, 0, len(a.data))
	for i := range a.data {
		arr = append(arr, a.data[i].GetBytesData()...)
	}
	return arr
}

func (a *Array) ToRedisFormat() []byte {
	if a.data == nil {
		return []byte("*-1\r\n")
	}

	arr := []byte(fmt.Sprintf("*%d%s", len(a.data), CRLF))

	for i := range a.data {
		arr = append(arr, a.data[i].ToRedisFormat()...)
	}
	return arr
}
