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

	resp1, err := parseSingleLine(msg1)
	if !bytes.Equal(resp1.GetBytesData(), []byte("OK")) || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg1)))
	}
	resp2, err := parseSingleLine(msg2)
	if !bytes.Equal(resp2.GetBytesData(), []byte("Error message")) || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg2)))
	}
	resp3, err := parseSingleLine(msg3)
	if !bytes.Equal(resp3.GetBytesData(), []byte("1000")) || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg3)))
	}
	resp4, err := parseSingleLine(msg4)
	if !bytes.Equal(resp4.GetBytesData(), []byte("-20")) || err != nil {
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

	resp1, err := parseBulkString(msg1)
	if !bytes.Equal(resp1.(*BulkString).data, []byte("hello")) || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg1)))
	}
	resp2, err := parseBulkString(msg2)
	if !bytes.Equal(resp2.(*BulkString).data, []byte("")) || err != nil {
		t.Error(fmt.Sprintf("Protocol error: %s", string(msg2)))
	}
}
