package protocol

import (
	"io"
	"net"

	v2buf "github.com/v2fly/v2ray-core/v5/common/buf"
	v2mux "github.com/v2fly/v2ray-core/v5/common/mux"
)

// Stream 表示多路复用的流。
type Stream interface {
	io.ReadWriteCloser
	StreamID() uint32
}

type v2Stream struct {
	session *v2mux.Session
	conn    net.Conn
}

// newV2Stream 创建一个 v2mux.Session 的高层包装流。
func newV2Stream(session *v2mux.Session, conn net.Conn) *v2Stream {
	return &v2Stream{
		session: session,
		conn:    conn,
	}
}

func (s *v2Stream) Read(p []byte) (int, error) {
	// 使用 v2ray 的 buf.Reader 来读取数据
	if s.session == nil {
		return 0, io.ErrClosedPipe
	}
	// v2ray session.NewReader 返回 buf.Reader
	reader := s.session.NewReader(v2buf.NewBufferedReader(v2buf.NewReader(s.conn)))
	mb, err := reader.ReadMultiBuffer()
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

func (s *v2Stream) Write(p []byte) (int, error) {
	// 构造 v2ray 的 MultiBuffer 并写入
	if s.session == nil {
		return 0, io.ErrClosedPipe
	}
	// 真实生产环境应考虑分片、流控
	b := v2buf.New()
	defer b.Release()
	_, err := b.Write(p)
	if err != nil {
		return 0, err
	}
	mb := v2buf.MultiBuffer{b}
	writer := v2mux.NewWriter(s.session.ID, s.session.Target, v2buf.NewWriter(s.conn), s.session.TransferType)
	err = writer.WriteMultiBuffer(mb)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *v2Stream) Close() error {
	if s.session != nil {
		return s.session.Close()
	}
	return nil
}

func (s *v2Stream) StreamID() uint32 {
	if s.session == nil {
		return 0
	}
	return uint32(s.session.ID)
}
