package parser

import (
	"errors"
	"fmt"
	"gRedis/logger"
	"strconv"
)

func parseSingleLine(msg []byte) (RedisData, error) {
	if len(msg) < 3 {
		logger.Error("Protocal error: format invalid")
		return nil, errors.New("Protocal error: format invalid")
	}

	msgType := msg[0]
	msgData := string(msg[1 : len(msg)-2]) // discaed "\r\n"
	var resp RedisData

	switch msgType {
	case '+':
		resp = NewSimpleString(msgData)
	case '-':
		resp = NewSimpleError(msgData)
	case ':':
		integerData, err := strconv.ParseInt(msgData, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Protocal error: %s", string(msg)))
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
