package server

import (
	"Nietzsche/internal/core"
	"Nietzsche/internal/core/io_multiplexing"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
)

type IOHandler struct {
	id            int
	ioMultiplexer io_multiplexing.IOMultiplexer
	mu            sync.Mutex
	server        *Server
	conns         map[int]net.Conn
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
		conns:         make(map[int]net.Conn), // map from fd to corresponding connection
	}, nil
}

func (h *IOHandler) AddConn(conn net.Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	tcpConn := conn.(*net.TCPConn)
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return err
	}

	var connFd int
	err = rawConn.Control(func(fd uintptr) {
		connFd = int(fd)
		//log.Printf("I/O Handler %d is monitoring fd %d", h.id, connFd)
		// Store the connection object so it's not garbage collected
		h.conns[connFd] = conn
		// Add to epoll
		h.ioMultiplexer.Monitor(io_multiplexing.Event{
			Fd: connFd,
			Op: io_multiplexing.OpRead,
		})
	})

	return err
}

func (h *IOHandler) closeConn(fd int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.conns[fd]; ok {
		conn.Close()
		delete(h.conns, fd)
	}
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
			h.mu.Lock()
			conn, ok := h.conns[connFd]
			h.mu.Unlock()
			if !ok {
				// Connection might have been closed by a concurrent write error
				continue
			}
			//cmd, err := readCommand(connFd)
			cmd, err := readCommandConn(conn)
			if err != nil {
				if err == io.EOF || err == syscall.ECONNRESET {
					//log.Printf("Client disconnected (fd: %d)", connFd)
				} else {
					log.Printf("Read error on fd %d: %v", connFd, err)
				}
				h.closeConn(connFd) // <-- Use our new closing function
				continue
			}

			replyCh := make(chan []byte, 1)
			task := &core.Task{
				Command: cmd,
				ReplyCh: replyCh,
			}
			h.server.dispatch(task)
			res := <-replyCh
			conn.Write(res)
		}
	}
}
