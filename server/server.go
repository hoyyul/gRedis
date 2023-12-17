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

	logger.Info("Server Listen at ", config.Host, ":", config.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}

		logger.Info(conn.RemoteAddr().String(), " connected")

		handler := NewHandler()
		go func(conn net.Conn) {
			defer conn.Close()
			handler.Handle(conn)
		}(conn)
	}
}
