package client

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/UltraTLS/UltraTLS/protocol"
	v2net "github.com/v2fly/v2ray-core/v5/common/net"
)

// ClientConfig 配置结构
type ClientConfig struct {
	ServerAddress string               // 远程服务器主连接地址
	LocalListen   string               // 本地监听地址
	Handler       func(stream protocol.Stream)
}

// StartClient 启动客户端
func StartClient(cfg *ClientConfig) error {
	inbound := protocol.NewTCPInbound(cfg.LocalListen)
	if err := inbound.Listen(); err != nil {
		return err
	}
	defer inbound.Close()
	log.Printf("client: listening on %s", cfg.LocalListen)
	for {
		conn, err := inbound.Accept()
		if err != nil {
			log.Printf("client accept error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		go func(clientConn net.Conn) {
			defer clientConn.Close()
			// 与服务器建立主连接
			outbound := protocol.NewTCPOutbound(10 * time.Second)
			serverConn, err := outbound.Connect("tcp", cfg.ServerAddress)
			if err != nil {
				log.Printf("client connect to server error: %v", err)
				return
			}
			defer outbound.Close()
			defer serverConn.Close()

			// 建立mux session
			session := protocol.NewSession(serverConn)
			defer session.Close()

			// 为每个本地连接开一个mux子流
			dest := v2net.TCPDestination(v2net.IPAddress(net.ParseIP("127.0.0.1")), 0)
			stream, err := session.OpenStream(dest)
			if err != nil {
				log.Printf("client open mux stream error: %v", err)
				return
			}
			defer stream.Close()

			// 业务处理
			if cfg.Handler != nil {
				cfg.Handler(stream)
			} else {
				// 默认双向转发
				go func() { _, _ = io.Copy(stream, clientConn) }()
				_, _ = io.Copy(clientConn, stream)
			}
		}(conn)
	}
}
