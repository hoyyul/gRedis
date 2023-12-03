package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"
)

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
	if data1.(*SimpleString).data != ss.data || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg1)))
	}
	data2, err := parseSingleLine(msg2)
	if data2.(*SimpleError).data != se.data || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg2)))
	}
	data3, err := parseSingleLine(msg3)
	if data3.(*Integer).data != i1.data || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg3)))
	}
	data4, err := parseSingleLine(msg4)
	if data4.(*Integer).data != i2.data || err != nil {
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
	if buf2.stringLen != 0 || buf2.multiLine || err != nil {
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
	if !bytes.Equal(data1.(*BulkString).data, bs1.data) || err != nil {
		t.Error(fmt.Sprintf("Protocol error. data: %v, expect: %v", data1.(*BulkString).data, bs1.data))
	}
	data2, err := parseBulkString(msg2)
	if !bytes.Equal(data2.(*BulkString).data, bs2.data) || err != nil {
		t.Error(fmt.Sprintf("Protocol error. data: %v, expect: %v", data2.(*BulkString).data, bs2.data))
	}
}

func TestParseArrayHeader(t *testing.T) {
	msg1 := []byte("*0\r\n")
	msg2 := []byte("*2\r\n")
	msg3 := []byte("*-1\r\n")
	buf1 := &readBuffer{}
	buf2 := &readBuffer{}
	buf3 := &readBuffer{}

	err := parseArrayHeader(msg1, buf1)
	if buf1.arrayLen != 0 || buf1.inArray || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg1)))
	}
	err = parseArrayHeader(msg2, buf2)
	if buf2.arrayLen != 2 || !buf2.inArray || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg2)))
	}
	err = parseArrayHeader(msg3, buf3)
	if buf3.arrayLen != -1 || buf3.inArray || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg3)))
	}
}
