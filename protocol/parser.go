package protocol

import (
	"errors"
	"fmt"
	"gRedis/logger"
	"strconv"
)

type readBuffer struct {
	stringLen int64 // bulk string
	multiLine bool
}

func parseSingleLine(msg []byte) (RedisData, error) {
	if len(msg) < 3 {
		logger.Error("Protocol error: format invalid")
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
			logger.Error(fmt.Sprintf("Protocol error error: %s", string(msg)))
			return nil, err
		}
		resp = NewInteger(integerData)
	}
	if resp == nil {
		logger.Error("Protocol error: msg is null")
		return nil, errors.New("Protocol error: " + string(msg))
	}

	return resp, nil
}

func parseBulkStringHeader(msg []byte, buf *readBuffer) error {
	// $5\r\n
	if len(msg) < 3 {
		logger.Error("Protocol error: format invalid")
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
	}
	// stringLen == 1 means null string.
	return nil
}

func parseBulkString(msg []byte) (RedisData, error) {
	if len(msg) < 2 {
		logger.Error("Protocol error: format invalid")
		return nil, errors.New("Protocol error: format invalid")
	}

	// discard "\r\n"
	msgData := msg[:len(msg)-2]
	resp := NewBulkString(msgData)

	return resp, nil
}
