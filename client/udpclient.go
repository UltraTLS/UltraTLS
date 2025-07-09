package client

import (
	"log"
	"github.com/UltraTLS/UltraTLS/api"
	"github.com/UltraTLS/UltraTLS/internal/netutil"
	"net"
	"time"
)

// UDPClientConfig 配置
type UDPClientConfig struct {
	RemoteAddr string
	Timeout    time.Duration
	Handler    api.UDPHandler
}

// StartUDPClient 启动UDP客户端
func StartUDPClient(cfg *UDPClientConfig) error {
	conn, err := netutil.DialUDP(cfg.RemoteAddr, cfg.Timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	buf := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("udpclient: read error: %v", err)
			break
		}
		if cfg.Handler != nil {
			resp, err := cfg.Handler.OnPacket(addr, buf[:n])
			if err == nil && len(resp) > 0 {
				conn.Write(resp)
			}
		}
	}
	return nil
}
