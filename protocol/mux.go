package protocol

import (
	"io"
	"net"
	"sync"

	v2mux "github.com/v2fly/v2ray-core/v5/common/mux"
)

// Session 多路复用会话接口，默认用 v2ray mux.cool。
type Session interface {
	OpenStream() (Stream, error)
	Close() error
}

// v2Session wraps a v2ray mux.SessionManager for client-side stream multiplexing.
type v2Session struct {
	conn    net.Conn
	manager *v2mux.SessionManager
	lock    sync.Mutex
	closed  bool
}

// NewSession 创建一个 v2mux.SessionManager 驱动的多路复用会话。
// conn 必须为 net.Conn，isClient 目前仅为兼容参数。
func NewSession(conn io.ReadWriteCloser, isClient bool) Session {
	return &v2Session{
		conn:    conn.(net.Conn),
		manager: v2mux.NewSessionManager(),
	}
}

// OpenStream 分配新流。
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

// Close 关闭底层连接和会话池。
func (s *v2Session) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.closed = true
	return s.conn.Close()
}
