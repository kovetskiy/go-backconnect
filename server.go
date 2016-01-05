package backconnect

import (
	"fmt"
	"net"
	"sync"
)

type ServeFunc func(*net.TCPConn, error)

type Server struct {
	address         string
	listener        *net.TCPListener
	listenerActions *sync.Mutex
	listening       bool
}

func NewServer() (*Server, error) {
	server := &Server{
		listenerActions: &sync.Mutex{},
	}

	return server, nil
}

func (server *Server) Listen(address string) error {
	tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return fmt.Errorf("can't resolve address %s: %s", address, err)
	}

	listener, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return fmt.Errorf("can't listen on %s: %s", tcpaddr, err)
	}

	server.address = address

	server.listenerActions.Lock()
	defer server.listenerActions.Unlock()

	err = server.Close()
	if err != nil {
		return fmt.Errorf(
			"can't stop listening on %s: %s",
			server.listener.Addr(), err,
		)
	}

	server.listener = listener
	server.listening = true

	return nil
}

func (server *Server) Close() error {
	if !server.listening {
		return nil
	}

	server.listening = false
	return server.listener.Close()
}

func (server *Server) Addr() net.Addr {
	return server.listener.Addr()
}

func (server *Server) Serve(callback ServeFunc) {
	for {
		if !server.listening {
			break
		}

		server.listenerActions.Lock()
		connection, err := server.listener.AcceptTCP()
		server.listenerActions.Unlock()

		go callback(connection, err)
	}
}
