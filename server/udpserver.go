package server

import (
	"log"
	"myproxy/api"
	"myproxy/internal/netutil"
	"net"
)

// UDPServerConfig 配置
type UDPServerConfig struct {
	ListenAddr string
	Handler    api.UDPHandler
}

// StartUDPServer 启动UDP服务端
func StartUDPServer(cfg *UDPServerConfig) error {
	conn, err := netutil.ListenUDP(cfg.ListenAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("udpserver: listening on %s", cfg.ListenAddr)
	buf := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("udpserver: read error: %v", err)
			continue
		}
		if cfg.Handler != nil {
			resp, err := cfg.Handler.OnPacket(addr, buf[:n])
			if err == nil && len(resp) > 0 {
				conn.WriteTo(resp, addr)
			}
		}
	}
}