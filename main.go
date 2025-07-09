package main

import (
	"flag"
	"log"
	"github.com/UltraTLS/UltraTLS/api"
	"github.com/UltraTLS/UltraTLS/client"
	"github.com/UltraTLS/UltraTLS/server"
	"time"
)

func main() {
	mode := flag.String("mode", "server", "server or client or udpserver or udpclient")
	addr := flag.String("addr", ":8080", "listen or connect address")
	flag.Parse()

	switch *mode {
	case "server":
		cfg := &server.ServerConfig{
			ListenAddr: *addr,
			Handler:    &api.DefaultHandler{},
		}
		log.Fatal(server.StartServer(cfg))
	case "client":
		cfg := &client.ClientConfig{
			RemoteAddr: *addr,
			Timeout:    5 * time.Second,
			Handler:    &api.DefaultHandler{},
		}
		log.Fatal(client.StartClient(cfg))
	case "udpserver":
		cfg := &server.UDPServerConfig{
			ListenAddr: *addr,
			Handler:    &api.DefaultUDPHandler{},
		}
		log.Fatal(server.StartUDPServer(cfg))
	case "udpclient":
		cfg := &client.UDPClientConfig{
			RemoteAddr: *addr,
			Timeout:    5 * time.Second,
			Handler:    &api.DefaultUDPHandler{},
		}
		log.Fatal(client.StartUDPClient(cfg))
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}
