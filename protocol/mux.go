package protocol

import (
	"io"
	"net"

	v2mux "github.com/v2fly/v2ray-core/v5/common/mux"
)

// Session 表示多路复用会话。
type Session interface {
	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
	Close() error
}

type v2Session struct {
	manager *v2mux.SessionManager
	conn    net.Conn
}

// NewSession 返回一个基于 v2ray mux.cool 的会话。
// isClient 为 true 时表示客户端，false 为服务端。
func NewSession(conn io.ReadWriteCloser, isClient bool) Session {
	return &v2Session{
		manager: v2mux.NewSessionManager(),
		conn:    conn.(net.Conn),
	}
}

func (s *v2Session) OpenStream() (Stream, error) {
	sess := s.manager.Allocate()
	if sess == nil {
		return nil, io.ErrClosedPipe
	}
	// 这里包装 Session，供 Stream 用
	return newV2Stream(sess, s.conn), nil
}

func (s *v2Session) AcceptStream() (Stream, error) {
	// v2ray mux.cool 没有直接暴露 AcceptStream，实际项目中可以用更复杂的事件驱动或 channel 实现
	return nil, io.EOF
}

func (s *v2Session) Close() error {
	return s.conn.Close()
}
