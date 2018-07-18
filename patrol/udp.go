package main

import (
	"fmt"
	"log"
	"net"
	"sabey.co/patrol"
)

func udpListen() string {
	return fmt.Sprintf("127.0.0.1:%d", patrol.LISTEN_UDP_PORT_DEFAULT)
}
func UDP() {
	defer log.Println("./patrol.UDP(): Stopped!!!")
	listen := p.GetConfig().UDP.Listen
	log.Printf("./patrol.UDP(): Listen: \"%s\"\n", listen)
	conn, err := net.ListenPacket("udp", listen)
	if err != nil {
		log.Printf("./patrol.UDP(): Failed to Listen: \"%s\"\n", err)
		return
	}
	defer conn.Close()
	// accept connections
	for {
		if err := p.HandleUDPConnection(conn); err != nil {
			// call shutdown
			shutdown()
			return
		}
	}
}
