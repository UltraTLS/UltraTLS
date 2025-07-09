package protocol

import (
	"io"
	"github.com/v2fly/v2ray-core/v5/common/mux"
)

type Stream interface {
	io.ReadWriteCloser
	StreamID() uint32
}

type v2Stream struct {
	session *mux.Session
}

func (s *v2Stream) Read(p []byte) (int, error) {
	return s.session.input.Read(p)
}

func (s *v2Stream) Write(p []byte) (int, error) {
	return s.session.output.Write(p)
}

func (s *v2Stream) Close() error {
	return s.session.Close()
}

func (s *v2Stream) StreamID() uint32 {
	return uint32(s.session.ID)
}