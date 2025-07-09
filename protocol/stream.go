package protocol

import (
	"io"
	"net"
	"sync"

	v2buf "github.com/v2fly/v2ray-core/v5/common/buf"
	v2mux "github.com/v2fly/v2ray-core/v5/common/mux"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

// Stream 是一个多路复用流，用 v2ray mux.cool 实现。
type Stream interface {
	io.ReadWriteCloser
	StreamID() uint32
}

type v2Stream struct {
	conn    net.Conn
	session *v2mux.Session
	closed  bool
	rLock   sync.Mutex
	wLock   sync.Mutex
}

func newV2Stream(conn net.Conn, session *v2mux.Session) *v2Stream {
	return &v2Stream{
		conn:    conn,
		session: session,
	}
}

// Read 从流中读取数据。
func (s *v2Stream) Read(p []byte) (int, error) {
	s.rLock.Lock()
	defer s.rLock.Unlock()
	if s.closed {
		return 0, io.EOF
	}
	// Directly use v2buf.NewReader as NewBufferedReader is not available.
	reader := v2buf.NewReader(s.conn)
	// s.session.NewReader takes a *v2buf.BufferedReader; NewReader already returns one.
	bufReader := s.session.NewReader(reader)
	mb, err := bufReader.ReadMultiBuffer()
	if err != nil {
		return 0, err
	}
	if len(mb) == 0 {
		return 0, io.EOF
	}
	defer v2buf.ReleaseMulti(mb)
	n := copy(p, mb[0].Bytes())
	return n, nil
}

// Write 向流中写入数据。
func (s *v2Stream) Write(p []byte) (int, error) {
	s.wLock.Lock()
	defer s.wLock.Unlock()
	if s.closed {
		return 0, io.ErrClosedPipe
	}
	b := v2buf.New()
	defer b.Release()
	if _, err := b.Write(p); err != nil {
		return 0, err
	}
	mb := v2buf.MultiBuffer{b}
	// Use a dummy TCP destination since v2mux.TCPDestination is unexported.
	// v2net.TCPDestination creates a net.Destination for TCP.
	writer := v2mux.NewWriter(s.session.ID, v2net.TCPDestination("0.0.0.0", 0), v2buf.NewWriter(s.conn), 0)
	if err := writer.WriteMultiBuffer(mb); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close 关闭该流。
func (s *v2Stream) Close() error {
	s.closed = true
	return s.session.Close()
}

// StreamID 返回流 ID。
func (s *v2Stream) StreamID() uint32 {
	return uint32(s.session.ID)
}
