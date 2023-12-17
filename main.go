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
}

func main() {
	config.Init()
	logger.Init(config.Conf)
	server.Start(config.Conf)
}
