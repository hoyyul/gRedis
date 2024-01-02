package main

import (
	"gRedis/config"
	"gRedis/logger"
	"gRedis/memdb"
	"gRedis/server"
)

func init() {
	memdb.RegisterKeyCommands()
	memdb.RegisterStringCommands()
	memdb.RegisterHashCommands()
	memdb.RegisterListCommands()
	memdb.RegisterSetCommands()
}

func main() {
	config.Init()
	logger.Init(config.Conf)
	server.Start(config.Conf)
}
