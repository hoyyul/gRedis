package memdb

import "gRedis/protocol"

var CmdTable = make(map[string]*command)

// 返回客户端一个redis data类型
type cmdExecutor func(db *MemDb, cmd [][]byte) protocol.RedisData

type command struct {
	executor cmdExecutor
}

func RegisterCommand(cmdName string, executor cmdExecutor) {
	CmdTable[cmdName] = &command{
		executor: executor,
	}
}
