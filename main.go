package main

import (
	"gRedis/config"
	"gRedis/logger"
	"gRedis/server"
)

func init() {

}

func main() {
	config.Init()
	logger.Init(config.Conf)
	server.Start(config.Conf)
}
