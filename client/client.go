package client

import (
	"io"
	"log"
	"myproxy/api"
	"myproxy/internal/netutil"
	"myproxy/protocol"
	"time"
)

// ClientConfig 配置
type ClientConfig struct {
	RemoteAddr string
	Timeout    time.Duration
	Handler    api.Handler
}

// StartClient 启动TCP客户端
func StartClient(cfg *ClientConfig) error {
	conn, err := netutil.DialTCP(cfg.RemoteAddr, cfg.Timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	session := protocol.NewSession(conn, true)
	if session == nil {
		return io.ErrUnexpectedEOF
	}
	stream, err := session.OpenStream()
	if err != nil {
		return err
	}
	defer stream.Close()
	if cfg.Handler != nil {
		cfg.Handler.OnConnect(conn)
	}
	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("client: read error: %v", err)
			}
			break
		}
		if cfg.Handler != nil {
			cfg.Handler.OnData(conn, buf[:n])
		}
	}
	if cfg.Handler != nil {
		cfg.Handler.OnClose(conn)
	}
	return nil
}