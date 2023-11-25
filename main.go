package main

import (
	"gRedis/config"
	"gRedis/logger"
)

func init() {

}

func main() {
	config.Init()
	logger.Init(config.Conf)
}
