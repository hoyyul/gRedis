package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestParseStream(t *testing.T) {
	var b []byte
	var reader io.Reader
	var ch <-chan *RedisResp

	// Null elements in array
	b = []byte("*3\r\n$5\r\nhello\r\n$-1\r\n$5\r\nworld\r\n")
	reader = bytes.NewReader(b)
	ch = ParseStream(reader)
	for resp := range ch {
		if resp.Err != nil {
			if resp.Err != io.EOF {
				t.Error(resp.Err)
			}
			break
		}

		arr := resp.Data.(*Array).GetData()
		if !bytes.Equal(arr[0].(*BulkString).GetData(), []byte("hello")) {
			t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr[0].GetBytesData(), []byte("hello")))
		}
		if arr[1].(*BulkString).GetData() != nil {
			t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr[1].GetBytesData(), nil))
		}
		if !bytes.Equal(arr[2].(*BulkString).GetData(), []byte("world")) {
			t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr[2].GetBytesData(), []byte("world")))
		}
	}

	// Null array
	b = []byte("*-1\r\n")
	reader = bytes.NewReader(b)
	ch = ParseStream(reader)
	for resp := range ch {
		if resp.Err != nil {
			if resp.Err != io.EOF {
				t.Error(resp.Err)
			}
			break
		}

		arr := resp.Data.(*Array)
		if arr.GetData() != nil || !bytes.Equal(arr.ToRedisFormat(), []byte("*-1\r\n")) {
			t.Error("Stream error.")
		}
	}

	// Empty array
	b = []byte("*0\r\n")
	reader = bytes.NewReader(b)
	ch = ParseStream(reader)
	for resp := range ch {
		if resp.Err != nil {
			if resp.Err != io.EOF {
				t.Error(resp.Err)
			}
			break
		}

		arr := resp.Data.(*Array)
		if len(arr.GetData()) != 0 || !bytes.Equal(arr.ToRedisFormat(), []byte("*0\r\n")) {
			t.Error("Stream error.")
		}
	}

	// Bulk string
	b = []byte("$5\r\nhello\r\n$-1\r\n$5\r\nworld\r\n")
	reader = bytes.NewReader(b)
	ch = ParseStream(reader)
	i := 0
	for resp := range ch {
		if resp.Err != nil {
			if resp.Err != io.EOF {
				t.Error(resp.Err)
			}
			break
		}

		bs := resp.Data.(*BulkString)
		if i == 0 {
			if !bytes.Equal(bs.GetData(), []byte("hello")) {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", bs.GetData(), []byte("hello")))
			}
		}
		if i == 1 {
			if bs.GetData() != nil {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", bs.GetData(), nil))
			}
		}
		if i == 2 {
			if !bytes.Equal(bs.GetData(), []byte("world")) {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", bs.GetData(), []byte("world")))
			}
		}
		i++
	}

	// Nested Array
	b = []byte("*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n") // send [1 2 3] and [Hello World] respectively
	reader = bytes.NewReader(b)
	ch = ParseStream(reader)
	k := 0
	for resp := range ch {
		if resp.Err != nil {
			if resp.Err != io.EOF {
				t.Error(resp.Err)
			}
			break
		}

		arr := resp.Data.(*Array)
		if k == 0 {
			if arr.GetData()[0].(*Integer).GetData() != 1 {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr.GetData()[0].(*Integer).GetData(), 1))
			}
			if arr.GetData()[1].(*Integer).GetData() != 2 {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr.GetData()[1].(*Integer).GetData(), 2))
			}
			if arr.GetData()[2].(*Integer).GetData() != 3 {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr.GetData()[2].(*Integer).GetData(), 3))
			}
		}
		if k == 1 {
			if arr.GetData()[0].(*SimpleString).GetData() != "Hello" {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr.GetData()[0].(*SimpleString).GetData(), "Hello"))
			}
			if arr.GetData()[1].(*SimpleError).GetData() != "World" {
				t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", arr.GetData()[1].(*SimpleError).GetData(), "World"))
			}
		}
		k++
	}
}

func TestReadLine(t *testing.T) {
	b := []byte("+OK\r\n:-2\r\n")
	ioReader := bytes.NewReader(b)
	bufioReader := bufio.NewReader(ioReader)
	buf := &readBuffer{}

	// simple message
	for i := 0; ; i++ {
		msg, err := readline(bufioReader, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Error("Stream error")
		}
		if !bytes.Equal(msg, b[i*5:(i+1)*5]) {
			t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", msg, b[i*5:(i+1)*5]))
		}
	}

	// bulk string message
	buf.stringLen = 7
	buf.multiLine = true
	b = []byte("1\r\n2\n34\r\n")
	ioReader = bytes.NewReader(b)
	bufioReader = bufio.NewReader(ioReader)
	msg, err := readline(bufioReader, buf)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(msg, b) {
		t.Error(fmt.Sprintf("Stream error. msg: %v expect %v", msg, b))
	}
}

func TestParseSingleLine(t *testing.T) {
	msg1 := []byte("+OK\r\n")
	msg2 := []byte("-Error message\r\n")
	msg3 := []byte(":1000\r\n")
	msg4 := []byte(":-20\r\n")
	ss := NewSimpleString("OK")
	se := NewSimpleError("Error message")
	i1 := NewInteger(1000)
	i2 := NewInteger(-20)

	data1, err := parseSingleLine(msg1)
	if data1.(*SimpleString).Data != ss.Data || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg1)))
	}
	data2, err := parseSingleLine(msg2)
	if data2.(*SimpleError).Data != se.Data || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg2)))
	}
	data3, err := parseSingleLine(msg3)
	if data3.(*Integer).Data != i1.Data || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg3)))
	}
	data4, err := parseSingleLine(msg4)
	if data4.(*Integer).Data != i2.Data || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg4)))
	}
}

func TestParseBulkStringHeader(t *testing.T) {
	msg1 := []byte("$5\r\n")
	msg2 := []byte("$-1\r\n")
	msg3 := []byte("$-2\r\n")
	buf1 := &readBuffer{}
	buf2 := &readBuffer{}
	buf3 := &readBuffer{}

	err := parseBulkStringHeader(msg1, buf1)
	if buf1.stringLen != 5 || !buf1.multiLine || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg1)))
	}
	err = parseBulkStringHeader(msg2, buf2)
	if buf2.stringLen != -1 || buf2.multiLine || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg2)))
	}
	err = parseBulkStringHeader(msg3, buf3)
	if buf3.stringLen != 0 || buf3.multiLine || err == nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg3)))
	}
}

func TestParseBulkString(t *testing.T) {
	msg1 := []byte("hello\r\n")
	msg2 := []byte("\r\n")
	bs1 := NewBulkString([]byte("hello"))
	bs2 := NewBulkString([]byte(""))

	data1, err := parseBulkString(msg1)
	if !bytes.Equal(data1.(*BulkString).Data, bs1.Data) || err != nil {
		t.Error(fmt.Sprintf("Protocol error. data: %v, expect: %v", data1.(*BulkString).Data, bs1.Data))
	}
	data2, err := parseBulkString(msg2)
	if !bytes.Equal(data2.(*BulkString).Data, bs2.Data) || err != nil {
		t.Error(fmt.Sprintf("Protocol error. data: %v, expect: %v", data2.(*BulkString).Data, bs2.Data))
	}
}

func TestParseArrayHeader(t *testing.T) {
	msg1 := []byte("*0\r\n")
	msg2 := []byte("*1\r\n")
	msg3 := []byte("*-1\r\n")
	buf1 := &readBuffer{}
	buf2 := &readBuffer{}
	buf3 := &readBuffer{}

	err := parseArrayHeader(msg1, buf1)
	if buf1.arrayLen != 0 || buf1.inArray || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg1)))
	}
	err = parseArrayHeader(msg2, buf2)
	if buf2.arrayLen != 1 || !buf2.inArray || buf2.arrayData == nil || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg2)))
	}
	err = parseArrayHeader(msg3, buf3)
	if buf3.arrayLen != -1 || buf3.inArray || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg3)))
	}
}
