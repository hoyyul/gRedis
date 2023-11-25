package server

import (
	"gRedis/config"
	"gRedis/logger"
	"net"
	"strconv"
)

func Init(config *config.Config) {
	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	defer listener.Close()

	if err != nil {
		logger.Panic(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Panic(err)
		}

		// go handler(conn)
	}
}
