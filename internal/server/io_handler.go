package server

import (
	"Nietzsche/internal/core"
	"Nietzsche/internal/core/io_multiplexing"
	"io"
	"log"
	"sync"
	"syscall"
)

type IOHandler struct {
	id            int
	ioMultiplexer io_multiplexing.IOMultiplexer
	mu            sync.Mutex
	server        *Server
}

func NewIOHandler(id int, server *Server) (*IOHandler, error) {
	multiplexer, err := io_multiplexing.CreateIOMultiplexer()
	if err != nil {
		return nil, err
	}

	return &IOHandler{
		id:            id,
		ioMultiplexer: multiplexer,
		server:        server,
	}, nil
}

func (h *IOHandler) AddConn(connFd int) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	log.Printf("I/O Handler %d is monitoring fd %d", h.id, connFd)
	return h.ioMultiplexer.Monitor(io_multiplexing.Event{
		Fd: connFd,
		Op: io_multiplexing.OpRead,
	})
}

func (h *IOHandler) Run() {
	log.Printf("I/O Handler %d started", h.id)
	for {
		events, err := h.ioMultiplexer.Wait()
		if err != nil {
			continue
		}

		for _, event := range events {
			connFd := event.Fd
			cmd, err := readCommand(connFd)
			if err != nil {
				if err == io.EOF || err == syscall.ECONNRESET {
					log.Println("client disconnected")
					_ = syscall.Close(connFd)
					continue
				}
				log.Println("read error:", err)
				continue
			}

			replyCh := make(chan []byte, 1)
			task := &core.Task{
				Command: cmd,
				ReplyCh: replyCh,
			}
			h.server.dispatch(task)
			res := <-replyCh
			syscall.Write(connFd, res)
		}
	}
}
