package protocol

// reference: https://redis.io/docs/reference/protocol-spec/
/*
To communicate with the Redis server,
Redis clients use a protocol called REdis Serialization Protocol (RESP).
*/

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	CRLF string = "\r\n"
)

type RedisData interface {
	GetBytesData() []byte
	ToRedisFormat() []byte
	String() string
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

func (s *SimpleString) String() string {
	return s.data
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

func (e *SimpleError) String() string {
	return e.data
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

func (i *Integer) String() string {
	return strconv.FormatInt(i.data, 10)
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
	if bs.data == nil {
		return []byte("$-1\r\n")
	}
	return []byte("$" + strconv.Itoa(len(bs.data)) + CRLF + string(bs.data) + CRLF)
}

func (bs *BulkString) String() string {
	return string(bs.data)
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
	arr := make([]byte, 0, len(a.GetData()))
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

func (a *Array) String() string {
	return strings.Join(a.ToStringCommand(), " ")
}

func (a *Array) ToCommand() [][]byte {
	arr := make([][]byte, 0, len(a.GetData()))
	for i := range a.GetData() {
		arr = append(arr, a.GetData()[i].GetBytesData())
	}
	return arr
}

func (a *Array) ToStringCommand() []string {
	arr := make([]string, 0, len(a.GetData()))
	for i := range a.GetData() {
		arr = append(arr, a.GetData()[i].String())
	}
	return arr
}
