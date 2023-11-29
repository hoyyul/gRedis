package server

import (
	"gRedis/config"
	"gRedis/logger"
	"net"
	"strconv"
)

func Start(config *config.Config) {
	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		logger.Panic(err)
		return
	}
	defer listener.Close()

	handler := NewHandler()
	for {
		conn, err := listener.Accept()
		if err != nil {
			conn.Close()
			logger.Panic(err)
		}

		go handler.handle(conn)
	}
}
