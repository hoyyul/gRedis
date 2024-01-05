package server

import (
	"fmt"
	"gRedis/config"
	"gRedis/logger"
	"gRedis/memdb"
	"gRedis/resp"
	"io"
	"net"
	"strconv"
	"strings"
)

type Manager struct {
	db  *memdb.MemDb
	dbs []*memdb.MemDb
}

func NewManager(config *config.Config) *Manager {
	dbs := make([]*memdb.MemDb, config.DbNum)
	for i := 0; i < len(dbs); i++ {
		dbs[i] = memdb.NewMemDb()
	}
	return &Manager{
		db:  dbs[0],
		dbs: dbs,
	}
}

func (m *Manager) Handle(conn net.Conn) {
	// parse conn
	ch := resp.ParseStream(conn)

	// close connection
	defer func() {
		err := conn.Close()
		if err != nil {
			logger.Error(err)
		}
	}()

	// read from client and pump redis to ch
	for redisResp := range ch {
		// hanle errs
		if redisResp.Err != nil {
			if redisResp.Err != io.EOF {
				logger.Panic("Connection: ", conn.RemoteAddr().String(), ", Panic: ", redisResp.Err)
			} else {
				logger.Info("Close connection: ", conn.RemoteAddr().String())
			}
			return
		}

		// get empty response
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
		redisData := m.ExecCommand(cmd)

		// write result to connection
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

func (m *Manager) ExecCommand(cmd [][]byte) resp.RedisData {
	cmdName := strings.ToLower(string((cmd[0])))
	command, ok := memdb.CmdTable[cmdName]

	if cmdName == "select" {
		return m.Select(cmd)
	}

	if ok {
		return command.Executor(m.db, cmd)
	} else {
		return resp.NewSimpleError(fmt.Sprintf("unknown command '%s'", string((cmd[0]))))
	}
}

func (m *Manager) Select(cmd [][]byte) resp.RedisData {
	if len(cmd) != 2 {
		return resp.NewSimpleError("wrong number of arguments for command")
	}

	dbIdx, err := strconv.Atoi(string(cmd[1]))
	if err != nil {
		return resp.NewSimpleError("value is not an integer")
	}

	if dbIdx >= len(m.dbs) || dbIdx < 0 {
		return resp.NewSimpleString(fmt.Sprintf("ERR DB index is out of range with maximum %d", len(m.dbs)))
	}

	m.db = m.dbs[dbIdx]

	return resp.NewSimpleString("OK")
}
