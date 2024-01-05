package server

import (
	"context"
	"gRedis/config"
	"gRedis/logger"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// start a redis server
func Start(config *config.Config) error {
	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		logger.Panic(err)
		return err
	}

	// gracefully shut down server
	defer func() {
		logger.Info("Shutting down server...")

		_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := listener.Close()
		if err != nil {
			logger.Error("failed to close server gracefully. Error: ", err)
		}
	}()

	logger.Info("Server Listen at ", config.Host, ":", config.Port)

	// handle signal termination
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// client chan
	clients := make(chan net.Conn)

	// create a resource manager
	mgr := NewManager(config)

	// start a goroutine to accept client connection
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Error(err)
				continue
			}
			clients <- conn
		}
	}()

	for {
		select {
		// start a go routine to handle client request
		case conn := <-clients:
			logger.Info(conn.RemoteAddr().String(), " connected")
			go mgr.Handle(conn)
		// exit server
		case <-osSignals:
			return nil
		}
	}
}
