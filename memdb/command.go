package memdb

import "gRedis/resp"

var CmdTable = make(map[string]*command)

// 返回客户端一个redis data类型
type cmdExecutor func(db *MemDb, cmd [][]byte) resp.RedisData

type command struct {
	Executor cmdExecutor
}

func RegisterCommand(cmdName string, executor cmdExecutor) {
	CmdTable[cmdName] = &command{
		Executor: executor,
	}
}
