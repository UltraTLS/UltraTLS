package protocol

import (
	"io"
	"net"
	"sync"

	v2mux "github.com/v2fly/v2ray-core/v5/common/mux"
)

// Session is a multiplexed session interface that supports opening and accepting streams.
type Session interface {
	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
	Close() error
}

// v2Session implements the Session interface using v2ray's mux.cool.
type v2Session struct {
	conn     net.Conn
	manager  *v2mux.SessionManager
	streamCh chan *v2mux.Session
	lock     sync.Mutex
	closed   bool
}

// NewSession creates a new multiplexed session.
// The parameter isClient can be used to differentiate behavior between client and server.
func NewSession(conn io.ReadWriteCloser, isClient bool) Session {
	s := &v2Session{
		conn:     conn.(net.Conn),
		manager:  v2mux.NewSessionManager(),
		streamCh: make(chan *v2mux.Session, 16), // buffered channel for incoming streams
	}
	// In a production implementation, you would launch a goroutine here to continuously
	// read from the underlying connection and decode incoming stream requests,
	// then push them onto s.streamCh.
	// For now, this is left as a placeholder.
	return s
}

// OpenStream allocates a new outbound stream.
func (s *v2Session) OpenStream() (Stream, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return nil, io.ErrClosedPipe
	}
	v2s := s.manager.Allocate()
	if v2s == nil {
		return nil, io.ErrClosedPipe
	}
	return newV2Stream(s.conn, v2s), nil
}

// AcceptStream blocks and accepts an inbound stream.
// In a production environment, this should be populated via a background goroutine
// that reads incoming frames and pushes new sessions onto s.streamCh.
func (s *v2Session) AcceptStream() (Stream, error) {
	// Here we block until a new stream is available.
	v2s, ok := <-s.streamCh
	if !ok {
		return nil, io.ErrClosedPipe
	}
	return newV2Stream(s.conn, v2s), nil
}

// Close terminates the session and underlying connection.
func (s *v2Session) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.closed = true
	close(s.streamCh)
	return s.conn.Close()
}
