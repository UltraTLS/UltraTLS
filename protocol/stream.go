package protocol

import (
	"io"
	"net"
	"sync"

	v2buf "github.com/v2fly/v2ray-core/v5/common/buf"
	v2mux "github.com/v2fly/v2ray-core/v5/common/mux"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

// Stream is a multiplexed stream implemented using v2ray mux.cool.
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

// Read reads data from the stream.
func (s *v2Stream) Read(p []byte) (int, error) {
	s.rLock.Lock()
	defer s.rLock.Unlock()
	if s.closed {
		return 0, io.EOF
	}
	// Create a new buffered reader from the connection.
	reader := v2buf.NewReader(s.conn)
	// v2mux.Session.NewReader expects a *v2buf.BufferedReader.
	br, ok := reader.(*v2buf.BufferedReader)
	if !ok {
		// Should not happen as NewReader returns *BufferedReader.
		return 0, io.ErrUnexpectedEOF
	}
	bufReader := s.session.NewReader(br)
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

// Write writes data to the stream.
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
	// For v2mux.NewWriter, we need to provide a Destination.
	// Since we don't have a meaningful destination here, we use a dummy TCP destination.
	// Convert "0.0.0.0" to an Address using v2net.IPAddress.
	dummyDest := v2net.TCPDestination(v2net.IPAddress(net.ParseIP("0.0.0.0")), 0)
	writer := v2mux.NewWriter(s.session.ID, dummyDest, v2buf.NewWriter(s.conn), 0)
	if err := writer.WriteMultiBuffer(mb); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close closes the stream.
func (s *v2Stream) Close() error {
	s.closed = true
	return s.session.Close()
}

// StreamID returns the ID of the stream.
func (s *v2Stream) StreamID() uint32 {
	return uint32(s.session.ID)
}
