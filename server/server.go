package server

import (
	"io"
	"log"
	"myproxy/api"
	"myproxy/internal/netutil"
	"myproxy/protocol"
	"net"
)

// ServerConfig 配置
type ServerConfig struct {
	ListenAddr string
	Handler    api.Handler
}

// StartServer 启动TCP服务端
func StartServer(cfg *ServerConfig) error {
	ln, err := netutil.ListenTCP(cfg.ListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("server: listening on %s", cfg.ListenAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("server: accept error: %v", err)
			continue
		}
		go handleConn(conn, cfg.Handler)
	}
}

func handleConn(conn net.Conn, handler api.Handler) {
	defer conn.Close()
	session := protocol.NewSession(conn, false)
	if session == nil {
		log.Printf("server: failed to create session")
		return
	}
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			log.Printf("server: accept stream error: %v", err)
			return
		}
		go handleStream(stream, handler, conn)
	}
}

func handleStream(stream protocol.Stream, handler api.Handler, conn net.Conn) {
	defer stream.Close()
	if handler != nil {
		handler.OnConnect(conn)
	}
	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("server: stream read error: %v", err)
			}
			break
		}
		if handler != nil {
			handler.OnData(conn, buf[:n])
		}
	}
	if handler != nil {
		handler.OnClose(conn)
	}
}