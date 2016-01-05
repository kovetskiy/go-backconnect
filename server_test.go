package backconnect

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Handler struct {
	Handled int
	Errors  []error
}

func (handler *Handler) Handle(conn *net.TCPConn, err error) {
	handler.Handled++

	if err != nil {
		handler.Errors = append(handler.Errors, err)
	}

	_, err = conn.Write([]byte("response"))
	if err != nil {
		handler.Errors = append(handler.Errors, err)
	}
}

func TestListen(t *testing.T) {
	server, err := NewServer()
	assert.NoError(t, err)

	err = server.Listen("localhost:12345")
	assert.NoError(t, err)

	handler := &Handler{}
	workers := &sync.WaitGroup{}

	go func() {
		server.Serve(func(conn *net.TCPConn, err error) {
			workers.Add(1)
			defer workers.Done()

			handler.Handle(conn, err)
		})
	}()

	dials := rand.Intn(50)
	for i := 1; i <= dials; i++ {
		dial(t, "localhost:12345", fmt.Sprintf("squancy-%d", i))
	}

	workers.Wait()

	assert.Equal(t, dials, handler.Handled, "mismatch handled count")
}

func dial(t *testing.T, address string, data ...string) {
	conn, err := net.Dial("tcp", "localhost:12345")
	if err != nil {
		t.Fatalf("can't dial to %s (%v): %s", address, data, err)
	}

	for _, line := range data {
		_, err = conn.Write([]byte(line))
		if err != nil {
			t.Fatalf("can't write to %s (%v): %s", address, data, err)
		}
	}

	_, err = conn.Read(nil)
	if err != nil && err != io.EOF {
		t.Fatalf("can't read data from %s (%v): %s", address, data, err)
	}

	err = conn.Close()
	if err != nil {
		t.Fatalf("can't close connection to %s (%v): %s", address, data, err)
	}
}
