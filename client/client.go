package client

import (
	"github.com/UltraTLS/UltraTLS/protocol"
	"log"
	"time"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

// ClientConfig
type ClientConfig struct {
	ServerAddress string            // 远程服务器主连接地址
	LocalListen   string            // 本地监听
	Handler       func(stream protocol.Stream)
}

func StartClient(cfg *ClientConfig) error {
	inbound := protocol.NewTCPInbound(cfg.LocalListen)
	if err := inbound.Listen(); err != nil {
		return err
	}
	defer inbound.Close()
	for {
		conn, err := inbound.Accept()
		if err != nil {
			log.Printf("client accept error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		go func(clientConn net.Conn) {
			// 建立主连接到服务端
			outbound := protocol.NewTCPOutbound(time.Second * 10)
			serverConn, err := outbound.Connect("tcp", cfg.ServerAddress)
			if err != nil {
				log.Printf("client connect to server error: %v", err)
				clientConn.Close()
				return
			}
			defer outbound.Close()
			// 建立mux session
			session := protocol.NewSession(serverConn)
			defer session.Close()
			// 为每个本地连接开一个mux子流
			stream, err := session.OpenStream(net.TCPDestination("127.0.0.1", 0)) // 可自定义目标
			if err != nil {
				log.Printf("client open mux stream error: %v", err)
				clientConn.Close()
				return
			}
			defer stream.Close()
			cfg.Handler(stream)
		}(conn)
	}
}
