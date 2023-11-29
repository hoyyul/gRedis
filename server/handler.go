package server

import (
	"bufio"
	"gRedis/logger"
	"io"
	"net"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) handle(conn net.Conn) {
	defer conn.Close()

	// Create a reader to read lines from the client
	reader := bufio.NewReader(conn)

	for {
		// Read a line (request) from the client
		_, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Client closed the connection, exit the loop
				logger.Info("connection close")
				return
			}
			logger.Panic(err)
			return
		}

		// parse request
		// todo...

		// excute parsed command
		// todo...

		// get response
		response := "ERROR: Unsupported command\r\n" // Example response

		// Send the response back to the client
		_, err = conn.Write([]byte(response))
		if err != nil {
			logger.Panic(err)
			return
		}
	}
}
