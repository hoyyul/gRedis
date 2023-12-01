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
	CR   string = "\r"
	LF   string = "\n"
	CRLF string = "\r\n"
)

type RedisData interface {
	GetBytesData() []byte
	ToRedisFormat() []byte
}

type SimpleString struct {
	data string
}

type SimpleError struct {
	data string
}

type Integer struct {
	data int64
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
