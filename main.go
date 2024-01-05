package main

import (
	"fmt"
	"gRedis/config"
	"gRedis/logger"
	"gRedis/memdb"
	"gRedis/server"
	"os"
)

func init() {
	memdb.RegisterKeyCommands()
	memdb.RegisterStringCommands()
	memdb.RegisterHashCommands()
	memdb.RegisterListCommands()
	memdb.RegisterSetCommands()
}

func main() {
	cfg, err := config.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = logger.Init(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = server.Start(config.Conf)
	if err != nil {
		os.Exit(1)
	}
}
