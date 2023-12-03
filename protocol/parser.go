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
	arrayLen  int64
	inArray   bool
	arrayData *Array
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
		var data RedisData
		msg, err := readline(streamReader, buf)
		if err != nil {
			// read all and close channel
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

		// make redis data
		if buf.multiLine {
			// bulk string msg
			data, err = parseBulkString(msg)
			buf.multiLine = false
			buf.stringLen = 0
		} else {
			// bulk string header
			if msg[0] == '$' {
				err = parseBulkStringHeader(msg, buf)

				if err != nil {
					logger.Error("Stream Error: ", err)
					ch <- &RedisResp{Err: err}
					buf = &readBuffer{}
				} else {
					if buf.stringLen == -1 { // null bulk string
						ch <- &RedisResp{Data: NewBulkString(nil)}
					}
				}
				continue
			}

			// array
			if msg[0] == '*' {
				err = parseArrayHeader(msg, buf)

				if err != nil {
					logger.Error("Stream Error: ", err)
					ch <- &RedisResp{Err: err}
					buf = &readBuffer{}
				} else {
					if buf.arrayLen == -1 { // null bulk string
						buf.arrayLen = 0
						ch <- &RedisResp{Data: NewArray(nil)}
					} else if buf.arrayLen == 0 {
						ch <- &RedisResp{Data: NewArray([]RedisData{})}
					}
				}

				continue
			}

			// simple message
			data, err = parseSingleLine(msg)
		}

		if err != nil {
			logger.Error("Stream Error: ", err)
			ch <- &RedisResp{Err: err}
			buf = &readBuffer{}
			continue
		}

		// send redis data
		ch <- &RedisResp{Data: data}
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
	msgType := msg[0]
	msgData := string(msg[1 : len(msg)-2]) // discard flag and "\r\n"
	var data RedisData

	switch msgType {
	case '+':
		data = NewSimpleString(msgData)
	case '-':
		data = NewSimpleError(msgData)
	case ':':
		integerData, err := strconv.ParseInt(msgData, 10, 64)
		if err != nil {
			return nil, err
		}
		data = NewInteger(integerData)
	default:
		// not valid?
	}
	if data == nil {
		return nil, errors.New("Protocol error: " + string(msg))
	}

	return data, nil
}

func parseBulkStringHeader(msg []byte, buf *readBuffer) error {
	// read header; discard flag and "\r\n"
	stringLen, err := strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if stringLen < -1 || err != nil {
		return errors.New("Protocol error: " + string(msg))
	}

	if stringLen > -1 {
		buf.multiLine = true
		buf.stringLen = stringLen
	}
	return nil
}

func parseBulkString(msg []byte) (RedisData, error) {
	// read data; discard "\r\n"
	msgData := msg[:len(msg)-2]
	data := NewBulkString(msgData)

	return data, nil
}

func parseArrayHeader(msg []byte, buf *readBuffer) error {
	// read header; discard flag and "\r\n"
	arrayLen, err := strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if arrayLen < -1 || err != nil {
		return errors.New("Protocol error: " + string(msg))
	}

	// len == -1 or len == 0 or len > 0
	buf.arrayLen = arrayLen

	// only for len > 0
	if arrayLen > 0 {
		buf.inArray = true
		buf.arrayData = NewArray([]RedisData{})
	}
	return nil
}
