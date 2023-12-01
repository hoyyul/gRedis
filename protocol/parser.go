package protocol

import (
	"bufio"
	"errors"
	"gRedis/logger"
	"io"
	"strconv"
)

type RedisResp struct {
	Data RedisData
	Err  error
}
type readBuffer struct {
	stringLen int64 // bulk string
	multiLine bool
}

func ParseStream(reader io.Reader) <-chan *RedisResp {
	ch := make(chan *RedisResp) // 双向通道，但是对外是只读通道
	go parse(reader, ch)
	return ch
}

func parse(reader io.Reader, ch chan *RedisResp) {
	streamReader := bufio.NewReader(reader)
	buf := &readBuffer{}

	for {
		var resp RedisData
		msg, err := readline(streamReader, buf)
		if err != nil {
			if err == io.EOF {
				ch <- &RedisResp{Err: err}
				close(ch)
				return
			}
			logger.Error("Stream Error: ", err)
			ch <- &RedisResp{Err: err}
			buf = &readBuffer{}
			continue
		}

		// simple msg or bulk string msg
		if buf.multiLine {
			// bulk string msg
			buf.multiLine = false
			buf.stringLen = 0
			resp, err = parseBulkString(msg)
		} else {
			// bulk string header
			if msg[0] == '$' {
				err = parseBulkStringHeader(msg, buf)

				if err != nil {
					logger.Error("Stream Error: ", err)
					ch <- &RedisResp{Err: err}
					buf = &readBuffer{}
				} else if buf.stringLen == -1 { // null bulk string
					resp = NewBulkString(nil)
					ch <- &RedisResp{Data: resp}
				}

				continue
			}

			// simple message
			resp, err = parseSingleLine(msg)
		}

		if err != nil {
			logger.Error("Stream Error: ", err)
			ch <- &RedisResp{Err: err}
			buf = &readBuffer{}
			continue
		}
		ch <- &RedisResp{Data: resp}

	}
}

func readline(reader *bufio.Reader, buf *readBuffer) (msg []byte, err error) {
	if buf.multiLine { // read bulk string
		msg = make([]byte, buf.stringLen+2)
		_, err = io.ReadFull(reader, msg)
		if err != nil {
			return nil, err
		}

		buf.stringLen = 0

		if msg[len(msg)-1] != '\n' || msg[len(msg)-2] != '\r' {
			return nil, errors.New("Protocol error: stream msg invalid")
		}
	} else {
		// read simple string.
		// \n is not allowed in simple string.
		// \n can be terminator for read line
		msg, err = reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		if msg[len(msg)-2] != '\r' {
			return nil, errors.New("Protocol error: stream msg invalid")
		}
	}
	return msg, nil
}

func parseSingleLine(msg []byte) (RedisData, error) {
	if len(msg) < 3 {
		return nil, errors.New("Protocol error: format invalid")
	}

	msgType := msg[0]
	msgData := string(msg[1 : len(msg)-2]) // discard "\r\n"
	var resp RedisData

	switch msgType {
	case '+':
		resp = NewSimpleString(msgData)
	case '-':
		resp = NewSimpleError(msgData)
	case ':':
		integerData, err := strconv.ParseInt(msgData, 10, 64)
		if err != nil {
			return nil, err
		}
		resp = NewInteger(integerData)
	}
	if resp == nil {
		return nil, errors.New("Protocol error: " + string(msg))
	}

	return resp, nil
}

func parseBulkStringHeader(msg []byte, buf *readBuffer) error {
	// $5\r\n
	if len(msg) < 3 {
		return errors.New("Protocol error: format invalid")
	}

	// discard "\r\n"
	stringLen, err := strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if stringLen < -1 || err != nil {
		return errors.New("Protocol error: " + string(msg))
	}

	if stringLen > -1 {
		buf.multiLine = true
		buf.stringLen = stringLen
	} else { // stringLen == 1 means null string.
		buf.multiLine = false
		buf.stringLen = 0
	}

	return nil
}

func parseBulkString(msg []byte) (RedisData, error) {
	if len(msg) < 2 {
		return nil, errors.New("Protocol error: format invalid")
	}

	// discard "\r\n"
	msgData := msg[:len(msg)-2]
	resp := NewBulkString(msgData)

	return resp, nil
}
