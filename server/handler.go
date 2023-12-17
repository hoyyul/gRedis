package server

import (
	"gRedis/logger"
	"gRedis/memdb"
	"gRedis/resp"
	"io"
	"net"
)

type Handler struct {
	db *memdb.MemDb
}

func NewHandler() *Handler {
	return &Handler{db: memdb.NewMemDb()}
}

func (h *Handler) Handle(conn net.Conn) {
	// parse conn
	ch := resp.ParseStream(conn)
	for redisResp := range ch {
		if redisResp.Err != nil {
			if redisResp.Err != io.EOF {
				logger.Panic("Connection: ", conn.RemoteAddr().String(), ", Panic: ", redisResp.Err)
			} else {
				logger.Info("Close connection: ", conn.RemoteAddr().String())
			}
			return
		}

		if redisResp.Data == nil {
			logger.Error("Get empty data from: ", conn.RemoteAddr().String())
			continue
		}

		// get parsed data
		arrayData, ok := redisResp.Data.(*resp.RedisArray)
		if !ok {
			logger.Error("Data from connection: ", conn.RemoteAddr().String(), "is not a valid array")
			continue
		}

		// excute parsed command
		cmd := arrayData.ToCommand()
		redisData := h.db.ExecCommand(cmd)

		if redisData != nil {
			_, err := conn.Write(redisData.ToRedisFormat())
			if err != nil {
				logger.Error("write response to ", conn.RemoteAddr().String(), " error: ", err.Error())
			}
		} else {
			errData := resp.NewSimpleError("unknown error")
			_, err := conn.Write(errData.ToRedisFormat())
			if err != nil {
				logger.Error("write response to ", conn.RemoteAddr().String(), " error: ", err.Error())
			}
		}
	}
}
