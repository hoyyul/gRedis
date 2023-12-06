package server

import (
	"gRedis/logger"
	"gRedis/protocol"
	"io"
	"net"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(conn net.Conn) {
	defer conn.Close()

	// parse conn
	ch := protocol.ParseStream(conn)
	for resp := range ch {
		if resp.Err != nil {
			if resp.Err != io.EOF {
				logger.Panic("Connection: ", conn.RemoteAddr().String(), ", Panic: ", resp.Err)
			} else {
				logger.Info("Close connection: ", conn.RemoteAddr().String())
			}
			return
		}

		if resp.Data == nil {
			logger.Error("Get empty array from: ", conn.RemoteAddr().String())
			continue
		}

		// get parsed data
		arrayData, ok := resp.Data.(*protocol.Array)
		if !ok {
			logger.Error("Data from connection: ", conn.RemoteAddr().String(), "is not a valid array")
			continue
		}

		// excute parsed command
		command := arrayData.ToCommand()
		// todo... get response
		response := "ERROR: Unsupported command\r\n"

		// Send the response back to the client
		_, err = conn.Write([]byte(response))
		if err != nil {
			logger.Panic(err)
			return
		}
	}
}
