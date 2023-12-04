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
	Data string // "OK"
}

type SimpleError struct {
	Data string // "ERROR"
}

type Integer struct {
	Data int64 // -15, 20
}

// binary strings
type BulkString struct {
	Data []byte // bytes("OK")
}

type Array struct {
	Data []RedisData // [SimpleString("OK"), SimpleError("ERROR"), Integer(-15), BulkString(bytes("OK"))]
}

// SimpleString
func NewSimpleString(data string) *SimpleString {
	return &SimpleString{
		Data: data,
	}
}

func (s *SimpleString) GetData() string {
	return s.Data
}

func (s *SimpleString) GetBytesData() []byte {
	return []byte(s.Data)
}

func (s *SimpleString) ToRedisFormat() []byte {
	return []byte(fmt.Sprintf("+%s%s", s.Data, CRLF)) // +OK\r\n
}

func (s *SimpleString) String() string {
	return s.Data
}

// SimpleError
func NewSimpleError(data string) *SimpleError {
	return &SimpleError{
		Data: data,
	}
}

func (e *SimpleError) GetData() string {
	return e.Data
}

func (e *SimpleError) GetBytesData() []byte {
	return []byte(e.Data)
}

func (e *SimpleError) ToRedisFormat() []byte {
	return []byte(fmt.Sprintf("-%s%s", e.Data, CRLF)) // -Error message\r\n
}

func (e *SimpleError) String() string {
	return e.Data
}

// Integer
func NewInteger(data int64) *Integer {
	return &Integer{
		Data: data,
	}
}

func (i *Integer) GetData() int64 {
	return i.Data
}

func (i *Integer) GetBytesData() []byte {
	return []byte(strconv.FormatInt(i.Data, 10)) // +42 -> "42", -42 -> "-42"
}

func (i *Integer) ToRedisFormat() []byte {
	return []byte(fmt.Sprintf(":%s%s", strconv.FormatInt(i.Data, 10), CRLF)) // [<+|->]<value>\r\n
}

func (i *Integer) String() string {
	return strconv.FormatInt(i.Data, 10)
}

// Bulk String
func NewBulkString(data []byte) *BulkString {
	return &BulkString{
		Data: data,
	}
}

func (bs *BulkString) GetData() []byte {
	return bs.Data
}

func (bs *BulkString) GetBytesData() []byte {
	return bs.Data
}

func (bs *BulkString) ToRedisFormat() []byte {
	if bs.Data == nil {
		return []byte("$-1\r\n")
	}
	return []byte("$" + strconv.Itoa(len(bs.Data)) + CRLF + string(bs.Data) + CRLF)
}

func (bs *BulkString) String() string {
	return string(bs.Data)
}

// Array
func NewArray(data []RedisData) *Array {
	return &Array{
		Data: data,
	}
}

func (a *Array) GetData() []RedisData {
	return a.Data
}

func (a *Array) GetBytesData() []byte {
	arr := make([]byte, 0, len(a.GetData()))
	for i := range a.Data {
		arr = append(arr, a.Data[i].GetBytesData()...)
	}
	return arr
}

func (a *Array) ToRedisFormat() []byte {
	if a.Data == nil {
		return []byte("*-1\r\n")
	}

	arr := []byte(fmt.Sprintf("*%d%s", len(a.Data), CRLF))

	for i := range a.Data {
		arr = append(arr, a.Data[i].ToRedisFormat()...)
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
