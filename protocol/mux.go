package protocol

import (
	"io"
	"net"

	v2mux "github.com/v2fly/v2ray-core/v5/common/mux"
)

// Session 是多路复用会话接口
type Session interface {
	OpenStream() (Stream, error)
	AcceptStream() (Stream, error)
	Close() error
}

type v2Session struct {
	manager *v2mux.SessionManager
	conn    net.Conn
}

func NewSession(conn io.ReadWriteCloser, isClient bool) Session {
	manager := v2mux.NewSessionManager()
	// 这里只做演示，实际你需要根据 isClient/server 选择不同逻辑
	return &v2Session{
		manager: manager,
		conn:    conn.(net.Conn),
	}
}

func (s *v2Session) OpenStream() (Stream, error) {
	// 分配一个 Session
	vs := s.manager.Allocate()
	if vs == nil {
		return nil, io.ErrClosedPipe
	}
	// 你需要自己实现 Stream，参考 stream.go
	return &v2Stream{session: vs}, nil
}

func (s *v2Session) AcceptStream() (Stream, error) {
	// 这里需要和 v2mux 的底层事件绑定
	// 实战中你可能要用 channel 或事件监听
	return nil, io.EOF
}

func (s *v2Session) Close() error {
	return s.conn.Close()
}